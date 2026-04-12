package service

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"your-project/config"
)

const (
	promptKnowledgeBaseDirEnv = "PROMPT_KNOWLEDGE_BASE_DIR"
	promptScanInterval        = 2 * time.Second
)

var promptVersionSuffixPattern = regexp.MustCompile(`(?i)^(.*?)[._-]?v(\d+)$`)

type promptFile struct {
	filePath   string
	version    string
	versionNum int
	modTime    time.Time
}

type promptCacheEntry struct {
	modTime time.Time
	size    int64
	content string
}

// PromptManager loads markdown prompts from knowledge_base and supports hot reload.
type PromptManager struct {
	baseDir      string
	scanInterval time.Duration

	mu       sync.RWMutex
	lastScan time.Time
	index    map[string][]promptFile
	cache    map[string]promptCacheEntry
}

func NewPromptManager() *PromptManager {
	return &PromptManager{
		baseDir:      resolvePromptBaseDir(),
		scanInterval: promptScanInterval,
		index:        make(map[string][]promptFile),
		cache:        make(map[string]promptCacheEntry),
	}
}

func (m *PromptManager) GetPrompt(key string) (string, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return "", fmt.Errorf("prompt key is empty")
	}

	promptKey, version := splitPromptKeyVersion(key)
	if version != "" {
		return m.GetPromptVersion(promptKey, version)
	}

	if looksLikePromptPath(promptKey) {
		return m.GetPromptByPath(promptKey)
	}

	if err := m.refreshIndexIfNeeded(); err != nil {
		return "", err
	}

	candidates := m.lookupPromptFiles(promptKey)
	if len(candidates) == 0 {
		return "", fmt.Errorf("prompt key not found: %s", promptKey)
	}

	selected, err := selectPromptByVersion(candidates, "")
	if err != nil {
		return "", err
	}
	return m.readPromptFile(selected.filePath)
}

func (m *PromptManager) GetPromptVersion(key, version string) (string, error) {
	key = strings.TrimSpace(key)
	version = strings.TrimSpace(version)
	if key == "" {
		return "", fmt.Errorf("prompt key is empty")
	}
	if version == "" {
		return m.GetPrompt(key)
	}

	if looksLikePromptPath(key) {
		return m.GetPromptByPath(key)
	}

	if err := m.refreshIndexIfNeeded(); err != nil {
		return "", err
	}

	candidates := m.lookupPromptFiles(key)
	if len(candidates) == 0 {
		return "", fmt.Errorf("prompt key not found: %s", key)
	}

	selected, err := selectPromptByVersion(candidates, version)
	if err != nil {
		return "", err
	}
	return m.readPromptFile(selected.filePath)
}

func (m *PromptManager) GetPromptByPath(promptPath string) (string, error) {
	promptPath = strings.TrimSpace(promptPath)
	if promptPath == "" {
		return "", fmt.Errorf("prompt path is empty")
	}

	if !strings.HasSuffix(strings.ToLower(promptPath), ".md") {
		promptPath += ".md"
	}

	resolved := promptPath
	if !filepath.IsAbs(resolved) {
		resolved = filepath.Join(m.baseDir, filepath.FromSlash(promptPath))
	}
	return m.readPromptFile(filepath.Clean(resolved))
}

func (m *PromptManager) RenderPrompt(key string, data interface{}) (string, error) {
	raw, err := m.GetPrompt(key)
	if err != nil {
		return "", err
	}

	tpl, err := template.New(normalizePromptKey(key)).Option("missingkey=error").Parse(raw)
	if err != nil {
		return "", fmt.Errorf("failed to parse prompt template %q: %w", key, err)
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render prompt template %q: %w", key, err)
	}
	return buf.String(), nil
}

