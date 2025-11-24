package engine

import (
	"context"
	"errors"
	"sync"
	"time"

	db "github.com/StepOne-ai/GoStresser/database"
	"github.com/StepOne-ai/GoStresser/models"
)

func RunLoadTest(scenario models.Scenario, runID uint, userID string) {
	// Validate inputs
	if scenario.RPS <= 0 {
		scenario.RPS = 1
	}
	if scenario.Duration <= 0 {
		scenario.Duration = 1
	}

	duration := time.Duration(scenario.Duration) * time.Second
	rps := scenario.RPS
	totalRequests := rps * scenario.Duration

	// Use context for coordinated cancellation
	ctx, cancel := context.WithTimeout(context.Background(), duration+10*time.Second)
	defer cancel()

	// Create channel with reasonable buffer
	results := make(chan LoadResult, rps*2) // Buffer for 2 seconds of results
	var wg sync.WaitGroup

	// Flag to ensure channel is closed only once
	var closeOnce sync.Once

	// Request generator goroutine
	go func() {
		defer func() {
			wg.Wait() // Wait for all requests to finish
			closeOnce.Do(func() {
				close(results) // Safe close - only once
			})
		}()

		startTime := time.Now()
		endTime := startTime.Add(duration)
		requestsSent := 0

		for requestsSent < totalRequests && time.Now().Before(endTime) {
			// Calculate how many requests should be sent by now
			elapsed := time.Since(startTime).Seconds()
			targetSent := int(elapsed * float64(rps))
			
			// Send requests in batches to avoid overwhelming
			batchSize := targetSent - requestsSent
			if batchSize > 100 { // Limit batch size
				batchSize = 100
			}
			if batchSize <= 0 {
				batchSize = 1
			}

			for i := 0; i < batchSize && requestsSent < totalRequests; i++ {
				select {
				case <-ctx.Done():
					return
				default:
					wg.Add(1)
					go func() {
						defer wg.Done()
						defer func() {
							// Recover from any panics in SendRequest
							if r := recover(); r != nil {
								// Log panic but don't crash the whole test
								// You might want to send an error result instead
								result := LoadResult{
									Error:   errors.New("panic in SendRequest"),
								}
								select {
								case results <- result:
								case <-ctx.Done():
								}
							}
						}()
						SendRequest(scenario, results)
					}()
					requestsSent++
				}
			}

			// Small sleep to prevent CPU spinning
			time.Sleep(10 * time.Millisecond)
		}
	}()

	// Metrics collection
	second := 0
	metricsBuffer := make([]LoadResult, 0, rps)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case result, ok := <-results:
			if !ok {
				// Channel closed normally - all requests completed
				if len(metricsBuffer) > 0 {
					second++
					metric := aggregateBatch(metricsBuffer)
					SaveMetrics(runID, second, metric)
				}
				goto cleanup
			}
			metricsBuffer = append(metricsBuffer, result)

			// Auto-flush buffer when full
			if len(metricsBuffer) >= rps {
				second++
				metric := aggregateBatch(metricsBuffer)
				SaveMetrics(runID, second, metric)
				metricsBuffer = metricsBuffer[:0]
			}

		case <-ticker.C:
			second++
			if len(metricsBuffer) > 0 {
				metric := aggregateBatch(metricsBuffer)
				SaveMetrics(runID, second, metric)
				metricsBuffer = metricsBuffer[:0]
			}

			if second >= scenario.Duration {
				// Stop condition met
				cancel() // Signal request generator to stop
				goto cleanup
			}

		case <-ctx.Done():
			// Timeout or cancellation
			goto cleanup
		}
	}

cleanup:
	// Ensure all requests are processed before cleanup
	// Drain remaining results
	for len(results) > 0 {
		result := <-results
		metricsBuffer = append(metricsBuffer, result)
		if len(metricsBuffer) >= rps {
			second++
			metric := aggregateBatch(metricsBuffer)
			SaveMetrics(runID, second, metric)
			metricsBuffer = metricsBuffer[:0]
		}
	}

	// Save any remaining metrics
	if len(metricsBuffer) > 0 {
		second++
		metric := aggregateBatch(metricsBuffer)
		SaveMetrics(runID, second, metric)
	}

	// Update test run status
	now := time.Now()
	db.DB.Model(&models.TestRun{}).Where("id = ?", runID).Updates(map[string]any{
		"status":   "completed",
		"ended_at": now,
		"duration": second,
	})

	GenerateReport(runID, userID)
}