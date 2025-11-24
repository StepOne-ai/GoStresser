package handlers

import (
	"strconv"
	"time"

	db "github.com/StepOne-ai/GoStresser/database"
	engine "github.com/StepOne-ai/GoStresser/engine"
	"github.com/StepOne-ai/GoStresser/models"
	"github.com/gin-gonic/gin"
)


func StartTestHandler(c *gin.Context) {
    userID := c.GetString("userID")
    scenarioID := c.Param("id")

    var scenario models.Scenario
    if err := db.DB.Where("id = ? AND user_id = ?", scenarioID, userID).First(&scenario).Error; err != nil {
        c.JSON(404, gin.H{"error": "Scenario not found"})
        return
    }

    testRun := models.TestRun{
        ScenarioID: scenario.ID,
        UserID:     userID,
        Status:     "running",
    }
    db.DB.Create(&testRun)

    go engine.RunLoadTest(scenario, testRun.ID, userID)

    c.JSON(200, gin.H{"run_id": testRun.ID})
}

func GetReportHandler(c *gin.Context) {
	runIDParam := c.Param("run_id")
	
	// Convert the run_id parameter from string to uint
	runID, err := strconv.ParseUint(runIDParam, 10, 32) // Assuming uint is 32-bit, use 64 if needed
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid run ID"})
		return
	}

	var report models.TestResult
	err = db.DB.Where("id = ?", uint(runID)).First(&report).Error
	if err != nil {
		// If the report was not found after generation, it might mean no metrics were available
		// Return a 404 or a 200 with an empty/default object depending on your API design
		// Here, we return a 404 as the report could not be generated or found.
		c.JSON(404, gin.H{"error": "Report not found"})
		return
	}

	c.JSON(200, report)
}

func GetAllReportsHandler(c *gin.Context) {
	userID := c.GetString("userID")

	// Struct for final response
	type ScenarioInfo struct {
		ID        uint   `json:"id"`
		Name      string `json:"name"`
		CreatedAt string `json:"created_at"` // formatted as string
	}

	var rawScenarios []struct {
		ID        uint      `gorm:"column:id"`
		Name      string    `gorm:"column:name"`
		CreatedAt time.Time `gorm:"column:created_at"`
	}

	err := db.DB.Table("test_results").
		Select("DISTINCT test_results.id, scenarios.name, scenarios.created_at").
		Joins("JOIN test_runs ON test_runs.id = test_results.run_id").
		Joins("JOIN scenarios ON scenarios.id = test_runs.scenario_id").
		Where("test_results.user_id = ?", userID).
		Find(&rawScenarios).Error

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch scenarios"})
		return
	}

	// Format timestamps (e.g., "2025-11-25 10:45:32")
	scenarios := make([]ScenarioInfo, len(rawScenarios))
	for i, s := range rawScenarios {
		// Use UTC or local time as needed — here using local for readability
		// If your DB stores UTC and you want to keep it in UTC, use s.CreatedAt.UTC()
		scenarios[i] = ScenarioInfo{
			ID:        s.ID,
			Name:      s.Name,
			CreatedAt: s.CreatedAt.Format("2006-01-02 15:04:05"), // ISO-like but without T/Z
		}
	}

	c.JSON(200, scenarios)
}

func DeleteReportHandler(c *gin.Context) {
	reportID := c.Param("report_id")
	var report models.TestResult
	if err := db.DB.Where("id = ?", reportID).First(&report).Error; err != nil {
		c.JSON(404, gin.H{"error": "Report not found"})
		return
	}
	db.DB.Delete(&report)
	c.JSON(200, gin.H{"message": "Report deleted successfully"})
}

func GetLiveMetricsHandler(c *gin.Context) {
    runID := c.Param("run_id")
    var metrics []models.TestMetric
    db.DB.Where("run_id = ?", runID).Order("timestamp desc").Limit(1).Find(&metrics)

    if len(metrics) == 0 {
        c.JSON(200, gin.H{"error": "No metrics yet"})
        return
    }

    c.JSON(200, metrics[0])
}