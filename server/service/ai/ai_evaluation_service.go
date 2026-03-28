package ai

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"your-project/model"
)

func (s *AIService) EvaluateAnswer(ctx context.Context, question *model.Question, answer string) (*EvaluationResult, error) {
	llmFunc := func(p string) (string, error) {
		return s.chat(ctx, p, "evaluation", jsonObjectResponseFormat())
	}

	reviewResult, err := EvaluateCandidateAnswer(question.Content, question.ExpectedAnswer, answer, llmFunc)
	if err != nil {
		log.Printf("AI review failed, fallback to local heuristic: %v", err)
		return s.evaluateAnswerLocal(question, answer), nil
	}

	evaluationText := s.EnsureChineseOutput(ctx, reviewResult.Comment, "Answer received. Quality needs improvement.")
	suggestionText := s.EnsureChineseOutput(ctx, reviewResult.Suggestion, "Please add core principles and implementation details.")
	suggestionItems := splitSuggestionText(suggestionText)

	dims := reviewResult.Dimensions
	if dims == nil {
		dims = estimateDimensions(reviewResult.Score)
	}

	richFeedback := map[string]interface{}{
		"evaluation":           evaluationText,
		"suggestions":          suggestionItems,
		"dimensions":           dims,
		"highlights":           reviewResult.Highlights,
		"gaps":                 reviewResult.Gaps,
		"model_answer_outline": reviewResult.ModelAnswerOutline,
		"follow_up":            reviewResult.FollowUp,
	}
	feedbackJSON, _ := json.Marshal(richFeedback)
	return &EvaluationResult{Score: reviewResult.Score, Feedback: string(feedbackJSON)}, nil
}

func splitSuggestionText(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return []string{"Add core principles and practical details."}
	}
	parts := strings.FieldsFunc(text, func(r rune) bool {
		return r == ';' || r == '；'
	})
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		p = strings.TrimLeft(p, "0123456789.、) ")
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	if len(result) == 0 {
		return []string{text}
	}
	return result
}

func estimateDimensions(totalScore int) *ReviewDimensions {
	base := totalScore
	return &ReviewDimensions{
		TechnicalDepth: clampScore(base - 5 + (base%7 - 3)),
		Expression:     clampScore(base + 3 + (base%5 - 2)),
		Logic:          clampScore(base + (base%6 - 3)),
		Completeness:   clampScore(base - 3 + (base%4 - 2)),
	}
}

func (s *AIService) evaluateAnswerLocal(question *model.Question, answer string) *EvaluationResult {
	content := strings.TrimSpace(answer)
	questionContent := ""
	questionTitle := ""
	expectedAnswer := ""
	if question != nil {
		questionContent = strings.TrimSpace(question.Content)
		questionTitle = strings.TrimSpace(question.Title)
		expectedAnswer = strings.TrimSpace(question.ExpectedAnswer)
	}
	promptContext := strings.TrimSpace(questionTitle + " " + questionContent)

	if IsInvalidAnswer(content) {
		return s.buildRichLocalFeedback(0, questionContent,
			"Answer is invalid or explicitly gives up.",
			[]string{"Start from key definition, then principles and examples", "Avoid empty responses like I don't know"},
			&ReviewDimensions{},
			nil,
			[]string{"No valid technical signal for evaluation"},
		)
	}

	if len([]rune(content)) < 10 {
		return s.buildRichLocalFeedback(0, questionContent,
			"Answer is too short and does not form a valid technical response.",
			[]string{"Cover key constraints first", "Provide implementation steps and data structures", "Add complexity and edge cases"},
			&ReviewDimensions{},
			nil,
			[]string{"No effective technical approach was formed"},
		)
	}

	signals := simpleAnswerSignals(promptContext, expectedAnswer, content)

	score := 18
	switch {
	case signals.runeLen >= 220:
		score += 22
	case signals.runeLen >= 140:
		score += 18
	case signals.runeLen >= 80:
		score += 14
	case signals.runeLen >= 40:
		score += 8
	case signals.runeLen >= 20:
		score += 4
	}
	if signals.hasStructure {
		score += 8
	}
	score += int(signals.keywordCoverage*45 + 0.5)

	techBonus := signals.technicalHits * 3
	if techBonus > 15 {
		techBonus = 15
	}
	score += techBonus
	if signals.keywordCoverage < 0.05 && signals.runeLen < 40 && score > 45 {
		score = 45
	}
	score = clampScore(score)

	var evaluation string
	var highlights []string
	var gaps []string
	if score >= 80 {
		evaluation = "Good coverage of key points with clear structure."
		highlights = []string{"Core points were covered", "Structured and reasonably deep"}
		gaps = []string{"Could add more low-level implementation details"}
	} else if score >= 60 {
		evaluation = "Answer is relevant but lacks depth in mechanism and evidence."
		if signals.hasStructure {
			highlights = []string{"Basic structure is present"}
		}
		gaps = []string{"Mechanism explanation is shallow", "Missing trade-offs and outcomes"}
	} else {
		evaluation = "Answer lacks relevance and completeness for a high score."
		gaps = []string{"Core points missing", "Causal chain is incomplete", "No verifiable practical evidence"}
	}

	dims := &ReviewDimensions{
		TechnicalDepth: clampScore(score - 12 + signals.technicalHits*2),
		Expression:     clampScore(score - 4),
		Logic:          clampScore(score - 5),
		Completeness:   clampScore(int(float64(score) * (0.65 + 0.5*signals.keywordCoverage))),
	}
	if signals.hasStructure {
		dims.Expression = clampScore(dims.Expression + 5)
		dims.Logic = clampScore(dims.Logic + 6)
	}
	if signals.keywordCoverage >= 0.25 {
		dims.Completeness = clampScore(dims.Completeness + 8)
	}
	alignDimensionsWithScoreSimple(dims, score)

	suggestions := make([]string, 0, 4)
	if signals.keywordCoverage < 0.2 {
		suggestions = append(suggestions, "Cover core keywords first, then details")
	}
	if signals.technicalHits < 2 {
		suggestions = append(suggestions, "Add low-level principles, complexity, and trade-offs")
	}
	if !signals.hasStructure {
		suggestions = append(suggestions, "Use a conclusion -> principle -> example -> boundary structure")
	}
	suggestions = append(suggestions, "Add measurable outcomes (latency, throughput, error rate)")
	if len(suggestions) > 3 {
		suggestions = suggestions[:3]
	}

	return s.buildRichLocalFeedback(score, questionContent, evaluation, suggestions, dims, highlights, gaps)
}

