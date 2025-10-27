package models

import "time"

type TestRun struct {
    ID          uint      `gorm:"primaryKey" json:"id"`
    ScenarioID  uint      `gorm:"not null" json:"scenario_id"`
    UserID      string    `gorm:"not null" json:"user_id"`
    Status      string    `gorm:"not null;default:running" json:"status"`
    StartedAt   time.Time `gorm:"not null;autoCreateTime" json:"started_at"`
    EndedAt     time.Time `json:"ended_at"`
    Duration    int       `json:"duration"`
}

func init() {
	Register(&TestRun{})
}