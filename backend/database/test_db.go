package database

import (
	"github.com/StepOne-ai/GoStresser/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func SetupTestDB() *gorm.DB {
	dsn := "host=localhost user=postgres password=postgres dbname=go_stresser_test port=5432 sslmode=disable"
    testDB, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	AutoMigrate(testDB)
    return testDB
}

func CreateTestUser(db *gorm.DB, email, password string) *models.User {
    user := &models.User{
        Email:    email,
        Password: password,
    }
    db.Create(user)
    return user
}