package service

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"your-project/model"
	"your-project/repository"
)

type RAGService struct {
	questionRepo *repository.QuestionRepository
	vectorStore  VectorStore
}

type VectorStore interface {
	Search(query string, limit int) ([]SimilarityResult, error)
	Index(questions []*model.Question) error
}

type SimilarityResult struct {
	Question *model.Question
	Score    float64
}

func NewRAGService() *RAGService {
	return &RAGService{
		questionRepo: repository.NewQuestionRepository(),
		vectorStore:  NewSimpleVectorStore(),
	}
}

func (s *RAGService) SearchSimilarQuestions(query string, position, difficulty string, limit int) ([]*model.Question, error) {
	allQuestions, err := s.questionRepo.GetQuestionsByPositionAndDifficulty(position, difficulty)
	if err != nil {
		return nil, fmt.Errorf("failed to get questions: %w", err)
	}

	if err := s.vectorStore.Index(allQuestions); err != nil {
		return nil, fmt.Errorf("failed to index questions: %w", err)
	}

	similarResults, err := s.vectorStore.Search(query, limit*2)
	if err != nil {
		return nil, fmt.Errorf("failed to search similar questions: %w", err)
	}

	var filteredQuestions []*model.Question
	for _, result := range similarResults {
		if result.Score > 0.3 {
			filteredQuestions = append(filteredQuestions, result.Question)
		}
		if len(filteredQuestions) >= limit {
			break
		}
	}

	return filteredQuestions, nil
}

func (s *RAGService) GenerateQuestionBasedOnContext(context string, position, difficulty string) (*model.Question, error) {
	similarQuestions, err := s.SearchSimilarQuestions(context, position, difficulty, 5)
	if err != nil {
		return nil, fmt.Errorf("failed to search similar questions: %w", err)
	}

	if len(similarQuestions) == 0 {
		return s.createDefaultQuestion(position, difficulty)
	}

	bestQuestion := similarQuestions[0]
	return s.adaptQuestion(bestQuestion, context), nil
}

func (s *RAGService) createDefaultQuestion(position, difficulty string) (*model.Question, error) {
	question := &model.Question{
		Title:      fmt.Sprintf("%s - %s Level Question", position, difficulty),
		Content:    fmt.Sprintf("请描述你在%s方面的经验，以及你如何处理相关的技术挑战。", position),
		Position:   position,
		Difficulty: difficulty,
		Category:   "general",
	}
	question.SetTags([]string{position, difficulty, "experience"})

	return question, nil
}

func (s *RAGService) adaptQuestion(original *model.Question, context string) *model.Question {
	adapted := *original
	tags := adapted.GetTags()
	if strings.Contains(context, "项目") || strings.Contains(context, "project") {
		adapted.Content = fmt.Sprintf("结合你之前的项目经验，%s", original.Content)
		tags = append(tags, "project-based")
	}

	if strings.Contains(context, "团队") || strings.Contains(context, "team") {
		adapted.Content = fmt.Sprintf("在团队协作的场景下，%s", original.Content)
		tags = append(tags, "team-collaboration")
	}
	adapted.SetTags(tags)

	return &adapted
}

type SimpleVectorStore struct {
	questions []*model.Question
}

func NewSimpleVectorStore() *SimpleVectorStore {
	return &SimpleVectorStore{
		questions: make([]*model.Question, 0),
	}
}

func (s *SimpleVectorStore) Index(questions []*model.Question) error {
	s.questions = questions
	return nil
}

func (s *SimpleVectorStore) Search(query string, limit int) ([]SimilarityResult, error) {
	var results []SimilarityResult

	for _, question := range s.questions {
		similarity := s.calculateSimilarity(query, question)
		if similarity > 0 {
			results = append(results, SimilarityResult{
				Question: question,
				Score:    similarity,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

func (s *SimpleVectorStore) calculateSimilarity(query string, question *model.Question) float64 {
	queryWords := s.tokenize(query)
	questionWords := s.tokenize(question.Content + " " + question.Title)

	if len(queryWords) == 0 || len(questionWords) == 0 {
		return 0
	}

	commonWords := 0
	querySet := make(map[string]bool)
	for _, word := range queryWords {
		querySet[word] = true
	}

	for _, word := range questionWords {
		if querySet[word] {
			commonWords++
		}
	}

	return float64(commonWords) / math.Sqrt(float64(len(queryWords)*len(questionWords)))
}

func (s *SimpleVectorStore) tokenize(text string) []string {
	words := strings.Fields(strings.ToLower(text))
	var filtered []string

	for _, word := range words {
		if len(word) > 2 && !s.isStopWord(word) {
			filtered = append(filtered, word)
		}
	}

	return filtered
}

func (s *SimpleVectorStore) isStopWord(word string) bool {
	stopWords := map[string]bool{
		"the": true, "and": true, "or": true, "but": true, "in": true,
		"on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "is": true, "are": true, "was": true,
	}
	return stopWords[word]
}
