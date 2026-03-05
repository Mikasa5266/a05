package service

import (
	"fmt"
	"strings"

	"your-project/model"
	"your-project/repository"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService() *UserService {
	return &UserService{
		userRepo: repository.NewUserRepository(),
	}
}

func CreateUser(username, email, password, role string) (*model.User, error) {
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

	user := &model.User{
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
		Role:     role,
	}

	if err := service.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
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
	question := &model.Question{
		Title:          title,
		Content:        content,
		Position:       position,
		Difficulty:     difficulty,
		Category:       category,
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
	return reportRepo.GetByUserID(userID, page, pageSize)
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
	reportRepo   *repository.ReportRepository
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
