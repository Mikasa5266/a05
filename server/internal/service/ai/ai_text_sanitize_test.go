package ai

import (
	"strings"
	"testing"

	"your-project/internal/model"
)

func TestSanitizeGeneratedText_RemovesReplacementChars(t *testing.T) {
	raw := "\ufffd\ufffd请结合实际项目说明你的实现思路"
	got := sanitizeGeneratedText(raw)
	if strings.ContainsRune(got, '\ufffd') {
		t.Fatalf("expected replacement rune removed, got %q", got)
	}
	if got != "请结合实际项目说明你的实现思路" {
		t.Fatalf("unexpected sanitized text: %q", got)
	}
}

func TestNormalizeToSelfContainedOpening_UsesChineseTemplate(t *testing.T) {
	svc := &AIService{}
	q := &model.Question{Category: "并发控制"}

	svc.NormalizeToSelfContainedOpening(q)

	if strings.Contains(strings.ToLower(q.Title), "core principles") {
		t.Fatalf("expected Chinese title template, got %q", q.Title)
	}
	if !strings.Contains(q.Title, "核心原理与实践应用") {
		t.Fatalf("expected Chinese normalized title, got %q", q.Title)
	}
	if strings.Contains(strings.ToLower(q.Content), "please explain") {
		t.Fatalf("expected Chinese content template, got %q", q.Content)
	}
}

func TestParseFollowUpModelOutput_CleansReplacementChars(t *testing.T) {
	question, noFollowUp, _ := parseFollowUpModelOutput("\ufffd\ufffd请继续说明锁粒度如何影响吞吐")
	if noFollowUp {
		t.Fatalf("expected follow-up question")
	}
	if strings.ContainsRune(question, '\ufffd') {
		t.Fatalf("expected replacement rune removed, got %q", question)
	}
}

