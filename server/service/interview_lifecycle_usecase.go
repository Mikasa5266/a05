package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"your-project/model"
)

type InterviewLifecycleUseCase interface {
	SubmitAnswer(userID, interviewID, questionID uint, answer, audioData, audioMime, questionTitle, questionContent string) (*model.AnswerResult, error)
	EndInterview(userID, interviewID uint) (*model.Interview, error)
}

type interviewLifecycleUseCase struct {
	service *InterviewService
}

var _ InterviewLifecycleUseCase = (*interviewLifecycleUseCase)(nil)

func NewInterviewLifecycleUseCase(service *InterviewService) InterviewLifecycleUseCase {
	return &interviewLifecycleUseCase{service: service}
}

func (u *interviewLifecycleUseCase) SubmitAnswer(userID, interviewID, questionID uint, answer, audioData, audioMime, questionTitle, questionContent string) (*model.AnswerResult, error) {
	_ = audioMime

	ctx := context.Background()

	interview, err := u.service.GetInterviewByID(userID, interviewID)
	if err != nil {
		return nil, err
	}

	if normalizeStatusValue(interview.Status) != interviewStatusInProgress {
		return nil, fmt.Errorf("interview is not in progress")
	}

	baseQuestion, err := u.service.questionRepo.GetByID(questionID)
	if err != nil {
		return nil, fmt.Errorf("question not found")
	}
	interview.AskedQuestionIDs = mergeAskedQuestionIDs(interview.AskedQuestionIDs, baseQuestion.ID)

	evalQuestion := baseQuestion
	if strings.TrimSpace(questionContent) != "" {
		tempQ := *baseQuestion
		tempQ.Content = strings.TrimSpace(questionContent)
		if strings.TrimSpace(questionTitle) != "" {
			tempQ.Title = strings.TrimSpace(questionTitle)
		}
		evalQuestion = &tempQ
	}

	var finalAnswer string
	if audioData != "" {
		transcribedText, err := u.service.aiService.TranscribeAudio(strings.TrimSpace(audioData))
		if err != nil {
			return nil, fmt.Errorf("failed to transcribe audio: %w", err)
		}
		finalAnswer = transcribedText
		interview.ASRCallCount++
	} else {
		finalAnswer = answer
	}

	evaluation, err := u.service.aiService.EvaluateAnswer(ctx, evalQuestion, finalAnswer)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate answer: %w", err)
	}
	shouldFollowUpHint, followUpContext := parseEvaluationFollowUpHint(evaluation.Feedback)

	result := &model.AnswerResult{
		InterviewID: interviewID,
		QuestionID:  baseQuestion.ID,
		Answer:      finalAnswer,
		Score:       evaluation.Score,
		Feedback:    evaluation.Feedback,
		CreatedAt:   time.Now(),
	}

	if err := u.service.interviewRepo.SaveAnswer(result); err != nil {
		return nil, fmt.Errorf("failed to save answer: %w", err)
	}

	answers, _ := u.service.interviewRepo.GetAnswersByInterviewID(interviewID)

	// Follow-up questions are generated in real time and kept ephemeral (no DB write).
	askedQuestionText := strings.TrimSpace(evalQuestion.Content)
	if askedQuestionText == "" {
		askedQuestionText = strings.TrimSpace(evalQuestion.Title)
	}
	shouldFollowUp, nextQuestion, err := u.service.decideNextQuestion(ctx, interview, baseQuestion, askedQuestionText, finalAnswer, evaluation.Score, shouldFollowUpHint, followUpContext)
	if err != nil {
		fmt.Printf("Dynamic question generation failed: %v\n", err)
	}

	if shouldFollowUp && nextQuestion != nil {
		nextQuestion.Source = "follow_up"
		nextQuestion.RAGEligible = false
		nextQuestion.ID = baseQuestion.ID
		nextQuestion.Position = baseQuestion.Position
		nextQuestion.Difficulty = baseQuestion.Difficulty
		nextQuestion.Category = baseQuestion.Category
		result.NextQuestion = nextQuestion
		interview.FollowUpCount++
	} else {
		interview.CurrentIndex++
		interview.FollowUpCount = 0
		interview.TopicIndex++

		allQuestions, _ := u.service.interviewRepo.GetInterviewQuestions(interviewID)
		excludeIDs := decodeAskedQuestionIDs(interview.AskedQuestionIDs)
		if interview.CurrentIndex < len(allQuestions) {
			nextQ, _ := u.service.questionRepo.GetByID(allQuestions[interview.CurrentIndex].QuestionID)
			if nextQ != nil {
				u.service.normalizeOpeningQuestion(ctx, nextQ)
				interview.CurrentTopic = nextQ.Category
				interview.AskedQuestionIDs = mergeAskedQuestionIDs(interview.AskedQuestionIDs, nextQ.ID)
				result.NextQuestion = nextQ
			}
		}

		if result.NextQuestion == nil {
			fallbackQ, pickErr := u.service.questionRepo.GetRandomQuestionForInterview(interview.Position, interview.Difficulty, excludeIDs)
			if pickErr == nil && fallbackQ != nil {
				u.service.normalizeOpeningQuestion(ctx, fallbackQ)
				interview.CurrentTopic = fallbackQ.Category
				interview.AskedQuestionIDs = mergeAskedQuestionIDs(interview.AskedQuestionIDs, fallbackQ.ID)
				result.NextQuestion = fallbackQ

				_ = u.service.interviewRepo.CreateInterviewQuestion(&model.InterviewQuestion{
					InterviewID: interviewID,
					QuestionID:  fallbackQ.ID,
					OrderIndex:  interview.CurrentIndex,
					IsAnswered:  false,
				})
			}
		}
	}

	allQuestions, _ := u.service.interviewRepo.GetInterviewQuestions(interviewID)
	if interview.TotalQuestionTarget > 0 && len(answers) >= interview.TotalQuestionTarget {
		completedStatus, transitionErr := transitionInterviewStatus(interview.Status, interviewStatusCompleted)
		if transitionErr != nil {
			return nil, transitionErr
		}
		interview.Status = completedStatus
		t := time.Now()
		interview.EndTime = &t
		result.InterviewCompleted = true
	} else if interview.CurrentIndex >= len(allQuestions) && result.NextQuestion == nil {
		completedStatus, transitionErr := transitionInterviewStatus(interview.Status, interviewStatusCompleted)
		if transitionErr != nil {
			return nil, transitionErr
		}
		interview.Status = completedStatus
		t := time.Now()
		interview.EndTime = &t
		result.InterviewCompleted = true
	}

	if err := u.service.interviewRepo.Update(interview); err != nil {
		return nil, fmt.Errorf("failed to update interview: %w", err)
	}

	return result, nil
}

func (u *interviewLifecycleUseCase) EndInterview(userID, interviewID uint) (*model.Interview, error) {
	interview, err := u.service.GetInterviewByID(userID, interviewID)
	if err != nil {
		return nil, err
	}

	if normalizeStatusValue(interview.Status) == interviewStatusCompleted {
		return interview, nil
	}

	completedStatus, transitionErr := transitionInterviewStatus(interview.Status, interviewStatusCompleted)
	if transitionErr != nil {
		return nil, transitionErr
	}
	interview.Status = completedStatus
	t := time.Now()
	interview.EndTime = &t

	if err := u.service.interviewRepo.Update(interview); err != nil {
		return nil, fmt.Errorf("failed to update interview: %w", err)
	}

	if interview.InterviewMode == "human" {
		inv, invErr := u.service.interviewRepo.GetInvitationByInterviewID(interviewID)
		if invErr == nil && inv != nil {
			nextStatus, inviteTransitionErr := transitionInvitationStatus(inv.Status, invitationStatusCompleted)
			if inviteTransitionErr == nil {
				inv.Status = nextStatus
				_ = u.service.interviewRepo.UpdateInvitation(inv)
			}
		}
	}

	return interview, nil
}
