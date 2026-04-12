package ai

import "testing"

func TestParseGroundedEvalResponse_WithFencedJSON(t *testing.T) {
	raw := "```json\n{\n  \"score\": 88,\n  \"reasoning\": \"回答覆盖主要要点，但缺少边界条件说明\",\n  \"should_follow_up\": true,\n  \"follow_up_context\": \"请继续说明并发场景下如何保证一致性\"\n}\n```"

	var parsed groundedEvalLLMResponse
	if err := parseGroundedEvalResponse(raw, &parsed); err != nil {
		t.Fatalf("expected parser success, got error: %v", err)
	}

	if parsed.Score != 88 {
		t.Fatalf("expected score 88, got %d", parsed.Score)
	}
	if !parsed.ShouldFollowUp {
		t.Fatalf("expected should_follow_up=true")
	}
	if parsed.FollowUpContext == "" {
		t.Fatalf("expected non-empty follow_up_context")
	}
}

func TestParseGroundedEvalResponse_ClearContextWhenNoFollowUp(t *testing.T) {
	raw := "{\"score\": 93, \"reasoning\": \"回答完整\", \"should_follow_up\": false, \"follow_up_context\": \"这段应被清空\"}"

	var parsed groundedEvalLLMResponse
	if err := parseGroundedEvalResponse(raw, &parsed); err != nil {
		t.Fatalf("expected parser success, got error: %v", err)
	}

	if parsed.ShouldFollowUp {
		t.Fatalf("expected should_follow_up=false")
	}
	if parsed.FollowUpContext != "" {
		t.Fatalf("expected empty follow_up_context, got %q", parsed.FollowUpContext)
	}
}

func TestParseFollowUpModelOutput_PlainTextQuestion(t *testing.T) {
	question, noFollowUp, reason := parseFollowUpModelOutput("请你具体说明这里的锁粒度如何影响吞吐与延迟")
	if noFollowUp {
		t.Fatalf("expected follow-up question, got no-follow-up (reason=%s)", reason)
	}
	if question == "" {
		t.Fatalf("expected non-empty question")
	}
}

func TestParseFollowUpModelOutput_NoFollowUp(t *testing.T) {
	question, noFollowUp, _ := parseFollowUpModelOutput("NO_FOLLOWUP")
	if !noFollowUp {
		t.Fatalf("expected no-follow-up")
	}
	if question != "" {
		t.Fatalf("expected empty question, got %q", question)
	}
}

func TestParseFollowUpModelOutput_LegacyJSONCompatibility(t *testing.T) {
	raw := `{
  "follow_up_needed": true,
  "reason": "需要核实其并发一致性理解",
  "question": {
    "title": "并发一致性追问",
    "content": "在高并发下你会如何避免重复消费？",
    "expected_answer": "应说明幂等、去重与事务策略"
  }
}`

	question, noFollowUp, _ := parseFollowUpModelOutput(raw)
	if noFollowUp {
		t.Fatalf("expected follow-up question from legacy json")
	}
	if question == "" {
		t.Fatalf("expected non-empty question from legacy json")
	}
}
