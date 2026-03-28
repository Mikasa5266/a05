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
	root := template.New("prompts").Option("missingkey=error")
	loaded := 0
	err := filepath.WalkDir(templatesDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || filepath.Ext(d.Name()) != ".tmpl" {
			return nil
		}

		relPath, relErr := filepath.Rel(templatesDir, path)
		if relErr != nil {
			return relErr
		}
		templateName := filepath.ToSlash(relPath)

		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("failed to read template %q: %w", path, readErr)
		}
		if _, parseErr := root.New(templateName).Parse(string(content)); parseErr != nil {
			return fmt.Errorf("failed to parse template %q: %w", path, parseErr)
		}
		loaded++
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk templates dir %q: %w", templatesDir, err)
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
