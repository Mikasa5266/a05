package prompt

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"text/template"
)

// PromptManager loads and renders text templates for LLM prompts.
type PromptManager struct {
	tpl *template.Template
}

func NewPromptManager() (*PromptManager, error) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("failed to resolve prompt manager path")
	}

	templatesDir := filepath.Join(filepath.Dir(currentFile), "templates")
	entries, err := os.ReadDir(templatesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read templates dir %q: %w", templatesDir, err)
	}

	root := template.New("prompts").Option("missingkey=error")
	loaded := 0
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".tmpl" {
			continue
		}

		fullPath := filepath.Join(templatesDir, entry.Name())
		content, readErr := os.ReadFile(fullPath)
		if readErr != nil {
			return nil, fmt.Errorf("failed to read template %q: %w", fullPath, readErr)
		}
		if _, parseErr := root.New(entry.Name()).Parse(string(content)); parseErr != nil {
			return nil, fmt.Errorf("failed to parse template %q: %w", fullPath, parseErr)
		}
		loaded++
	}

	if loaded == 0 {
		return nil, fmt.Errorf("no .tmpl files found in %q", templatesDir)
	}

	return &PromptManager{tpl: root}, nil
}

func (m *PromptManager) Render(templateName string, data interface{}) (string, error) {
	if m == nil || m.tpl == nil {
		return "", fmt.Errorf("prompt manager is not initialized")
	}

	var buf bytes.Buffer
	if err := m.tpl.ExecuteTemplate(&buf, templateName, data); err != nil {
		return "", fmt.Errorf("failed to render template %q: %w", templateName, err)
	}
	return buf.String(), nil
}
