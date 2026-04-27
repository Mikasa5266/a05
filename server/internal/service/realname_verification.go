package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	chinaMainlandPhonePattern = regexp.MustCompile(`^1[3-9][0-9]{9}$`)
	chinaIDCardPattern        = regexp.MustCompile(`^[0-9]{17}[0-9Xx]$`)
)

var idCardChecksumWeights = []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
var idCardChecksumCodes = []byte("10X98765432")

type VerifiedIdentity struct {
	RealName   string
	Phone      string
	IDCardHash string
	IDCardMask string
	VerifiedAt time.Time
}

func BuildVerifiedIdentity(realName, phone, idCardNo string) (*VerifiedIdentity, error) {
	normalizedName := strings.TrimSpace(realName)
	if len([]rune(normalizedName)) < 2 {
		return nil, fmt.Errorf("real_name is invalid")
	}
	if len([]rune(normalizedName)) > 80 {
		return nil, fmt.Errorf("real_name is too long")
	}
	if attackType, found := DetectSecurityThreat(normalizedName); found {
		return nil, fmt.Errorf("real_name contains %s risk", attackType)
	}

	normalizedPhone := normalizePhone(phone)
	if !chinaMainlandPhonePattern.MatchString(normalizedPhone) {
		return nil, fmt.Errorf("phone is invalid")
	}

	normalizedIDCard := strings.ToUpper(strings.TrimSpace(idCardNo))
	if !chinaIDCardPattern.MatchString(normalizedIDCard) {
		return nil, fmt.Errorf("id_card_no is invalid")
	}
	if !isValidChinaIDCardChecksum(normalizedIDCard) {
		return nil, fmt.Errorf("id_card_no checksum is invalid")
	}

	hash := sha256.Sum256([]byte(normalizedIDCard))

	return &VerifiedIdentity{
		RealName:   normalizedName,
		Phone:      normalizedPhone,
		IDCardHash: hex.EncodeToString(hash[:]),
		IDCardMask: maskIDCard(normalizedIDCard),
		VerifiedAt: time.Now(),
	}, nil
}

func normalizePhone(phone string) string {
	replacer := strings.NewReplacer(" ", "", "-", "", "(", "", ")", "")
	return strings.TrimSpace(replacer.Replace(phone))
}

func maskIDCard(idCard string) string {
	if len(idCard) < 10 {
		return ""
	}
	return idCard[:6] + "********" + idCard[len(idCard)-4:]
}

func isValidChinaIDCardChecksum(idCard string) bool {
	if len(idCard) != 18 {
		return false
	}
	sum := 0
	for i := 0; i < 17; i++ {
		digit := idCard[i] - '0'
		if digit > 9 {
			return false
		}
		sum += int(digit) * idCardChecksumWeights[i]
	}
	code := idCardChecksumCodes[sum%11]
	return code == idCard[17]
}
