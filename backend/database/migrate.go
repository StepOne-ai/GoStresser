package database

import (
    "github.com/StepOne-ai/GoStresser/models"
    "gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
    return db.AutoMigrate(models.GetRegisteredModels()...)
}