func (m *PromptManager) refreshIndexIfNeeded() error {
	m.mu.RLock()
	shouldRefresh := time.Since(m.lastScan) >= m.scanInterval || len(m.index) == 0
	m.mu.RUnlock()
	if !shouldRefresh {
		return nil
	}

	files, err := m.scanPromptFiles()
	if err != nil {
		return err
	}

	nextIndex := make(map[string][]promptFile)
	for _, file := range files {
		for _, key := range file.keys {
			normalized := normalizePromptKey(key)
			if normalized == "" {
				continue
			}
			nextIndex[normalized] = append(nextIndex[normalized], promptFile{
				filePath:   file.filePath,
				version:    file.version,
				versionNum: file.versionNum,
				modTime:    file.modTime,
			})
		}
	}

	m.mu.Lock()
	m.index = nextIndex
	m.lastScan = time.Now()
	m.mu.Unlock()

	return nil
}

type scannedPromptFile struct {
	filePath   string
	version    string
	versionNum int
	modTime    time.Time
	keys       []string
}

func (m *PromptManager) scanPromptFiles() ([]scannedPromptFile, error) {
	if strings.TrimSpace(m.baseDir) == "" {
		return nil, fmt.Errorf("prompt base directory is empty")
	}

	entries := make([]scannedPromptFile, 0, 64)
	walkErr := filepath.WalkDir(m.baseDir, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.ToLower(filepath.Ext(d.Name())) != ".md" {
			return nil
		}

		info, statErr := d.Info()
		if statErr != nil {
			return statErr
		}

		relPath, relErr := filepath.Rel(m.baseDir, filePath)
		if relErr != nil {
			return relErr
		}

		relSlash := filepath.ToSlash(relPath)
		relNoExt := strings.TrimSuffix(relSlash, path.Ext(relSlash))
		dirPart, fileName := path.Split(relNoExt)
		keyName, version, versionNum := splitPromptFileVersion(fileName)

		canonicalKey := path.Clean(path.Join(dirPart, keyName))
		aliases := []string{canonicalKey, keyName}
		if strings.HasPrefix(canonicalKey, "prompts/") {
			aliases = append(aliases, strings.TrimPrefix(canonicalKey, "prompts/"))
		}

		dedup := make(map[string]struct{}, len(aliases))
		keys := make([]string, 0, len(aliases))
		for _, key := range aliases {
			normalized := normalizePromptKey(key)
			if normalized == "" {
				continue
			}
			if _, exists := dedup[normalized]; exists {
				continue
			}
			dedup[normalized] = struct{}{}
			keys = append(keys, normalized)
		}

		if len(keys) == 0 {
			return nil
		}

		entries = append(entries, scannedPromptFile{
			filePath:   filepath.Clean(filePath),
			version:    version,
			versionNum: versionNum,
			modTime:    info.ModTime(),
			keys:       keys,
		})
		return nil
	})
	if walkErr != nil {
		return nil, fmt.Errorf("failed to scan prompt directory %q: %w", m.baseDir, walkErr)
	}
	return entries, nil
}

func (m *PromptManager) lookupPromptFiles(key string) []promptFile {
	normalized := normalizePromptKey(key)
	m.mu.RLock()
	candidates := m.index[normalized]
	m.mu.RUnlock()
	if len(candidates) == 0 {
		return nil
	}

	copied := make([]promptFile, len(candidates))
	copy(copied, candidates)
	return copied
}

func selectPromptByVersion(candidates []promptFile, requestedVersion string) (promptFile, error) {
	if len(candidates) == 0 {
		return promptFile{}, fmt.Errorf("prompt candidate list is empty")
	}

	requestedVersion = normalizePromptVersion(requestedVersion)
	if requestedVersion != "" {
		matched := make([]promptFile, 0, len(candidates))
		for _, candidate := range candidates {
			if promptVersionMatch(candidate, requestedVersion) {
				matched = append(matched, candidate)
			}
		}
		if len(matched) == 0 {
			return promptFile{}, fmt.Errorf("prompt version %q not found", requestedVersion)
		}
		candidates = matched
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].versionNum != candidates[j].versionNum {
			return candidates[i].versionNum > candidates[j].versionNum
		}
		if !candidates[i].modTime.Equal(candidates[j].modTime) {
			return candidates[i].modTime.After(candidates[j].modTime)
		}
		return candidates[i].filePath < candidates[j].filePath
	})

	return candidates[0], nil
}

