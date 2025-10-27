package models

import "time"

type Scenario struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"not null" json:"user_id"`
	Name      string    `gorm:"not null" json:"name"`
	TargetURL string    `gorm:"default:http://localhost:8080" json:"target_url"`
	Method    string    `gorm:"not null;default:GET" json:"method"`
	RPS       int       `gorm:"not null;default:100" json:"rps"`
	Duration  int       `gorm:"not null;default:10" json:"duration"`
	CreatedAt time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
}

func init() {
	Register(&Scenario{})
}