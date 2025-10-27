package models

import (
    "time"
    "gorm.io/gorm"
	"github.com/StepOne-ai/GoStresser/utils"
)

type TestUser struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    Email     string    `gorm:"uniqueIndex;not null" json:"email"`
    Password  string    `gorm:"not null" json:"-"` 
    CreatedAt time.Time `json:"created_at"`
}

func (u *TestUser) BeforeCreate(tx *gorm.DB) error {
    if u.Password != "" {
        hashed, err := utils.HashPassword(u.Password)
        if err != nil {
            return err
        }
        u.Password = hashed
    }
    return nil
}