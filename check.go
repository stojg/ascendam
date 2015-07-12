package main

import "time"

type Check struct {
	error      error
	statusCode int
	startTime  time.Time
	stopTime   time.Time
}

func (c *Check) Start() {
	c.startTime = time.Now()
}

func (c *Check) Stop() {
	c.stopTime = time.Now()
}

func (c *Check) LoadTime() time.Duration {
	return c.stopTime.Sub(c.startTime)
}

func (c *Check) Error() error {
	return c.error
}

func (c *Check) StatusCode() int {
	return c.statusCode
}

func (c *Check) Ok() bool {
	return c.error == nil
}
