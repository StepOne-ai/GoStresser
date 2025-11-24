package engine

import (
	"testing"

	db "github.com/StepOne-ai/GoStresser/database"
	"github.com/StepOne-ai/GoStresser/models"
	"github.com/stretchr/testify/assert"
)

func TestGenerateReport(t *testing.T) {
	db.DB = db.SetupTestDB()
	defer db.DB.Exec("DROP TABLE IF EXISTS test_runs, test_metrics, test_results")

	// Create a run
	run := models.TestRun{ScenarioID: 1, UserID: "test", Status: "running"}
	db.DB.Create(&run)

	// Insert metrics
	metrics := []models.TestMetric{
		{RunID: run.ID, RPS: 10, AvgLatency: 100, ErrorCount: 1},
		{RunID: run.ID, RPS: 10, AvgLatency: 120, ErrorCount: 0},
	}
	for _, m := range metrics {
		db.DB.Create(&m)
	}

	GenerateReport(run.ID, "test")

	var result models.TestResult
	db.DB.First(&result, "run_id = ?", run.ID)

	assert.Equal(t, uint(run.ID), result.RunID)
	assert.Equal(t, 20, result.TotalRequests)
	assert.Equal(t, 5.0, result.ErrorRate) // 1/20 = 5%
	assert.Greater(t, result.AvgLatency, 0.0)
}