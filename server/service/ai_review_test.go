package service

import "testing"

func TestEvaluateCandidateAnswer_HardFailForOpenTechnicalQuestion(t *testing.T) {
	question := "在10万条学生记录里，如何快速找出年龄大于20岁的学生并去重？请结合Java集合框架说明思路。"
	expected := "可使用Stream filter(age>20)+collect(toCollection(LinkedHashSet))或HashSet去重，并分析时间复杂度。"
	answer := "她是学生。"

	llm := func(prompt string) (string, error) {
		return `{
  "score": 92,
  "dimensions": {
    "technical_depth": 90,
    "expression": 94,
    "logic": 89,
    "completeness": 93
  },
  "comment": "回答完整。",
  "suggestion": "继续保持。"
}`, nil
	}

	result, err := EvaluateCandidateAnswer(question, expected, answer, llm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Score != 0 {
		t.Fatalf("expected score 0, got %d", result.Score)
	}
}

func TestEvaluateCandidateAnswer_KeepsReasonableScoreForSubstantiveAnswer(t *testing.T) {
	question := "如何在Java中对List去重并保持原有顺序？"
	expected := "使用LinkedHashSet保持插入顺序，再转回List；或使用Stream distinct。"
	answer := "我会先遍历List，使用LinkedHashSet去重以保持顺序，然后再转成List返回。时间复杂度约O(n)。"

	llm := func(prompt string) (string, error) {
		return `{
  "score": 75,
  "dimensions": {
    "technical_depth": 76,
    "expression": 73,
    "logic": 74,
    "completeness": 75
  },
  "comment": "回答较完整。",
  "suggestion": "可补充边界场景。"
}`, nil
	}

	result, err := EvaluateCandidateAnswer(question, expected, answer, llm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Score <= 0 {
		t.Fatalf("expected positive score for substantive answer, got %d", result.Score)
	}
}

