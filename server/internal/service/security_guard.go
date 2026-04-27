package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"your-project/config"
	"your-project/internal/model"
	"your-project/internal/repository"

	"github.com/gin-gonic/gin"
)

var defaultBlockedWords = []string{
	"恐怖主义",
	"爆炸物制作",
	"仇恨言论",
	"极端组织",
	"贩毒",
	"人口贩卖",
	"煽动暴力",
	"组织自杀",
}

var (
	sqlInjectionPattern = regexp.MustCompile(`(?i)(union\s+select|or\s+1\s*=\s*1|and\s+1\s*=\s*1|sleep\s*\(|benchmark\s*\(|drop\s+table|truncate\s+table|insert\s+into\s+\w+\s+select|--\s|/\*|\*/)`)
	xssPattern          = regexp.MustCompile(`(?i)(<\s*script\b|javascript:|onerror\s*=|onload\s*=|<\s*iframe\b|<\s*svg\b|<\s*img\b[^>]*onerror)`)
)

var (
	blockedWordsOnce sync.Once
	blockedWords     []string

	rateLimitMu      sync.Mutex
	rateLimitBuckets = make(map[string][]time.Time)

	securityLogKeyOnce sync.Once
	securityLogKey     []byte
)

func loadBlockedWords() []string {
	blockedWordsOnce.Do(func() {
		raw := strings.TrimSpace(os.Getenv("SECURITY_BLOCKED_WORDS"))
		if raw == "" {
			blockedWords = append([]string(nil), defaultBlockedWords...)
			return
		}

		parts := strings.FieldsFunc(raw, func(r rune) bool {
			return r == ',' || r == ';' || r == '|' || r == '\n'
		})
		seen := make(map[string]struct{}, len(parts))
		normalized := make([]string, 0, len(parts))
		for _, part := range parts {
			item := strings.ToLower(strings.TrimSpace(part))
			if item == "" {
				continue
			}
			if _, ok := seen[item]; ok {
				continue
			}
			seen[item] = struct{}{}
			normalized = append(normalized, item)
		}
		if len(normalized) == 0 {
			blockedWords = append([]string(nil), defaultBlockedWords...)
			return
		}
		sort.Strings(normalized)
		blockedWords = normalized
	})

	return append([]string(nil), blockedWords...)
}

func MatchBlockedWords(content string) []string {
	normalized := strings.ToLower(strings.TrimSpace(content))
	if normalized == "" {
		return []string{}
	}

	hits := make([]string, 0, 4)
	for _, word := range loadBlockedWords() {
		if word == "" {
			continue
		}
		if strings.Contains(normalized, strings.ToLower(word)) {
			hits = append(hits, word)
		}
	}
	if len(hits) == 0 {
		return []string{}
	}
	sort.Strings(hits)
	return hits
}

func ValidateCommunityContent(content string) error {
	if attackType, found := DetectSecurityThreat(content); found {
		return fmt.Errorf("内容存在%s风险特征，请修改后重试", attackType)
	}

	hits := MatchBlockedWords(content)
	if len(hits) == 0 {
		return nil
	}
	if len(hits) > 3 {
		hits = hits[:3]
	}
	return fmt.Errorf("内容包含违规关键词，请修改后重试（命中：%s）", strings.Join(hits, "、"))
}

func DetectSecurityThreat(content string) (string, bool) {
	normalized := strings.ToLower(strings.TrimSpace(content))
	if normalized == "" {
		return "", false
	}
	if sqlInjectionPattern.MatchString(normalized) {
		return "SQL注入", true
	}
	if xssPattern.MatchString(normalized) {
		return "XSS", true
	}
	return "", false
}

func ValidateSafeTextField(fieldName, value string) error {
	if attackType, found := DetectSecurityThreat(value); found {
		return fmt.Errorf("%s包含%s风险特征", fieldName, attackType)
	}
	return nil
}

func SanitizeUserInput(value string) string {
	return html.EscapeString(strings.TrimSpace(value))
}

