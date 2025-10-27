// internal/handlers/scenario_test.go
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/StepOne-ai/GoStresser/database"
	"github.com/StepOne-ai/GoStresser/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRouter() *gin.Engine {
    r := gin.New()
    r.Use(gin.Logger(), gin.Recovery())

    // Mock auth middleware
    r.Use(func(c *gin.Context) {
        c.Set("userID", "test-user-id")
        c.Next()
    })

    // Register routes — NO DB passed to handlers
    api := r.Group("/api/v1")
    {
        api.POST("/scenarios", CreateScenarioHandler)
        api.GET("/scenarios", GetScenariosHandler)
        api.GET("/scenarios/:id", GetScenarioHandler)
        api.PUT("/scenarios/:id", UpdateScenarioHandler)
        api.DELETE("/scenarios/:id", DeleteScenarioHandler)
    }

    return r
}

func TestCreateScenario(t *testing.T) {
    database.DB = database.SetupTestDB()
    defer database.DB.Exec("DROP TABLE IF EXISTS scenarios, users")

    router := setupRouter()

    w := httptest.NewRecorder()
    req, _ := http.NewRequest("POST", "/api/v1/scenarios", 
        strings.NewReader(`{
            "name": "Test Scenario",
            "target_url": "http://localhost:3000",
            "method": "GET",
            "rps": 100,
            "duration": 10
        }`))
    req.Header.Set("Content-Type", "application/json")

    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusCreated, w.Code)

    var created models.Scenario
    err := json.Unmarshal(w.Body.Bytes(), &created)
    require.NoError(t, err)
    assert.Equal(t, "Test Scenario", created.Name)
    assert.Equal(t, "test-user-id", created.UserID)
}

func TestGetScenarios(t *testing.T) {
    database.DB = database.SetupTestDB() // ← SET GLOBAL DB
    defer database.DB.Exec("DROP TABLE IF EXISTS scenarios, users")

    router := setupRouter()

    // Create test scenario
    database.DB.Create(&models.Scenario{
        UserID:    "test-user-id",
        Name:      "Test Scenario",
        TargetURL: "http://localhost:3000",
        Method:    "GET",
        RPS:       100,
        Duration:  10,
    })

    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/api/v1/scenarios", nil)

    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)

    var scenarios []models.Scenario
    err := json.Unmarshal(w.Body.Bytes(), &scenarios)
    require.NoError(t, err)
    assert.Len(t, scenarios, 1)
    assert.Equal(t, "Test Scenario", scenarios[0].Name)
}

func TestGetScenario(t *testing.T) {
    database.DB = database.SetupTestDB() // ← SET GLOBAL DB
    defer database.DB.Exec("DROP TABLE IF EXISTS scenarios, users")

    router := setupRouter()

    scenario := models.Scenario{
        UserID:    "test-user-id",
        Name:      "Test Scenario",
        TargetURL: "http://localhost:3000",
        Method:    "GET",
        RPS:       100,
        Duration:  10,
    }
    database.DB.Create(&scenario)

    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/api/v1/scenarios/"+fmt.Sprint(scenario.ID), nil)

    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)

    var fetched models.Scenario
    err := json.Unmarshal(w.Body.Bytes(), &fetched)
    require.NoError(t, err)
    assert.Equal(t, scenario.Name, fetched.Name)
}

func TestUpdateScenario(t *testing.T) {
    database.DB = database.SetupTestDB() // ← SET GLOBAL DB
    defer database.DB.Exec("DROP TABLE IF EXISTS scenarios, users")

    router := setupRouter()

    scenario := models.Scenario{
        UserID:    "test-user-id",
        Name:      "Old Name",
        TargetURL: "http://localhost:3000",
        Method:    "GET",
        RPS:       100,
        Duration:  10,
    }
    database.DB.Create(&scenario)

    w := httptest.NewRecorder()
    req, _ := http.NewRequest("PUT", "/api/v1/scenarios/"+fmt.Sprint(scenario.ID), 
        strings.NewReader(`{"name": "New Name"}`))
    req.Header.Set("Content-Type", "application/json")

    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)

    var updated models.Scenario
    err := json.Unmarshal(w.Body.Bytes(), &updated)
    require.NoError(t, err)
    assert.Equal(t, "New Name", updated.Name)
}

