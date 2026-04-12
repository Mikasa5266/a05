package ai

import (
	"bytes"
	"fmt"
	"strings"
	"sync/atomic"
	"text/template"
)

var dynamicPromptManagerRegistry atomic.Value

func registerDynamicPromptManager(pm DynamicPromptManager) {
	dynamicPromptManagerRegistry.Store(pm)
}

func getRegisteredDynamicPromptManager() DynamicPromptManager {
	v := dynamicPromptManagerRegistry.Load()
	if v == nil {
		return nil
	}
	pm, ok := v.(DynamicPromptManager)
	if !ok {
		return nil
	}
	return pm
}

func tryRenderDynamicPrompt(key string, data interface{}) (string, bool) {
	pm := getRegisteredDynamicPromptManager()
	if pm == nil {
		return "", false
	}

	if rendered, err := pm.RenderPrompt(key, data); err == nil && strings.TrimSpace(rendered) != "" {
		return strings.TrimSpace(rendered), true
	}

	raw, err := pm.GetPrompt(key)
	if err != nil || strings.TrimSpace(raw) == "" {
		return "", false
	}

	tpl, err := template.New(key).Option("missingkey=error").Parse(raw)
	if err != nil {
		return "", false
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", false
	}

	result := strings.TrimSpace(buf.String())
	if result == "" {
		return "", false
	}
	return result, true
}

func getPromptOrFallback(key, fallback string) string {
	pm := getRegisteredDynamicPromptManager()
	if pm == nil {
		return fallback
	}

	prompt, err := pm.GetPrompt(key)
	if err != nil {
		return fallback
	}
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return fallback
	}
	return prompt
}

func formatPromptFallback(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}
