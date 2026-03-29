package main

import (
	"fmt"
	"log"

	"your-project/config"
	"your-project/model"
	"your-project/repository"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// Load Config
	if err := config.LoadConfig("config.yaml"); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Init DB connection manually since we are outside main app
	db, err := initDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	repository.SetDB(db)

	createAccount(db, "enterprise_user", "enterprise@test.com", "123456", "enterprise")
	createAccount(db, "university_user", "university@test.com", "123456", "university")
}

func initDatabase() (*gorm.DB, error) {
	cfg := config.GetConfig()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Database,
	)
	return gorm.Open(mysql.Open(dsn), &gorm.Config{})
}

func createAccount(db *gorm.DB, username, email, password, role string) {
	// Check if user exists
	var existing model.User
	if err := db.Where("email = ?", email).First(&existing).Error; err == nil {
		fmt.Printf("User %s already exists\n", email)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	user := model.User{
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
		Role:     role,
	}

	if err := db.Create(&user).Error; err != nil {
		log.Fatalf("Failed to create user %s: %v", username, err)
	}

	fmt.Printf("Created user: %s (Role: %s, Password: %s)\n", email, role, password)
}
