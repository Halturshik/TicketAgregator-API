package ratelimit

import "time"

type Policy struct {
	Name   string
	Limit  int64
	Window time.Duration
}

type Decision struct {
	Allowed    bool
	RetryAfter time.Duration
}
