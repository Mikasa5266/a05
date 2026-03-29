package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"your-project/model"
	"your-project/repository"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService() *UserService {
	return &UserService{
		userRepo: repository.NewUserRepository(),
	}
}

func createUserWithRole(username, email, password, role string) (*model.User, error) {
	service := NewUserService()

	existingUser, _ := service.userRepo.GetByEmail(email)
	if existingUser != nil {
		return nil, fmt.Errorf("email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	if role == "" {
		role = "student"
	}

	uuid, err := generateUUIDv4()
	if err != nil {
		return nil, fmt.Errorf("failed to generate user uuid: %w", err)
	}

	user := &model.User{
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
		Role:     role,
		UUID:     uuid,
	}

	if err := service.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func createUserWithRoleTx(tx *gorm.DB, username, email, password, role string) (*model.User, error) {
	var existing model.User
	err := tx.Where("email = ?", email).First(&existing).Error
	if err == nil {
		return nil, fmt.Errorf("email already exists")
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	if role == "" {
		role = "student"
	}

	uuid, err := generateUUIDv4()
	if err != nil {
		return nil, fmt.Errorf("failed to generate user uuid: %w", err)
	}

	user := &model.User{
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
		Role:     role,
		UUID:     uuid,
	}

	if err := tx.Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func CreateStudentUser(username, email, password string) (*model.User, error) {
	return createUserWithRole(username, email, password, "student")
}

// CreateUser is kept for backward compatibility and now only allows student direct registration.
func CreateUser(username, email, password, role string) (*model.User, error) {
	if role != "" && role != "student" {
		return nil, fmt.Errorf("enterprise/university must use application endpoint")
	}
	return CreateStudentUser(username, email, password)
}

func ApplyEnterprise(username, email, password, companyName, contactName, contactPhone, businessScope string) (*model.User, *model.Enterprise, error) {
	db := repository.GetDB()
	if strings.TrimSpace(companyName) == "" {
		return nil, nil, fmt.Errorf("company_name is required")
	}

	var createdUser *model.User
	var createdEnterprise *model.Enterprise
	if err := db.Transaction(func(tx *gorm.DB) error {
		user, err := createUserWithRoleTx(tx, username, email, password, "enterprise")
		if err != nil {
			return err
		}
		createdUser = user

		enterprise := &model.Enterprise{
			UserID:        user.ID,
			CompanyName:   companyName,
			ContactName:   contactName,
			ContactPhone:  contactPhone,
			BusinessScope: businessScope,
			AuditStatus:   "pending",
		}
		if err := tx.Create(enterprise).Error; err != nil {
			return err
		}
		createdEnterprise = enterprise
		return nil
	}); err != nil {
		return nil, nil, err
	}

	return createdUser, createdEnterprise, nil
}

func ApplyUniversity(username, email, password, universityName, contactName, contactPhone, department string) (*model.User, *model.University, error) {
	db := repository.GetDB()
	if strings.TrimSpace(universityName) == "" {
		return nil, nil, fmt.Errorf("university_name is required")
	}

	var createdUser *model.User
	var createdUniversity *model.University
	if err := db.Transaction(func(tx *gorm.DB) error {
		user, err := createUserWithRoleTx(tx, username, email, password, "university")
		if err != nil {
			return err
		}
		createdUser = user

		university := &model.University{
			UserID:         user.ID,
			UniversityName: universityName,
			ContactName:    contactName,
			ContactPhone:   contactPhone,
			Department:     department,
			AuditStatus:    "pending",
		}
		if err := tx.Create(university).Error; err != nil {
			return err
		}
		createdUniversity = university
		return nil
	}); err != nil {
		return nil, nil, err
	}

	return createdUser, createdUniversity, nil
}

func GetAuditStatusForUser(user *model.User) (string, string, error) {
	if user == nil {
		return "", "", fmt.Errorf("user is nil")
	}

	switch user.Role {
	case "enterprise":
		var enterprise model.Enterprise
		if err := repository.GetDB().Where("user_id = ?", user.ID).First(&enterprise).Error; err != nil {
			// Backward compatibility: legacy enterprise users may not have onboarding rows.
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "approved", "", nil
			}
			return "", "", err
		}
		return enterprise.AuditStatus, enterprise.AuditRemark, nil
	case "university":
		var university model.University
		if err := repository.GetDB().Where("user_id = ?", user.ID).First(&university).Error; err != nil {
			// Backward compatibility: legacy university users may not have onboarding rows.
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "approved", "", nil
			}
			return "", "", err
		}
		return university.AuditStatus, university.AuditRemark, nil
	default:
		return "approved", "", nil
	}
}

func AuditApplication(role string, applicationID uint, status, remark string, adminID uint) error {
	status = strings.TrimSpace(strings.ToLower(status))
	if status != "approved" && status != "rejected" {
		return fmt.Errorf("status must be approved or rejected")
	}

	now := time.Now()
	switch role {
	case "enterprise":
		var enterprise model.Enterprise
		if err := repository.GetDB().First(&enterprise, applicationID).Error; err != nil {
			return err
		}
		enterprise.AuditStatus = status
		enterprise.AuditRemark = strings.TrimSpace(remark)
		enterprise.AuditedBy = &adminID
		enterprise.AuditedAt = &now
		return repository.GetDB().Save(&enterprise).Error
	case "university":
		var university model.University
		if err := repository.GetDB().First(&university, applicationID).Error; err != nil {
			return err
		}
		university.AuditStatus = status
		university.AuditRemark = strings.TrimSpace(remark)
		university.AuditedBy = &adminID
		university.AuditedAt = &now
		return repository.GetDB().Save(&university).Error
	default:
		return fmt.Errorf("unsupported role")
	}
}

func AuthenticateUser(email, password string) (*model.User, error) {
	service := NewUserService()

	user, err := service.userRepo.GetByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if _, err := ensureUserUUID(service.userRepo, user); err != nil {
		return nil, fmt.Errorf("failed to ensure user identity: %w", err)
	}

	return user, nil
}

func GetUserByID(userID uint) (*model.User, error) {
	service := NewUserService()
	return service.userRepo.GetByID(userID)
}

func UpdateUserProfile(userID uint, username, email string) (*model.User, error) {
	service := NewUserService()

	user, err := service.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	if username != "" {
		user.Username = username
	}
	if email != "" {
		existingUser, _ := service.userRepo.GetByEmail(email)
		if existingUser != nil && existingUser.ID != userID {
			return nil, fmt.Errorf("email already exists")
		}
		user.Email = email
	}

	if err := service.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

func UpdateUserAvatar(userID uint, avatarURL string) (*model.User, error) {
	service := NewUserService()
	user, err := service.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	user.Avatar = avatarURL // Assuming model.User has Avatar field? Check model/user.go
	if err := service.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update avatar: %w", err)
	}
	return user, nil
}

func UpdateUserPassword(userID uint, oldPassword, newPassword string) error {
	service := NewUserService()
	user, err := service.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return fmt.Errorf("invalid old password")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	user.Password = string(hashedPassword)
	if err := service.userRepo.Update(user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	return nil
}

func GetQuestions(position, difficulty, category string) ([]*model.Question, error) {
	ensureRepos()
	return questionRepo.GetQuestions(position, difficulty, category)
}

func GetQuestionByID(questionID uint) (*model.Question, error) {
	ensureRepos()
	return questionRepo.GetByID(questionID)
}

func CreateQuestion(title, content, position, difficulty, category string, tags []string, expectedAnswer string) (*model.Question, error) {
	ensureRepos()
	validator := MustGetAIService()
	candidate := &model.Question{
		Title:    strings.TrimSpace(title),
		Content:  strings.TrimSpace(content),
		Category: strings.TrimSpace(category),
	}
	if validator.IsContextDependentOpeningQuestion(candidate) {
		return nil, fmt.Errorf("question looks like follow-up/context-dependent and cannot be added to official bank")
	}

	question := &model.Question{
		Title:          title,
		Content:        content,
		Position:       position,
		Difficulty:     difficulty,
		Category:       category,
		Source:         "standard",
		RAGEligible:    true,
		Tags:           strings.Join(tags, ","),
		ExpectedAnswer: expectedAnswer,
	}

	if err := questionRepo.Create(question); err != nil {
		return nil, fmt.Errorf("failed to create question: %w", err)
	}

	return question, nil
}

func GetUserReports(userID uint, page, pageSize int) ([]*model.Report, int64, error) {
	ensureRepos()
	return reportRepo.ListByUserPaged(userID, page, pageSize)
}

func GetReportByID(userID, reportID uint) (*model.Report, error) {
	ensureRepos()
	report, err := reportRepo.GetByID(reportID)
	if err != nil {
		return nil, err
	}

	if report.UserID != userID {
		return nil, fmt.Errorf("unauthorized access")
	}

	return report, nil
}

func GenerateInterviewReport(userID, interviewID uint) (*model.Report, error) {
	reportService := NewReportService()
	return reportService.GenerateInterviewReport(userID, interviewID)
}

var (
	userRepo     *repository.UserRepository
	questionRepo *repository.QuestionRepository
	reportRepo   repository.ReportRepository
)

func initRepos() {
	userRepo = repository.NewUserRepository()
	questionRepo = repository.NewQuestionRepository()
	reportRepo = repository.NewReportRepository()
}

func ensureRepos() {
	if userRepo == nil || questionRepo == nil || reportRepo == nil {
		initRepos()
	}
}
