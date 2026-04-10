package service

import (
	"strings"
	"testing"

	"your-project/model"
)

func TestCleanQuestionPayloadTextRemovesFenceAndInvalidRunes(t *testing.T) {
	raw := "```json\r\n{\"stem\":\"缓存如何击穿\ufffd\"}\r\n```"
	cleaned := cleanQuestionPayloadText(raw)
	if strings.Contains(cleaned, "```") {
		t.Fatalf("expected markdown fence removed, got %q", cleaned)
	}
	if strings.ContainsRune(cleaned, '\ufffd') {
		t.Fatalf("expected replacement rune removed, got %q", cleaned)
	}
	if !strings.Contains(cleaned, "stem") {
		t.Fatalf("expected payload content preserved, got %q", cleaned)
	}
}

func TestEvaluateQuestionAnswerChoiceMode(t *testing.T) {
	question := &model.Question{
		QuestionType:   model.QuestionTypeTechnicalKnowledge,
		StandardAnswer: "B",
	}
	question.SetOptions([]model.QuestionOption{
		{Key: "A", Text: "只讲概念"},
		{Key: "B", Text: "先定位瓶颈再优化"},
	})

	result := evaluateQuestionAnswer(question, "```json\nb\n```")
	if !result.IsCorrect {
		t.Fatalf("expected correct choice answer")
	}
	if result.AnswerMode != "choice" {
		t.Fatalf("expected choice mode, got %s", result.AnswerMode)
	}
}

func TestEvaluateQuestionAnswerKeywordMode(t *testing.T) {
	question := &model.Question{
		QuestionType:   model.QuestionTypeScenario,
		StandardAnswer: "限流|降级|熔断|监控",
		KnowledgePoint: "流量治理",
	}

	result := evaluateQuestionAnswer(question, "先做限流和降级，再补监控与告警。")
	if !result.IsCorrect {
		t.Fatalf("expected keyword answer to pass, got %#v", result)
	}
	if result.AnswerMode != "keyword" {
		t.Fatalf("expected keyword mode, got %s", result.AnswerMode)
	}
}
