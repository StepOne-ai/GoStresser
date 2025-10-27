package engine

import (
	"time"
	"github.com/StepOne-ai/GoStresser/models"
)

func aggregateBatch(batch []LoadResult) models.TestMetric {
	if len(batch) == 0 {
		return models.TestMetric{
			RPS:        0,
			AvgLatency: 0,
			MaxLatency: 0,
			MinLatency: 0,
			ErrorCount: 0,
			CPU:        0,
			RAM:        0,
		}
	}

	var (
		totalLatency time.Duration
		maxLatency   = batch[0].Latency
		minLatency   = batch[0].Latency
		errorCount   int
	)

	for _, r := range batch {
		if r.Error != nil || r.Status >= 500 {
			errorCount++
		}
		totalLatency += r.Latency
		if r.Latency > maxLatency {
			maxLatency = r.Latency
		}
		if r.Latency < minLatency {
			minLatency = r.Latency
		}
	}

	avgLatency := float64(totalLatency.Milliseconds()) / float64(len(batch))

	return models.TestMetric{
		RPS:        len(batch),
		AvgLatency: avgLatency,
		MaxLatency: float64(maxLatency.Milliseconds()),
		MinLatency: float64(minLatency.Milliseconds()),
		ErrorCount: errorCount,
		CPU:        0, // TODO: implement if needed
		RAM:        0,
	}
}