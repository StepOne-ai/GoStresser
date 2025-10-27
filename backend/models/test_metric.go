package models

type TestMetric struct {
    ID         uint    `gorm:"primaryKey" json:"id"`
    RunID      uint    `gorm:"not null" json:"run_id"`
    Timestamp  int     `gorm:"not null" json:"timestamp"` // seconds since test start
    RPS        int     `gorm:"not null" json:"rps"`
    AvgLatency float64 `gorm:"not null" json:"avg_latency"` // ms
    MaxLatency float64 `gorm:"not null" json:"max_latency"` // ms
    MinLatency float64 `gorm:"not null" json:"min_latency"` // ms
    ErrorCount int     `gorm:"not null" json:"error_count"`
    CPU        float64 `json:"cpu"`
    RAM        float64 `json:"ram"`
}

func init() {
	Register(&TestMetric{})
}