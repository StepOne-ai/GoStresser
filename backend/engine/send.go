package engine

import (
	"net/http"
	"time"
	"github.com/StepOne-ai/GoStresser/models"
)

func SendRequest(scenario models.Scenario, results chan<- LoadResult) {
	client := &http.Client{Timeout: 10 * time.Second} // more realistic timeout
	req, err := http.NewRequest(scenario.Method, scenario.TargetURL, nil)
	if err != nil {
		results <- LoadResult{Error: err}
		return
	}

	start := time.Now()
	resp, err := client.Do(req)
	latency := time.Since(start)

	var status int
	if resp != nil {
		status = resp.StatusCode
		resp.Body.Close()
	}

	results <- LoadResult{
		Latency: latency,
		Error:   err,
		Status:  status,
	}
}