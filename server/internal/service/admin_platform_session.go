package service

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	adminPlatformDefaultUsername = "admin"
	adminPlatformDefaultPassword = "admin123"
	adminPlatformSessionTTL      = 12 * time.Hour
)

type adminPlatformSession struct {
	Username  string
	ExpiresAt time.Time
}

var (
	adminPlatformSessionMu    sync.RWMutex
	adminPlatformSessionStore = map[string]adminPlatformSession{}
)

func ExpectedAdminPlatformCredentials() (string, string) {
	username := strings.TrimSpace(os.Getenv("ADMIN_PLATFORM_USERNAME"))
	password := strings.TrimSpace(os.Getenv("ADMIN_PLATFORM_PASSWORD"))
	if username == "" {
		username = adminPlatformDefaultUsername
	}
	if password == "" {
		password = adminPlatformDefaultPassword
	}
	return username, password
}

func VerifyAdminPlatformCredentials(username, password string) bool {
	expectedUsername, expectedPassword := ExpectedAdminPlatformCredentials()
	actualUsername := strings.TrimSpace(username)
	actualPassword := strings.TrimSpace(password)

	usernameMatch := subtle.ConstantTimeCompare([]byte(actualUsername), []byte(expectedUsername)) == 1
	passwordMatch := subtle.ConstantTimeCompare([]byte(actualPassword), []byte(expectedPassword)) == 1
	return usernameMatch && passwordMatch
}

func CreateAdminPlatformSession(username string) (string, time.Time, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", time.Time{}, err
	}
	token := hex.EncodeToString(tokenBytes)
	expiresAt := time.Now().Add(adminPlatformSessionTTL)

	adminPlatformSessionMu.Lock()
	adminPlatformSessionStore[token] = adminPlatformSession{
		Username:  strings.TrimSpace(username),
		ExpiresAt: expiresAt,
	}
	adminPlatformSessionMu.Unlock()

	return token, expiresAt, nil
}

func ValidateAdminPlatformSession(token string) (string, bool) {
	token = strings.TrimSpace(token)
	if token == "" {
		return "", false
	}

	adminPlatformSessionMu.RLock()
	session, ok := adminPlatformSessionStore[token]
	adminPlatformSessionMu.RUnlock()
	if !ok {
		return "", false
	}

	if time.Now().After(session.ExpiresAt) {
		adminPlatformSessionMu.Lock()
		delete(adminPlatformSessionStore, token)
		adminPlatformSessionMu.Unlock()
		return "", false
	}

	return session.Username, true
}

func RevokeAdminPlatformSession(token string) {
	token = strings.TrimSpace(token)
	if token == "" {
		return
	}

	adminPlatformSessionMu.Lock()
	delete(adminPlatformSessionStore, token)
	adminPlatformSessionMu.Unlock()
}

func AdminPlatformSessionTTL() time.Duration {
	return adminPlatformSessionTTL
}
