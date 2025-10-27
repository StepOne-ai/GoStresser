package engine

import (
	"sort"

	db "github.com/StepOne-ai/GoStresser/database"
	"github.com/StepOne-ai/GoStresser/models"
)

func SaveMetrics(runID uint, timestamp int, metric models.TestMetric) {
	metric.RunID = runID
	metric.Timestamp = timestamp
	db.DB.Create(&metric)
}

func GenerateReport(runID uint) {
	var metrics []models.TestMetric
	db.DB.Where("run_id = ?", runID).Find(&metrics)

	if len(metrics) == 0 {
		// No data — create empty report
		db.DB.Create(&models.TestResult{
			RunID:         runID,
			TotalRequests: 0,
			AvgLatency:    0,
			P95Latency:    0,
			P99Latency:    0,
			ErrorRate:     0,
		})
		return
	}

	var (
		totalRequests int
		totalErrors   int
		allLatencies  []float64
	)

	for _, m := range metrics {
		totalRequests += m.RPS
		totalErrors += m.ErrorCount
		// Use per-second avg as representative latency sample
		if m.RPS > 0 {
			allLatencies = append(allLatencies, m.AvgLatency)
		}
	}

	if totalRequests == 0 {
		totalRequests = 1 // avoid div/0
	}

	errorRate := float64(totalErrors) / float64(totalRequests) * 100

	var avgLat, p95, p99 float64
	if len(allLatencies) > 0 {
		sort.Float64s(allLatencies)
		n := len(allLatencies)
		avgLat = allLatencies[n/2]
		if n > 1 {
			p95 = allLatencies[min(n-1, int(float64(n)*0.95))]
			p99 = allLatencies[min(n-1, int(float64(n)*0.99))]
		} else {
			p95 = avgLat
			p99 = avgLat
		}
	}

	result := models.TestResult{
		RunID:         runID,
		TotalRequests: totalRequests,
		AvgLatency:    avgLat,
		P95Latency:    p95,
		P99Latency:    p99,
		ErrorRate:     errorRate,
	}

	db.DB.Create(&result)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}