func (m *PromptManager) readPromptFile(filePath string) (string, error) {
	stat, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to stat prompt file %q: %w", filePath, err)
	}

	m.mu.RLock()
	cached, exists := m.cache[filePath]
	m.mu.RUnlock()
	if exists && cached.size == stat.Size() && cached.modTime.Equal(stat.ModTime()) {
		return cached.content, nil
	}

	raw, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt file %q: %w", filePath, err)
	}
	content := strings.TrimSpace(strings.ToValidUTF8(string(raw), ""))

	m.mu.Lock()
	m.cache[filePath] = promptCacheEntry{modTime: stat.ModTime(), size: stat.Size(), content: content}
	m.mu.Unlock()

	return content, nil
}

func splitPromptFileVersion(fileName string) (base string, version string, versionNum int) {
	trimmed := strings.TrimSpace(fileName)
	if trimmed == "" {
		return "", "", 0
	}

	matched := promptVersionSuffixPattern.FindStringSubmatch(trimmed)
	if len(matched) != 3 {
		return trimmed, "", 0
	}

	vNum, err := strconv.Atoi(matched[2])
	if err != nil {
		return trimmed, "", 0
	}
	base = strings.TrimSpace(matched[1])
	if base == "" {
		base = trimmed
	}
	return base, fmt.Sprintf("v%d", vNum), vNum
}

func splitPromptKeyVersion(key string) (promptKey string, version string) {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		return "", ""
	}
	parts := strings.Split(trimmed, "@")
	if len(parts) != 2 {
		return trimmed, ""
	}
	return strings.TrimSpace(parts[0]), normalizePromptVersion(parts[1])
}

func normalizePromptVersion(version string) string {
	normalized := strings.ToLower(strings.TrimSpace(version))
	if normalized == "" {
		return ""
	}
	if strings.HasPrefix(normalized, "v") {
		return normalized
	}
	return "v" + normalized
}

func promptVersionMatch(candidate promptFile, requested string) bool {
	if requested == "" {
		return true
	}
	normalizedCandidate := normalizePromptVersion(candidate.version)
	if normalizedCandidate == requested {
		return true
	}
	if candidate.versionNum > 0 {
		return fmt.Sprintf("v%d", candidate.versionNum) == requested
	}
	return false
}

func looksLikePromptPath(key string) bool {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		return false
	}
	if strings.HasSuffix(strings.ToLower(trimmed), ".md") {
		return true
	}
	return strings.Contains(trimmed, "/") || strings.Contains(trimmed, "\\")
}

func normalizePromptKey(key string) string {
	trimmed := strings.TrimSpace(strings.ToLower(filepath.ToSlash(key)))
	trimmed = strings.TrimPrefix(trimmed, "./")
	trimmed = strings.TrimPrefix(trimmed, "/")
	return strings.TrimSpace(path.Clean(trimmed))
}

func resolvePromptBaseDir() string {
	candidates := make([]string, 0, 8)

	if envPath := strings.TrimSpace(os.Getenv(promptKnowledgeBaseDirEnv)); envPath != "" {
		candidates = append(candidates, envPath)
	}

	cfg := config.GetConfig()
	if cfg != nil && strings.TrimSpace(cfg.Prompt.TemplatesDir) != "" {
		candidates = append(candidates, cfg.Prompt.TemplatesDir)
	}

	cwd, err := os.Getwd()
	if err == nil {
		candidates = append(candidates,
			filepath.Join(cwd, "knowledge_base", "prompts"),
			filepath.Join(cwd, "knowledge_base"),
			filepath.Join(cwd, "..", "knowledge_base", "prompts"),
			filepath.Join(cwd, "..", "knowledge_base"),
			filepath.Join(cwd, "..", "..", "knowledge_base", "prompts"),
			filepath.Join(cwd, "..", "..", "knowledge_base"),
		)
	}

	for _, candidate := range candidates {
		if strings.TrimSpace(candidate) == "" {
			continue
		}
		cleaned := filepath.Clean(candidate)
		stat, statErr := os.Stat(cleaned)
		if statErr != nil {
			continue
		}
		if stat.IsDir() {
			return cleaned
		}
	}

	if len(candidates) > 0 {
		return filepath.Clean(candidates[0])
	}
	return filepath.Clean("knowledge_base")
}
