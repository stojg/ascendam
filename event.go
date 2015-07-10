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

	// if this is an UP event add now - lastTime to Uptime
	if !e.LastTime.IsZero() {
		if state == UP {
			e.UpDuration += (time.Now().Sub(e.LastTime))
		}

		if state == DOWN {
			e.DownDuration += (time.Now().Sub(e.LastTime))
		}
	}

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
	var avgTime time.Duration
	nonOutages := 0
	for _, event := range e.Events {
		if event.State != UP {
			continue
		}
		avgTime = avgTime + event.LoadTime
		nonOutages += 1
	}
	if avgTime.Nanoseconds() == 0 {
		return 0
	}
	return avgTime / time.Duration(nonOutages)
}