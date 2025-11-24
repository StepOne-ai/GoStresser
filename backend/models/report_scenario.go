package models

type ReportScenario struct {
	ID        string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	ScenarioID string `json:"scenario_id"`
	ReportID  string `json:"report_id"`
}

func init() {
	Register(&ReportScenario{})
}