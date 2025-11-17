package handlers

import (
	"net/http"

	db "github.com/StepOne-ai/GoStresser/database"
	"github.com/StepOne-ai/GoStresser/engine"
	"github.com/StepOne-ai/GoStresser/models"
	"github.com/gin-gonic/gin"
)

type CreateScenarioRequest struct {
    Name      string `json:"name" binding:"required"`
    TargetURL string `json:"target_url" binding:"required,url"`
    Method    string `json:"method" binding:"required,oneof=GET POST PUT DELETE"`
    RPS       int    `json:"rps" binding:"required,min=1,max=10000"`
    Duration  int    `json:"duration" binding:"required,min=1,max=3600"`
}

type UpdateScenarioRequest struct {
    Name      string `json:"name"`
    TargetURL string `json:"target_url" binding:"omitempty,url"`
    Method    string `json:"method" binding:"omitempty,oneof=GET POST PUT DELETE"`
    RPS       int    `json:"rps" binding:"omitempty,min=1,max=10000"`
    Duration  int    `json:"duration" binding:"omitempty,min=1,max=3600"`
}

// POST /api/v1/scenarios
func CreateScenarioHandler(c *gin.Context) {
    userID := c.GetString("userID")
    var req CreateScenarioRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    scenario := models.Scenario{
        UserID:    userID,
        Name:      req.Name,
        TargetURL: req.TargetURL,
        Method:    req.Method,
        RPS:       req.RPS,
        Duration:  req.Duration,
    }

    if err := db.DB.Create(&scenario).Error; err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "такой сценарий уже существует"})
        return
    }

    c.JSON(http.StatusCreated, scenario)
}

// GET /api/v1/scenarios
func GetScenariosHandler(c *gin.Context) {
    userID := c.GetString("userID")
    var scenarios []models.Scenario
    if err := db.DB.Where("user_id = ?", userID).Find(&scenarios).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch scenarios"})
        return
    }
    c.JSON(http.StatusOK, scenarios)
}

// GET /api/v1/scenarios/:id
func GetScenarioHandler(c *gin.Context) {
    userID := c.GetString("userID")
    id := c.Param("id")
    var scenario models.Scenario
    if err := db.DB.Where("id = ? AND user_id = ?", id, userID).First(&scenario).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Scenario not found"})
        return
    }
    c.JSON(http.StatusOK, scenario)
}

// PUT /api/v1/scenarios/:id
func UpdateScenarioHandler(c *gin.Context) {
    userID := c.GetString("userID")
    id := c.Param("id")

    var existing models.Scenario
    if err := db.DB.Where("id = ? AND user_id = ?", id, userID).First(&existing).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Scenario not found"})
        return
    }

    var req UpdateScenarioRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Only update fields that are provided
    if req.Name != "" {
        existing.Name = req.Name
    }
    if req.TargetURL != "" {
        existing.TargetURL = req.TargetURL
    }
    if req.Method != "" {
        existing.Method = req.Method
    }
    if req.RPS != 0 {
        existing.RPS = req.RPS
    }
    if req.Duration != 0 {
        existing.Duration = req.Duration
    }

    if err := db.DB.Save(&existing).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update scenario"})
        return
    }

    c.JSON(http.StatusOK, existing)
}

// DELETE /api/v1/scenarios/:id
func DeleteScenarioHandler(c *gin.Context) {
    userID := c.GetString("userID")
    id := c.Param("id")

    result := db.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Scenario{})
    if result.Error != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete scenario"})
        return
    }
    if result.RowsAffected == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Scenario not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Scenario deleted successfully"})
}

// POST /api/v1/scenarios/:id/run
func RunScenarioHandler(c *gin.Context) {
	userID := c.GetString("userID")
	scenarioIDStr := c.Param("id")

	var scenario models.Scenario
	if err := db.DB.Where("id = ? AND user_id = ?", scenarioIDStr, userID).First(&scenario).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Scenario not found"})
		return
	}

	// Create a new test run
	run := models.TestRun{
		ScenarioID: scenario.ID,
		UserID:     userID,
		Status:     "running",
		Duration:   scenario.Duration,
	}
	if err := db.DB.Create(&run).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create test run"})
		return
	}

	// Start the test in the background
	go engine.RunLoadTest(scenario, run.ID)

	// Respond immediately
	c.JSON(http.StatusAccepted, gin.H{
		"message": "Test run started",
		"run_id":  run.ID,
	})
}