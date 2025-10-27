package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	db "github.com/StepOne-ai/GoStresser/database"
	"github.com/StepOne-ai/GoStresser/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRouterWithRun() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.Use(func(c *gin.Context) {
		c.Set("userID", "test-user-id")
		c.Next()
	})

	api := r.Group("/api/v1")
	{
		api.POST("/scenarios", CreateScenarioHandler)
		api.POST("/scenarios/:id/run", RunScenarioHandler) // ← added
		api.GET("/runs/:id/report", func(c *gin.Context) {
			var result models.TestResult
			id := c.Param("id")
			if err := db.DB.Where("run_id = ?", id).First(&result).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
				return
			}
			c.JSON(http.StatusOK, result)
		})
	}
	return r
}

func TestFullScenarioLifecycle(t *testing.T) {
	db.DB = db.SetupTestDB()
	defer db.DB.Exec("DROP TABLE IF EXISTS scenarios, test_runs, test_metrics, test_results")

	router := setupRouterWithRun()

	// 1. Create scenario
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/scenarios",
		strings.NewReader(`{
			"name": "Full Lifecycle Test",
			"target_url": "http://httpbin.org/get",
			"method": "GET",
			"rps": 5,
			"duration": 2
		}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var scenario models.Scenario
	err := json.Unmarshal(w.Body.Bytes(), &scenario)
	require.NoError(t, err)

	// 2. Run scenario
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", fmt.Sprintf("/api/v1/scenarios/%d/run", scenario.ID), nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusAccepted, w.Code)

	// 3. Wait for completion (duration + buffer)
	time.Sleep(3 * time.Second)

	// 4. Fetch report
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/run/%d/report", 1), nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var report models.TestResult
	err = json.Unmarshal(w.Body.Bytes(), &report)
	require.NoError(t, err)

	assert.Equal(t, uint(1), report.RunID)
	assert.Equal(t, 10, report.TotalRequests) // 5 RPS * 2 sec
	assert.GreaterOrEqual(t, report.AvgLatency, 0.0)
	assert.GreaterOrEqual(t, report.ErrorRate, 0.0)
}