func AllowRateLimitedAction(userID uint, action string, limit int, window time.Duration) bool {
	if userID == 0 || strings.TrimSpace(action) == "" || limit <= 0 || window <= 0 {
		return true
	}

	key := fmt.Sprintf("%d:%s", userID, strings.TrimSpace(action))
	now := time.Now()
	windowStart := now.Add(-window)

	rateLimitMu.Lock()
	defer rateLimitMu.Unlock()

	records := rateLimitBuckets[key]
	kept := records[:0]
	for _, ts := range records {
		if ts.After(windowStart) {
			kept = append(kept, ts)
		}
	}

	if len(kept) >= limit {
		rateLimitBuckets[key] = kept
		return false
	}

	rateLimitBuckets[key] = append(kept, now)
	return true
}

func RecordSecurityAudit(c *gin.Context, userID uint, action, outcome string, statusCode int, detail map[string]interface{}) {
	if c == nil {
		return
	}

	detailJSON := "{}"
	if len(detail) > 0 {
		if payload, err := json.Marshal(detail); err == nil {
			detailJSON = string(payload)
		}
	}

	sourceIP, sourcePort := splitHostPort(c.Request.RemoteAddr)
	targetHost, targetPort := splitHostPort(c.Request.Host)
	userAgent := truncateString(strings.TrimSpace(c.Request.UserAgent()), 255)

	path := strings.TrimSpace(c.FullPath())
	if path == "" {
		path = strings.TrimSpace(c.Request.URL.Path)
	}

	entry := model.SecurityAuditLog{
		UserID:            userID,
		Action:            truncateString(strings.TrimSpace(action), 80),
		Outcome:           truncateString(strings.TrimSpace(outcome), 20),
		Method:            truncateString(strings.TrimSpace(c.Request.Method), 10),
		Path:              truncateString(path, 255),
		StatusCode:        statusCode,
		SourceIP:          truncateString(ProtectSecurityLogField(sourceIP), 255),
		SourcePort:        truncateString(ProtectSecurityLogField(sourcePort), 255),
		TargetHost:        truncateString(ProtectSecurityLogField(targetHost), 255),
		TargetPort:        truncateString(ProtectSecurityLogField(targetPort), 255),
		ClientFingerprint: userAgent,
		DetailJSON:        detailJSON,
	}

	_ = repository.GetDB().Create(&entry).Error
}

func ProtectSecurityLogField(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	key := resolveSecurityLogEncryptKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return trimmed
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return trimmed
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return trimmed
	}
	ciphertext := gcm.Seal(nil, nonce, []byte(trimmed), nil)
	payload := append(nonce, ciphertext...)
	return "enc:v1:" + base64.StdEncoding.EncodeToString(payload)
}

func RevealSecurityLogField(value string) string {
	trimmed := strings.TrimSpace(value)
	if !strings.HasPrefix(trimmed, "enc:v1:") {
		return trimmed
	}

	raw := strings.TrimPrefix(trimmed, "enc:v1:")
	payload, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return trimmed
	}

	key := resolveSecurityLogEncryptKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return trimmed
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return trimmed
	}

	nonceSize := gcm.NonceSize()
	if len(payload) <= nonceSize {
		return trimmed
	}
	nonce := payload[:nonceSize]
	ciphertext := payload[nonceSize:]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return trimmed
	}
	return string(plain)
}

func resolveSecurityLogEncryptKey() []byte {
	securityLogKeyOnce.Do(func() {
		keyMaterial := strings.TrimSpace(os.Getenv("SECURITY_LOG_ENCRYPT_KEY"))
		if keyMaterial == "" {
			cfg := config.GetConfig()
			if cfg != nil {
				keyMaterial = strings.TrimSpace(cfg.JWT.Secret)
			}
		}
		if keyMaterial == "" {
			keyMaterial = "a05-security-log-default-key-change-me"
		}
		sum := sha256.Sum256([]byte(keyMaterial))
		securityLogKey = sum[:]
	})
	return securityLogKey
}

func splitHostPort(raw string) (string, string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ""
	}
	host, port, err := net.SplitHostPort(raw)
	if err == nil {
		return strings.TrimSpace(host), strings.TrimSpace(port)
	}
	return raw, ""
}

func truncateString(value string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return string(runes[:max])
}
