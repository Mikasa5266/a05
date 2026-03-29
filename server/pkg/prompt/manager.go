package prompt

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
	"text/template"

	"your-project/config"
)

const promptTemplatesDirEnv = "PROMPT_TEMPLATES_DIR"

//go:embed templates/*
var embeddedTemplatesFS embed.FS

// PromptManager loads and renders text templates for LLM prompts.
type PromptManager struct {
	tpl *template.Template
}

func NewPromptManager() (*PromptManager, error) {
	root := template.New("prompts").Option("missingkey=error")

	configuredDir, hasConfiguredDir := resolveConfiguredTemplatesDir()
	if hasConfiguredDir {
		loaded, err := loadTemplates(root, os.DirFS(configuredDir), ".")
		if err == nil && loaded > 0 {
			return &PromptManager{tpl: root}, nil
		}

		root = template.New("prompts").Option("missingkey=error")
	}

	loaded, err := loadTemplates(root, embeddedTemplatesFS, "templates")
	if err != nil {
		if hasConfiguredDir {
			return nil, fmt.Errorf("failed to load embedded templates after configured dir %q: %w", configuredDir, err)
		}
		return nil, fmt.Errorf("failed to load embedded templates: %w", err)
	}

	if loaded == 0 {
		if hasConfiguredDir {
			return nil, fmt.Errorf("no .tmpl files found in configured dir %q or embedded templates", configuredDir)
		}
		return nil, fmt.Errorf("no .tmpl files found in embedded templates")
	}

	return &PromptManager{tpl: root}, nil
}

func resolveConfiguredTemplatesDir() (string, bool) {
	if fromEnv := strings.TrimSpace(os.Getenv(promptTemplatesDirEnv)); fromEnv != "" {
		return fromEnv, true
	}

	cfg := config.GetConfig()
	if cfg == nil {
		return "", false
	}

	fromConfig := strings.TrimSpace(cfg.Prompt.TemplatesDir)
	if fromConfig == "" {
		return "", false
	}

	return fromConfig, true
}

func loadTemplates(root *template.Template, sourceFS fs.FS, baseDir string) (int, error) {
	loaded := 0
	err := fs.WalkDir(sourceFS, baseDir, func(currentPath string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || path.Ext(d.Name()) != ".tmpl" {
			return nil
		}

		templateName := strings.TrimPrefix(currentPath, baseDir)
		templateName = strings.TrimPrefix(templateName, "/")
		templateName = path.Clean(templateName)
		if templateName == "." || templateName == "" {
			return fmt.Errorf("invalid template path %q", currentPath)
		}

		content, readErr := fs.ReadFile(sourceFS, currentPath)
		if readErr != nil {
			return fmt.Errorf("failed to read template %q: %w", currentPath, readErr)
		}
		if _, parseErr := root.New(templateName).Parse(string(content)); parseErr != nil {
			return fmt.Errorf("failed to parse template %q: %w", currentPath, parseErr)
		}

		loaded++
		return nil
	})
	if err != nil {
		return 0, err
	}

	return loaded, nil
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
