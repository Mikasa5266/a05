package ai

import "strings"

func stripOptionalCodeFence(text string) string {
	trimmed := strings.TrimSpace(text)
	if !strings.HasPrefix(trimmed, "```") {
		return trimmed
	}
	lines := strings.Split(trimmed, "\n")
	if len(lines) < 2 {
		return strings.Trim(trimmed, "`")
	}
	if strings.HasPrefix(strings.TrimSpace(lines[len(lines)-1]), "```") {
		return strings.TrimSpace(strings.Join(lines[1:len(lines)-1], "\n"))
	}
	return strings.TrimSpace(strings.Join(lines[1:], "\n"))
}

func normalizeGroundedEvalResponse(parsed groundedEvalLLMResponse) groundedEvalLLMResponse {
	parsed.Score = clampScore(parsed.Score)
	parsed.Comment = strings.TrimSpace(parsed.Comment)
	parsed.Reasoning = strings.TrimSpace(parsed.Reasoning)
	parsed.FollowUpContext = strings.TrimSpace(parsed.FollowUpContext)

	if parsed.Comment == "" && parsed.Reasoning != "" {
		parsed.Comment = parsed.Reasoning
	}

	if parsed.ShouldFollowUpRaw != nil {
		parsed.ShouldFollowUp = *parsed.ShouldFollowUpRaw
		if !parsed.ShouldFollowUp {
			parsed.FollowUpContext = ""
		}
	} else {
		parsed.ShouldFollowUp = parsed.FollowUpContext != ""
	}

	return parsed
}

func truncateEvalRunes(text string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(strings.TrimSpace(text))
	if len(runes) <= max {
		return string(runes)
	}
	return strings.TrimSpace(string(runes[:max]))
}

func splitSuggestionText(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return []string{"请补充核心原理、实现细节与可验证结果。"}
	}
	parts := strings.FieldsFunc(text, func(r rune) bool {
		return r == ';' || r == '；' || r == '\n'
	})
	result := make([]string, 0, len(parts))
	for _, p := range parts {
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

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
