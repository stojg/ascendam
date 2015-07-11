package main

import "time"

type Event struct {
	LoadTime time.Duration
	State    bool
}

type EventList struct {
	Events       []Event
	Outages      int
	UpDuration   time.Duration
	DownDuration time.Duration
	LastTime     time.Time
}

func (e *EventList) Add(state bool, loadTime time.Duration) {

	e.calculateDuration(state)
	e.LastTime = time.Now()

	if state == DOWN {
		e.Outages += 1
	}

	e.Events = append(e.Events, Event{
		State:    state,
		LoadTime: loadTime,
	})
}

func (e *EventList) AvgLoadTime() time.Duration {
	var sumLoadTime time.Duration
	upChecks := 0
	for _, event := range e.Events {
		if event.State != UP {
			continue
		}
		upChecks += 1
		sumLoadTime += event.LoadTime
	}
	if sumLoadTime.Nanoseconds() == 0 {
		return 0
	}
	return sumLoadTime / time.Duration(upChecks)
}

func (e *EventList) calculateDuration(state bool) {
	// don't save up or down duration on the first run
	if e.LastTime.IsZero() {
		return
	}
	if state == UP {
		e.UpDuration += time.Now().Sub(e.LastTime)
	}
	if state == DOWN {
		e.DownDuration += time.Now().Sub(e.LastTime)
	}
}
