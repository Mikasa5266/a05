package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

func normalizeShadowHintText(text string, fallback string) string {
	clean := strings.TrimSpace(strings.ReplaceAll(text, "\n", " "))
	clean = strings.Join(strings.Fields(clean), " ")
	if clean == "" {
		clean = fallback
	}
	runes := []rune(clean)
	if len(runes) > 72 {
		clean = strings.TrimSpace(string(runes[:72]))
	}
	return clean
}

func extractShadowHintFocus(question string) string {
	trimmed := strings.TrimSpace(question)
	if trimmed == "" {
		return "this question"
	}
	replacer := strings.NewReplacer(
		"please", "",
		"how", "",
		"what", "",
		"why", "",
		"?", "",
	)
	focus := strings.TrimSpace(replacer.Replace(strings.ToLower(trimmed)))
	if focus == "" {
		focus = trimmed
	}
	runes := []rune(focus)
	if len(runes) > 18 {
		focus = string(runes[:18])
	}
	if strings.TrimSpace(focus) == "" {
		return "this question"
	}
	return focus
}

func buildShadowHintFallbacks(question string) []string {
	focus := extractShadowHintFocus(question)
	return []string{
		fmt.Sprintf("Give a short judgment around '%s' first.", focus),
		fmt.Sprintf("Split '%s' into mechanism and a concrete scenario.", focus),
		fmt.Sprintf("Start with viewpoint -> mechanism -> result around '%s'.", focus),
	}
}

func looksTemplateLikeHint(text string) bool {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return true
	}
	templateMarkers := []string{"three-step", "four-step", "template", "conclusion"}
	for _, marker := range templateMarkers {
		if strings.Contains(strings.ToLower(trimmed), marker) {
			return true
		}
	}
	return false
}

func extractShadowHintAnchors(referenceAnswer, knowledgeContext string) []string {
	merged := strings.TrimSpace(referenceAnswer + "\n" + knowledgeContext)
	if merged == "" {
		return nil
	}
	segments := strings.FieldsFunc(merged, func(r rune) bool {
		switch r {
		case '\n', '.', ';', ',', ':', '(', ')', '。', '；', '，', '：', '（', '）', '、':
			return true
		default:
			return false
		}
	})

	anchors := make([]string, 0, 6)
	seen := map[string]bool{}
	for _, seg := range segments {
		item := strings.TrimSpace(seg)
		if item == "" {
			continue
		}
		runes := []rune(item)
		if len(runes) < 4 {
			continue
		}
		if len(runes) > 20 {
			item = string(runes[:20])
		}
		if seen[item] {
			continue
		}
		seen[item] = true
		anchors = append(anchors, item)
		if len(anchors) >= 6 {
			break
		}
	}
	return anchors
}

func containsAnyShadowAnchor(hint string, anchors []string) bool {
	if strings.TrimSpace(hint) == "" || len(anchors) == 0 {
		return false
	}
	for _, anchor := range anchors {
		if strings.TrimSpace(anchor) == "" {
			continue
		}
		if strings.Contains(hint, anchor) {
			return true
		}
	}
	return false
}

func (s *AIService) GenerateShadowCoachHintLevels(ctx context.Context, position, question, transcript, style, referenceAnswer, knowledgeContext string) ([]string, error) {
	trimmedQuestion := strings.TrimSpace(question)
	fallbacks := buildShadowHintFallbacks(trimmedQuestion)
	anchors := extractShadowHintAnchors(referenceAnswer, knowledgeContext)
	anchorText := "none"
	if len(anchors) > 0 {
		anchorText = strings.Join(anchors, ",")
	}
	if trimmedQuestion == "" {
		return fallbacks, nil
	}

	prompt, err := s.renderPrompt("shadow_coach_hint_levels.tmpl", map[string]interface{}{
		"Position":         position,
		"Style":            style,
		"Question":         trimmedQuestion,
		"Transcript":       strings.TrimSpace(transcript),
		"ReferenceAnswer":  strings.TrimSpace(referenceAnswer),
		"KnowledgeContext": strings.TrimSpace(knowledgeContext),
		"AnchorText":       anchorText,
	})
	if err != nil {
		return fallbacks, nil
	}

	raw, err := s.chat(ctx, prompt, "shadow_hint", jsonObjectResponseFormat())
	if err != nil {
		return fallbacks, nil
	}

	var parsed struct {
		Level1 string `json:"level_1"`
		Level2 string `json:"level_2"`
		Level3 string `json:"level_3"`
	}
	if unmarshalErr := json.Unmarshal([]byte(raw), &parsed); unmarshalErr != nil {
		return fallbacks, nil
	}

	hints := []string{
		normalizeShadowHintText(parsed.Level1, fallbacks[0]),
		normalizeShadowHintText(parsed.Level2, fallbacks[1]),
		normalizeShadowHintText(parsed.Level3, fallbacks[2]),
	}

	for idx := range hints {
		if looksTemplateLikeHint(hints[idx]) {
			hints[idx] = fallbacks[idx]
		}
	}

	if hints[1] == hints[0] {
		hints[1] = fallbacks[1]
	}
	if hints[2] == hints[1] || hints[2] == hints[0] {
		hints[2] = fallbacks[2]
	}

	if len(anchors) > 0 {
		if !containsAnyShadowAnchor(hints[1], anchors) {
			hints[1] = normalizeShadowHintText(
				fmt.Sprintf("Mention '%s' and add one mechanism + action.", anchors[0]),
				fallbacks[1],
			)
		}
		if !containsAnyShadowAnchor(hints[2], anchors) {
			hints[2] = normalizeShadowHintText(
				fmt.Sprintf("Speak in viewpoint -> mechanism -> result and include '%s'.", anchors[0]),
				fallbacks[2],
			)
		}
	}

	return hints, nil
}

func (s *AIService) GenerateShadowCoachHint(ctx context.Context, position, question, transcript, style string, silenceSeconds int) (string, error) {
	hints, err := s.GenerateShadowCoachHintLevels(ctx, position, question, transcript, style, "", "")
	if err != nil {
		return buildShadowHintFallbacks(question)[0], nil
	}
	if len(hints) == 0 {
		return buildShadowHintFallbacks(question)[0], nil
	}
	if silenceSeconds >= 60 && len(hints) >= 3 {
		return hints[2], nil
	}
	if silenceSeconds >= 40 && len(hints) >= 2 {
		return hints[1], nil
	}
	return hints[0], nil
}
