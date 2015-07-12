package main

import (
	"errors"
	"testing"
	"time"
)

func TestLoadTime(t *testing.T) {
	c := &Check{}
	c.Start()
	time.Sleep(time.Millisecond)
	c.Stop()
	loadTime := c.LoadTime()
	if loadTime.Nanoseconds() == 0 {
		t.Errorf("expected Loadtime to be above 0")
	}
	if loadTime < time.Millisecond {
		t.Errorf("expected Loadtime to be more than %s, got %s", time.Millisecond, loadTime)
	}
}

func TestUptime(t *testing.T) {
	list := &CheckList{}
	check := &Check{startTime: time.Now()}
	list.Add(check)
	time.Sleep(time.Millisecond)
	uptime := list.Uptime()
	if uptime.Nanoseconds() == 0 {
		t.Errorf("expected Uptime to be above 0")
	}
	if uptime < time.Millisecond {
		t.Errorf("expected Uptime to be more than %s, got %s", time.Millisecond, uptime)
	}
	downTime := list.Downtime()
	if downTime.Nanoseconds() != 0 {
		t.Errorf("expected Downtime to be 0")
	}
}

func TestDowntime(t *testing.T) {
	check := &Check{startTime: time.Now(), error: errors.New("fake error")}
	list := &CheckList{}
	list.Add(check)
	time.Sleep(time.Millisecond)

	uptime := list.Uptime()
	if uptime.Nanoseconds() != 0 {
		t.Errorf("expected Uptime to be 0, got %s", uptime)
	}

	downTime := list.Downtime()
	if downTime.Nanoseconds() == 0 {
		t.Errorf("expected Downtime to more than 0")
	}
}

func TestUpDownUp(t *testing.T) {
	list := &CheckList{}
	upcheck := &Check{startTime: time.Now()}
	list.Add(upcheck)
	time.Sleep(time.Millisecond)

	downcheck := &Check{startTime: time.Now(), error: errors.New("fake error")}
	list.Add(downcheck)
	time.Sleep(time.Millisecond)

	upcheck = &Check{startTime: time.Now()}
	list.Add(upcheck)
	time.Sleep(time.Millisecond)

	upTime := list.Uptime()
	downTime := list.Downtime()

	if list.Down() != 1 {
		t.Errorf("There should be 1 recorded down checks")
	}
	if list.Up() != 2 {
		t.Errorf("There should be 2 recorded up checks")
	}

	if upTime.Nanoseconds() < downTime.Nanoseconds() {
		t.Errorf("downtime (%s) should be less than uptime (%s)", downTime, upTime)
	}
}

func TestDownUpDown(t *testing.T) {
	list := &CheckList{}

	downcheck := &Check{startTime: time.Now(), error: errors.New("fake error")}
	list.Add(downcheck)
	time.Sleep(time.Millisecond)

	if list.downTime != 0 {
		t.Errorf("check.downtime shouldn't have been set at this point")
	}

	upcheck := &Check{startTime: time.Now()}
	list.Add(upcheck)
	time.Sleep(time.Millisecond)

	if list.downTime == 0 {
		t.Errorf("check.downtime should have been set at this point")
	}

	downcheck = &Check{startTime: time.Now(), error: errors.New("fake error")}
	list.Add(downcheck)
	time.Sleep(time.Millisecond)

	upTime := list.Uptime()
	downTime := list.Downtime()

	if list.Down() != 2 {
		t.Errorf("There should be 2 recorded down checks")
	}
	if list.Up() != 1 {
		t.Errorf("There should be 1 recorded up checks")
	}

	if upTime.Nanoseconds() > downTime.Nanoseconds() {
		t.Errorf("downtime (%s) should be higher than uptime (%s)", downTime, upTime)
	}
}

func TestUpUP(t *testing.T) {
	list := &CheckList{}

	upcheck := &Check{startTime: time.Now()}
	list.Add(upcheck)
	time.Sleep(time.Millisecond)

	upcheck = &Check{startTime: time.Now()}
	list.Add(upcheck)
	time.Sleep(time.Millisecond)

	upTime := list.Uptime()
	downTime := list.Downtime()

	if downTime != 0 {
		t.Errorf("check.downtime should be zero, got %s", downTime)
	}

	if upTime > 2*time.Millisecond && upTime*time.Millisecond < 3 {
		t.Errorf("check.uptime %s", upTime)
	}
}

func TestDownDown(t *testing.T) {

	list := &CheckList{}

	downcheck := &Check{startTime: time.Now(), error: errors.New("fake error")}
	list.Add(downcheck)
	time.Sleep(time.Millisecond)

	downcheck = &Check{startTime: time.Now(), error: errors.New("fake error")}
	list.Add(downcheck)
	time.Sleep(time.Millisecond)

	upTime := list.Uptime()
	downTime := list.Downtime()

	if upTime != 0 {
		t.Errorf("check.downtime shouldn't be zero, got %s", downTime)
	}

	if downTime > 2*time.Millisecond && downTime*time.Millisecond < 3 {
		t.Errorf("check.dowtime %s", downTime)
	}
}

func TestDownDownUpUp(t *testing.T) {
	list := &CheckList{}

	downcheck := &Check{startTime: time.Now(), error: errors.New("fake error")}
	list.Add(downcheck)
	time.Sleep(time.Millisecond)

	downcheck = &Check{startTime: time.Now(), error: errors.New("fake error")}
	list.Add(downcheck)
	time.Sleep(time.Millisecond)

	upcheck := &Check{startTime: time.Now()}
	list.Add(upcheck)
	time.Sleep(time.Millisecond)

	upcheck = &Check{startTime: time.Now()}
	list.Add(upcheck)
	time.Sleep(time.Millisecond)

	upTime := list.Uptime()
	downTime := list.Downtime()

	if upTime > 3*time.Millisecond || upTime < 0 {
		t.Errorf("check.uptime shouldn't be above 5 or below 0, got %s", upTime)
	}

	if downTime > 3*time.Millisecond || downTime < 0 {
		t.Errorf("check.downtime shouldn't be above 5 or below 0, got %s", downTime)
	}

}
