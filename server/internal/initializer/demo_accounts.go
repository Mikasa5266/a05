package initializer

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"your-project/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type demoAccountSeed struct {
	Username       string
	Email          string
	Role           string
	Password       string
	Avatar         string
	CompanyName    string
	UniversityName string
}

func EnsureDemoAccounts(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database is nil")
	}

	seeds := []demoAccountSeed{
	// 学生组（三字用户名，贴合日常，不土气）
	{Username: "林知许", Email: "student01@test.com", Role: "student", Password: "123456", Avatar: "https://api.dicebear.com/7.x/adventurer/svg?seed=student01"},
	{Username: "苏念安", Email: "student02@test.com", Role: "student", Password: "123456", Avatar: "https://api.dicebear.com/7.x/adventurer/svg?seed=student02"},
	{Username: "陆景然", Email: "student03@test.com", Role: "student", Password: "123456", Avatar: "https://api.dicebear.com/7.x/adventurer/svg?seed=student03"},
	{Username: "许清和", Email: "student04@test.com", Role: "student", Password: "123456", Avatar: "https://api.dicebear.com/7.x/adventurer/svg?seed=student04"},
	{Username: "温予安", Email: "student05@test.com", Role: "student", Password: "123456", Avatar: "https://api.dicebear.com/7.x/adventurer/svg?seed=student05"},

	// 企业组（三字公司名，正规企业命名，不浮夸）
	{Username: "星联科技", Email: "enterprise01@test.com", Role: "enterprise", Password: "123456", Avatar: "https://api.dicebear.com/7.x/shapes/svg?seed=enterprise01", CompanyName: "星联科技有限公司"},
	{Username: "元启数据", Email: "enterprise02@test.com", Role: "enterprise", Password: "123456", Avatar: "https://api.dicebear.com/7.x/shapes/svg?seed=enterprise02", CompanyName: "元启数据科技有限公司"},
	{Username: "海跃软件", Email: "enterprise03@test.com", Role: "enterprise", Password: "123456", Avatar: "https://api.dicebear.com/7.x/shapes/svg?seed=enterprise03", CompanyName: "海跃软件股份有限公司"},

	// 高校组（正规大学名，贴合“正经”需求）
	{Username: "江南大学", Email: "university01@test.com", Role: "university", Password: "123456", Avatar: "https://api.dicebear.com/7.x/shapes/svg?seed=university01", UniversityName: "江南大学"},
	{Username: "华南理工", Email: "university02@test.com", Role: "university", Password: "123456", Avatar: "https://api.dicebear.com/7.x/shapes/svg?seed=university02", UniversityName: "华南理工大学"},
}

	return db.Transaction(func(tx *gorm.DB) error {
		for _, seed := range seeds {
			if err := upsertDemoAccount(tx, seed); err != nil {
				return err
			}
		}
		return nil
	})
}

func upsertDemoAccount(tx *gorm.DB, seed demoAccountSeed) error {
	email := strings.TrimSpace(strings.ToLower(seed.Email))
	if email == "" {
		return fmt.Errorf("invalid demo account email")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(seed.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash demo account password: %w", err)
	}

	var user model.User
	err = tx.Where("email = ?", email).First(&user).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return fmt.Errorf("failed to query demo user %s: %w", email, err)
		}

		uuid, uuidErr := generateUUIDv4()
		if uuidErr != nil {
			return fmt.Errorf("failed to generate user uuid for %s: %w", email, uuidErr)
		}

		user = model.User{
			UUID:     uuid,
			Username: seed.Username,
			Email:    email,
			Password: string(hashedPassword),
			Role:     seed.Role,
			Avatar:   seed.Avatar,
		}
		if createErr := tx.Create(&user).Error; createErr != nil {
			return fmt.Errorf("failed to create demo user %s: %w", email, createErr)
		}
	} else {
		updateFields := map[string]interface{}{
			"username": seed.Username,
			"password": string(hashedPassword),
			"role":     seed.Role,
			"avatar":   seed.Avatar,
		}
		if strings.TrimSpace(user.UUID) == "" {
			uuid, uuidErr := generateUUIDv4()
			if uuidErr != nil {
				return fmt.Errorf("failed to generate user uuid for %s: %w", email, uuidErr)
			}
			updateFields["uuid"] = uuid
		}
		if saveErr := tx.Model(&model.User{}).Where("id = ?", user.ID).Updates(updateFields).Error; saveErr != nil {
			return fmt.Errorf("failed to update demo user %s: %w", email, saveErr)
		}
	}

	switch seed.Role {
	case "enterprise":
		if err := upsertDemoEnterprise(tx, user.ID, seed); err != nil {
			return err
		}
	case "university":
		if err := upsertDemoUniversity(tx, user.ID, seed); err != nil {
			return err
		}
	}

	return nil
}

func upsertDemoEnterprise(tx *gorm.DB, userID uint, seed demoAccountSeed) error {
	var enterprise model.Enterprise
	err := tx.Where("user_id = ?", userID).First(&enterprise).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return fmt.Errorf("failed to query enterprise profile: %w", err)
		}
		entity := model.Enterprise{
			UserID:      userID,
			CompanyName: seed.CompanyName,
			AuditStatus: "approved",
		}
		if createErr := tx.Create(&entity).Error; createErr != nil {
			return fmt.Errorf("failed to create enterprise profile: %w", createErr)
		}
		return nil
	}

	updates := map[string]interface{}{
		"company_name": seed.CompanyName,
		"audit_status": "approved",
	}
	if saveErr := tx.Model(&model.Enterprise{}).Where("id = ?", enterprise.ID).Updates(updates).Error; saveErr != nil {
		return fmt.Errorf("failed to update enterprise profile: %w", saveErr)
	}
	return nil
}

func upsertDemoUniversity(tx *gorm.DB, userID uint, seed demoAccountSeed) error {
	var university model.University
	err := tx.Where("user_id = ?", userID).First(&university).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return fmt.Errorf("failed to query university profile: %w", err)
		}
		entity := model.University{
			UserID:         userID,
			UniversityName: seed.UniversityName,
			AuditStatus:    "approved",
		}
		if createErr := tx.Create(&entity).Error; createErr != nil {
			return fmt.Errorf("failed to create university profile: %w", createErr)
		}
		return nil
	}

	updates := map[string]interface{}{
		"university_name": seed.UniversityName,
		"audit_status":    "approved",
	}
	if saveErr := tx.Model(&model.University{}).Where("id = ?", university.ID).Updates(updates).Error; saveErr != nil {
		return fmt.Errorf("failed to update university profile: %w", saveErr)
	}
	return nil
}

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
