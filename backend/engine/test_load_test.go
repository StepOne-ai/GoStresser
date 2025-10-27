package engine

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	db "github.com/StepOne-ai/GoStresser/database"
	"github.com/StepOne-ai/GoStresser/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunLoadTest(t *testing.T) {
	// Setup test DB
	db.DB = db.SetupTestDB()
	defer db.DB.Exec("DROP TABLE IF EXISTS scenarios, test_runs, test_metrics, test_results")

	// Mock target server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond) // simulate latency
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create scenario
	scenario := models.Scenario{
		UserID:    "test-user",
		Name:      "Test Run",
		TargetURL: server.URL,
		Method:    "GET",
		RPS:       5,
		Duration:  2, // 2 seconds → 10 total requests
	}
	db.DB.Create(&scenario)

	// Create test run
	run := models.TestRun{
		ScenarioID: scenario.ID,
		UserID:     scenario.UserID,
		Status:     "running",
	}
	db.DB.Create(&run)

	// Run the load test
	RunLoadTest(scenario, run.ID)

	// Wait a bit more to ensure cleanup
	time.Sleep(100 * time.Millisecond)

	// Verify TestRun is completed
	var updatedRun models.TestRun
	err := db.DB.First(&updatedRun, run.ID).Error
	require.NoError(t, err)
	assert.Equal(t, "completed", updatedRun.Status)
	assert.Equal(t, 2, updatedRun.Duration) // should match scenario duration

	// Verify metrics: 2 entries (1 per second)
	var metrics []models.TestMetric
	err = db.DB.Where("run_id = ?", run.ID).Find(&metrics).Error
	require.NoError(t, err)
	assert.Len(t, metrics, 2)

	for _, m := range metrics {
		assert.Equal(t, uint(run.ID), m.RunID)
		assert.Equal(t, 5, m.RPS)            // 5 RPS per second
		assert.Greater(t, m.AvgLatency, 0.0) // latency recorded
		assert.GreaterOrEqual(t, m.ErrorCount, 0)
	}

	// Verify final report
	var report models.TestResult
	err = db.DB.Where("run_id = ?", run.ID).First(&report).Error
	require.NoError(t, err)
	assert.Equal(t, uint(run.ID), report.RunID)
	assert.Equal(t, 10, report.TotalRequests) // 5 RPS * 2 sec
	assert.GreaterOrEqual(t, report.ErrorRate, 0.0)
	assert.Greater(t, report.AvgLatency, 0.0)
}