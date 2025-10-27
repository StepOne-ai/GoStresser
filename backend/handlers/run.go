package handlers

import (
	"github.com/StepOne-ai/GoStresser/models"
	engine "github.com/StepOne-ai/GoStresser/engine"
	db "github.com/StepOne-ai/GoStresser/database"
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

    go engine.RunLoadTest(scenario, testRun.ID)

    c.JSON(200, gin.H{"run_id": testRun.ID})
}

func GetReportHandler(c *gin.Context) {
    runID := c.Param("run_id")
    var metrics []models.TestMetric
    db.DB.Where("run_id = ?", runID).Find(&metrics)
    
    c.JSON(200, metrics)
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