type simpleSignals struct {
	runeLen         int
	keywordCoverage float64
	technicalHits   int
	hasStructure    bool
}

func simpleAnswerSignals(promptContext, expectedAnswer, content string) simpleSignals {
	merged := strings.ToLower(strings.TrimSpace(promptContext + " " + expectedAnswer))
	answer := strings.ToLower(strings.TrimSpace(content))
	keywords := strings.Fields(merged)
	hit := 0
	seen := map[string]bool{}
	for _, k := range keywords {
		k = strings.TrimSpace(k)
		if len([]rune(k)) < 2 || seen[k] {
			continue
		}
		seen[k] = true
		if strings.Contains(answer, k) {
			hit++
		}
	}
	total := len(seen)
	coverage := 0.0
	if total > 0 {
		coverage = float64(hit) / float64(total)
	}

	techTerms := []string{"complexity", "cache", "index", "transaction", "thread", "lock", "性能", "复杂度", "索引", "并发"}
	techHits := 0
	for _, t := range techTerms {
		if strings.Contains(answer, strings.ToLower(t)) {
			techHits++
		}
	}

	hasStructure := strings.Contains(answer, "first") || strings.Contains(answer, "then") || strings.Contains(answer, "最后") || strings.Contains(answer, "首先")

	return simpleSignals{
		runeLen:         len([]rune(strings.TrimSpace(content))),
		keywordCoverage: coverage,
		technicalHits:   techHits,
		hasStructure:    hasStructure,
	}
}

func alignDimensionsWithScoreSimple(dims *ReviewDimensions, score int) {
	if dims == nil {
		return
	}
	avg := (dims.TechnicalDepth + dims.Expression + dims.Logic + dims.Completeness) / 4
	delta := score - avg
	if delta == 0 {
		return
	}
	dims.TechnicalDepth = clampScore(dims.TechnicalDepth + delta)
	dims.Expression = clampScore(dims.Expression + delta)
	dims.Logic = clampScore(dims.Logic + delta)
	dims.Completeness = clampScore(dims.Completeness + delta)
}

func (s *AIService) buildRichLocalFeedback(score int, questionContent, evaluation string, suggestions []string, dims *ReviewDimensions, highlights, gaps []string) *EvaluationResult {
	richFeedback := map[string]interface{}{
		"evaluation":           evaluation,
		"suggestions":          suggestions,
		"dimensions":           dims,
		"highlights":           highlights,
		"gaps":                 gaps,
		"model_answer_outline": "Start with definition, then mechanism, scenario and caveats.",
		"follow_up":            "Can you describe how this was used in your project with concrete details?",
	}
	feedbackJSON, _ := json.Marshal(richFeedback)
	return &EvaluationResult{Score: score, Feedback: string(feedbackJSON)}
}
