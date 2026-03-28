package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"your-project/model"
)

func (s *AIService) EnsureChineseOutput(ctx context.Context, text, fallback string) string {
	normalized := normalizeFeedbackText(text)
	safeFallback := normalizeFeedbackText(fallback)
	if safeFallback == "" {
		safeFallback = "回答内容已收到，建议补充更具体的技术细节与实践案例。"
	}
	if normalized == "" {
		return safeFallback
	}
	if isMostlyChinese(normalized, 0.45) {
		return normalized
	}

	prompt := fmt.Sprintf("Rewrite the following content into natural simplified Chinese only:\n\n%s", normalized)
	rewritten, err := s.chat(ctx, prompt, "chat", nil)
	if err != nil {
		return safeFallback
	}
	rewritten = normalizeFeedbackText(rewritten)
	if rewritten == "" || !isMostlyChinese(rewritten, 0.45) {
		return safeFallback
	}
	return rewritten
}

func (s *AIService) EnsureQuestionChinese(ctx context.Context, question *model.Question) {
	if question == nil {
		return
	}
	sanitizeQuestionTextFields(question)
	needRewrite := !isMostlyChinese(question.Title, 0.3) || !isMostlyChinese(question.Content, 0.35) || !isMostlyChinese(question.ExpectedAnswer, 0.3)
	if !needRewrite {
		sanitizeQuestionTextFields(question)
		return
	}

	prompt, err := s.renderPrompt("ensure_question_chinese.tmpl", map[string]interface{}{"Title": question.Title, "Content": question.Content, "ExpectedAnswer": question.ExpectedAnswer})
	if err == nil {
		response, chatErr := s.chat(ctx, prompt, "chat", jsonObjectResponseFormat())
		if chatErr == nil {
			var localized struct {
				Title          string `json:"title"`
				Content        string `json:"content"`
				ExpectedAnswer string `json:"expected_answer"`
			}
			if unmarshalErr := json.Unmarshal([]byte(sanitizeGeneratedText(response)), &localized); unmarshalErr == nil {
				if strings.TrimSpace(localized.Title) != "" {
					question.Title = sanitizeGeneratedText(localized.Title)
				}
				if strings.TrimSpace(localized.Content) != "" {
					question.Content = sanitizeGeneratedText(localized.Content)
				}
				if strings.TrimSpace(localized.ExpectedAnswer) != "" {
					question.ExpectedAnswer = sanitizeGeneratedText(localized.ExpectedAnswer)
				}
			}
		}
	}

	if !isMostlyChinese(question.Title, 0.3) {
		question.Title = "技术问题"
	}
	if !isMostlyChinese(question.Content, 0.35) {
		question.Content = "请结合实际项目经验，系统说明你的思路、关键实现和取舍。"
	}
	if !isMostlyChinese(question.ExpectedAnswer, 0.3) {
		question.ExpectedAnswer = "回答应包含核心原理、实现步骤、关键细节与风险边界。"
	}
	sanitizeQuestionTextFields(question)
}

func normalizeFeedbackText(s string) string {
	text := sanitizeGeneratedText(strings.TrimSpace(s))
	if text == "" {
		return "回答内容已收到，建议补充更具体的技术细节与实践案例。"
	}
	if strings.HasPrefix(text, "{") || strings.HasPrefix(text, "[") {
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(text), &obj); err == nil {
			keys := []string{"feedback", "analysis", "comment", "summary", "advice", "suggestion"}
			parts := make([]string, 0, len(keys))
			for _, key := range keys {
				if v, ok := obj[key]; ok {
					if line, ok := v.(string); ok && strings.TrimSpace(line) != "" {
						parts = append(parts, sanitizeGeneratedText(strings.TrimSpace(line)))
					}
				}
			}
			if len(parts) > 0 {
				text = strings.Join(parts, "\n")
			}
		}
	}
	re := regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, " ")
	text = strings.TrimSpace(strings.Trim(text, "`"))
	return sanitizeGeneratedText(text)
}

func sanitizeGeneratedText(s string) string {
	if strings.TrimSpace(s) == "" {
		return ""
	}

	cleaned := strings.ToValidUTF8(s, "")
	cleaned = strings.ReplaceAll(cleaned, "\uFFFD", "")
	cleaned = strings.ReplaceAll(cleaned, "\x00", "")
	cleaned = strings.ReplaceAll(cleaned, "\uFEFF", "")
	cleaned = strings.ReplaceAll(cleaned, "\u200B", "")
	return strings.TrimSpace(cleaned)
}

func sanitizeQuestionTextFields(question *model.Question) {
	if question == nil {
		return
	}

	question.Title = sanitizeGeneratedText(question.Title)
	question.Content = sanitizeGeneratedText(question.Content)
	question.ExpectedAnswer = sanitizeGeneratedText(question.ExpectedAnswer)
}

func isMostlyChinese(text string, ratio float64) bool {
	content := strings.TrimSpace(text)
	if content == "" {
		return false
	}
	hanCount := 0
	letterCount := 0
	for _, r := range content {
		if unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsDigit(r) {
			continue
		}
		if unicode.IsLetter(r) {
			letterCount++
		}
		if r >= 0x4E00 && r <= 0x9FFF {
			hanCount++
		}
	}
	if letterCount == 0 {
		return false
	}
	return float64(hanCount)/float64(letterCount) >= ratio
}
