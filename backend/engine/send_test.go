package engine

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/StepOne-ai/GoStresser/models"
	"github.com/stretchr/testify/assert"
)

func TestSendRequest_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	results := make(chan LoadResult, 1)
	scenario := models.Scenario{
		Method:    "GET",
		TargetURL: server.URL,
	}

	SendRequest(scenario, results)
	result := <-results

	assert.NoError(t, result.Error)
	assert.Equal(t, http.StatusOK, result.Status)
	assert.Greater(t, result.Latency, 0*time.Nanosecond)
}

func TestSendRequest_Error(t *testing.T) {
	results := make(chan LoadResult, 1)
	scenario := models.Scenario{
		Method:    "GET",
		TargetURL: "http://nonexistent.local",
	}

	SendRequest(scenario, results)
	result := <-results

	assert.Error(t, result.Error)
	assert.Equal(t, 0, result.Status)
}