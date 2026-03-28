package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"your-project/model"
	"your-project/repository"
)

func generateUUIDv4() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	buf[6] = (buf[6] & 0x0f) | 0x40
	buf[8] = (buf[8] & 0x3f) | 0x80

	hexStr := hex.EncodeToString(buf)
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hexStr[0:8],
		hexStr[8:12],
		hexStr[12:16],
		hexStr[16:20],
		hexStr[20:32],
	), nil
}

func generateInvitationCode() (string, error) {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return strings.ToUpper(hex.EncodeToString(buf)), nil
}

func ensureUserUUID(userRepo *repository.UserRepository, user *model.User) (string, error) {
	if user == nil {
		return "", fmt.Errorf("user is nil")
	}

	uuid := strings.TrimSpace(user.UUID)
	if uuid != "" {
		return uuid, nil
	}

	newUUID, err := generateUUIDv4()
	if err != nil {
		return "", err
	}

	user.UUID = newUUID
	if err := userRepo.Update(user); err != nil {
		return "", err
	}

	return newUUID, nil
}