func TestDeleteScenario(t *testing.T) {
    database.DB = database.SetupTestDB() // ← SET GLOBAL DB
    defer database.DB.Exec("DROP TABLE IF EXISTS scenarios, users")

    router := setupRouter()

    scenario := models.Scenario{
        UserID:    "test-user-id",
        Name:      "To Delete",
        TargetURL: "http://localhost:3000",
        Method:    "GET",
        RPS:       100,
        Duration:  10,
    }
    database.DB.Create(&scenario)

    w := httptest.NewRecorder()
    req, _ := http.NewRequest("DELETE", "/api/v1/scenarios/"+fmt.Sprint(scenario.ID), nil)

    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)

    var count int64
    database.DB.Model(&models.Scenario{}).Where("id = ?", scenario.ID).Count(&count)
    assert.Equal(t, int64(0), count)
}

func TestScenarioLifecycle(t *testing.T) {
    // Setup the test database
    database.DB = database.SetupTestDB()
    defer database.DB.Exec("DROP TABLE IF EXISTS scenarios, users, results")

    // Setup the router with mocked auth middleware
    router := setupRouter()

    // Step 1: Create a new scenario
    w := httptest.NewRecorder()
    createReq, _ := http.NewRequest("POST", "/api/v1/scenarios",
        strings.NewReader(`{
            "name": "Lifecycle Test Scenario",
            "target_url": "http://localhost:3000",
            "method": "GET",
            "rps": 10,
            "duration": 5
        }`))
    createReq.Header.Set("Content-Type", "application/json")
    router.ServeHTTP(w, createReq)

    assert.Equal(t, http.StatusCreated, w.Code)

    var createdScenario models.Scenario
    err := json.Unmarshal(w.Body.Bytes(), &createdScenario)
    require.NoError(t, err)
    assert.Equal(t, "Lifecycle Test Scenario", createdScenario.Name)

    scenarioID := createdScenario.ID

    // Step 2: Run the scenario
    w = httptest.NewRecorder()
    runReq, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/scenarios/%d/run", scenarioID), nil)
    router.ServeHTTP(w, runReq)

    assert.Equal(t, http.StatusOK, w.Code)

    var runResponse map[string]any
    err = json.Unmarshal(w.Body.Bytes(), &runResponse)
    require.NoError(t, err)
    assert.Contains(t, runResponse, "message")
    assert.Equal(t, "Scenario execution started", runResponse["message"])

    // Simulate waiting for the scenario to complete (adjust sleep time based on actual implementation)
    // In a real-world app, you might poll the results endpoint instead of sleeping.
    time.Sleep(6 * time.Second)

    // Step 3: Fetch results
    w = httptest.NewRecorder()
    resultsReq, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/scenarios/%d/results", scenarioID), nil)
    router.ServeHTTP(w, resultsReq)

    assert.Equal(t, http.StatusOK, w.Code)

    var results []models.TestResult
    err = json.Unmarshal(w.Body.Bytes(), &results)
    require.NoError(t, err)
    assert.Greater(t, len(results), 0, "Expected at least one result from the scenario execution")

    // Step 4: Generate a report
    w = httptest.NewRecorder()
    reportReq, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/scenarios/%d/report", scenarioID), nil)
    router.ServeHTTP(w, reportReq)

    assert.Equal(t, http.StatusOK, w.Code)

    var report map[string]interface{}
    err = json.Unmarshal(w.Body.Bytes(), &report)
    require.NoError(t, err)
    assert.Contains(t, report, "summary")
    assert.Contains(t, report, "details")
    assert.NotEmpty(t, report["summary"], "Report summary should not be empty")
    assert.NotEmpty(t, report["details"], "Report details should not be empty")
}