package main

import (
	"time"
)

const (
	UP   = true
	DOWN = false
)

// @todo, need to be able configure this list so that for example 503 or other codes are measured
// as downtimes
type CheckList struct {
	startTime       time.Time
	outageStartTime time.Time
	checks          []*Check
	down            int
	up              int
	downTime        time.Duration
	loadTime        time.Duration
	lastState       bool
}

func (e *CheckList) Total() int {
	return e.up + e.down
}

func (e *CheckList) Down() int {
	return e.down
}

func (e *CheckList) Up() int {
	return e.up
}

func (e *CheckList) Add(check *Check) {
	// start the clock
	if e.up+e.down == 0 {
		e.startTime = check.startTime
	}

	if !check.Ok() {
		e.down += 1
		e.lastState = DOWN
		if e.outageStartTime.IsZero() {
			e.outageStartTime = check.startTime
		}
		return
	}

	e.up += 1
	if e.lastState == DOWN {
		e.downTime = check.startTime.Sub(e.outageStartTime)
	}
	e.lastState = UP
	e.loadTime += check.LoadTime()
}

func (e *CheckList) AvgLoadTime() time.Duration {
	if e.up == 0 {
		return 0
	}
	return e.loadTime / time.Duration(e.up)
}

func (c *CheckList) Downtime() time.Duration {
	if c.down == 0 {
		return 0
	}
	if c.up == 0 {
		return time.Since(c.outageStartTime)
	}
	if c.lastState == DOWN {
		c.downTime = time.Since(c.outageStartTime)
	}
	return c.downTime
}

func (c *CheckList) Uptime() time.Duration {
	if c.up == 0 {
		return 0
	}
	return time.Since(c.startTime) - c.Downtime()
}
