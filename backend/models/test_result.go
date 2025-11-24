package models

type TestResult struct {
    ID            uint    `gorm:"primaryKey" json:"id"`
    UserID        string  `gorm:"not null" json:"user_id"`
    RunID         uint    `gorm:"not null;uniqueIndex" json:"run_id"`
    TotalRequests int     `gorm:"not null" json:"total_requests"`
    AvgLatency    float64 `gorm:"not null" json:"avg_latency"`
    P95Latency    float64 `gorm:"not null" json:"p95_latency"`
    P99Latency    float64 `gorm:"not null" json:"p99_latency"`
    ErrorRate     float64 `gorm:"not null" json:"error_rate"`
    PeakCPU       float64 `json:"peak_cpu"`
    PeakRAM       float64 `json:"peak_ram"`
    Recommendations string `json:"recommendations"`
}

func init() {
	Register(&TestResult{})
}