package engine

import "time"

type LoadResult struct {
    Latency time.Duration
    Error   error
    Status  int
}