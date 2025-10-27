package models

import (
    "time"
    "gorm.io/gorm"
	utils "github.com/StepOne-ai/GoStresser/utils"
)

type User struct {
    ID        string    `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
    Email     string    `gorm:"uniqueIndex;not null" json:"email"`
    Password  string    `gorm:"not null" json:"-"` // never expose
    CreatedAt time.Time `json:"created_at"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
    if u.Password != "" {
        hashed, err := utils.HashPassword(u.Password)
        if err != nil {
            return err
        }
        u.Password = hashed
    }
    return nil
}

func init() {
    Register(&User{})
}