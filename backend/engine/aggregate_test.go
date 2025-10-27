package engine

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAggregateBatch(t *testing.T) {
	batch := []LoadResult{
		{Latency: 100 * time.Millisecond, Status: 200},
		{Latency: 200 * time.Millisecond, Status: 500}, // error
		{Latency: 150 * time.Millisecond, Error: nil, Status: 200},
	}

	metric := aggregateBatch(batch)

	assert.Equal(t, 3, metric.RPS)
	assert.Equal(t, 1, metric.ErrorCount)
	assert.InDelta(t, 150.0, metric.AvgLatency, 0.1)
	assert.Equal(t, 200.0, metric.MaxLatency)
	assert.Equal(t, 100.0, metric.MinLatency)
}