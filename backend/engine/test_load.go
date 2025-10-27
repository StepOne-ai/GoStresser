package engine

import (
	"sync"
	"time"

	db "github.com/StepOne-ai/GoStresser/database"
	"github.com/StepOne-ai/GoStresser/models"
)

func RunLoadTest(scenario models.Scenario, runID uint) {
	duration := time.Duration(scenario.Duration) * time.Second
	totalRequests := scenario.RPS * scenario.Duration

	results := make(chan LoadResult, totalRequests)
	var wg sync.WaitGroup

	// Start request workers
	wg.Add(totalRequests)
	for i := 0; i < totalRequests; i++ {
		go func() {
			defer wg.Done()
			SendRequest(scenario, results)
		}()
		time.Sleep(time.Second / time.Duration(scenario.RPS))
	}

	// Close results channel once all workers are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Metric collection per second
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	second := 0
	remaining := totalRequests

	for {
		select {
		case <-ticker.C:
			second++
			// Collect up to RPS results for this second
			batchSize := scenario.RPS
			if remaining < batchSize {
				batchSize = remaining
			}
			remaining -= batchSize

			batch := make([]LoadResult, 0, batchSize)
			for i := 0; i < batchSize; i++ {
				if result, ok := <-results; ok {
					batch = append(batch, result)
				} else {
					// Channel closed early
					break
				}
			}

			metric := aggregateBatch(batch)
			SaveMetrics(runID, second, metric)

			if second >= scenario.Duration || remaining <= 0 {
				goto endTest
			}

		case <-time.After(duration + 2*time.Second): // grace period
			goto endTest
		}
	}

endTest:
	now := time.Now()
	db.DB.Model(&models.TestRun{}).Where("id = ?", runID).Updates(map[string]interface{}{
		"status":   "completed",
		"ended_at": now,
		"duration": second,
	})

	GenerateReport(runID)
}