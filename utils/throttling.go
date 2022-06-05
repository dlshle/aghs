package utils

import (
	"fmt"
	"sync"
	"time"

	"github.com/dlshle/gommon/ctimer"
)

const (
	RecordCleanJobInterval = time.Minute * 5
)

type Record struct {
	Id               string
	WindowExpiration time.Time
	WindowDuration   time.Duration
	HitsUnderWindow  int
	Limit            int
}

func NewThrottleRecord(id string, limit int, duration time.Duration) *Record {
	return &Record{
		Id:               id,
		WindowExpiration: time.Now().Add(duration).UTC(),
		WindowDuration:   duration,
		HitsUnderWindow:  1,
		Limit:            limit,
	}
}

func (r *Record) resetThrottleWindowBy(windowBegin time.Time) {
	r.WindowExpiration = windowBegin.Add(r.WindowDuration)
	if r.HitsUnderWindow > r.Limit {
		r.HitsUnderWindow -= r.Limit
		exceededHits := r.HitsUnderWindow - r.Limit
		// penalty on exceeded hits are added to the next throttling window
		r.WindowExpiration = r.WindowExpiration.Add(time.Minute * time.Duration(exceededHits))
	}
	r.HitsUnderWindow = 0
}

func (r *Record) Hit() (Record, error) {
	now := time.Now().UTC()
	if r.WindowExpiration.Before(now) {
		r.resetThrottleWindowBy(now)
	}
	if r.HitsUnderWindow > r.Limit {
		return *r, fmt.Errorf(
			"number of requests exceeded throttle limit in window by %d hits, next window begins at %s",
			r.HitsUnderWindow-r.Limit,
			r.WindowExpiration.String())
	}
	r.HitsUnderWindow++
	return *r, nil
}

type Controller interface {
	Hit(id string, limit int, duration time.Duration) (Record, error)
	Clear()
}

type controller struct {
	windowMap     *sync.Map
	cleanJobTimer ctimer.CTimer
}

func NewThrottleController() Controller {
	controller := &controller{
		windowMap: new(sync.Map),
	}
	cleanJobTimer := ctimer.New(RecordCleanJobInterval, controller.cleanJob)
	controller.cleanJobTimer = cleanJobTimer
	cleanJobTimer.Repeat()
	return controller
}

// periodically clean the expired records
func (c *controller) cleanJob() {
	now := time.Now()
	cleaned := 0
	c.windowMap.Range(func(key, value interface{}) bool {
		record := value.(*Record)
		// if now is later than window end time, that means the window(record) did not receive hit in current window
		if now.After(record.WindowExpiration) {
			c.windowMap.Delete(key)
			cleaned++
		}
		return true
	})
}

func (c *controller) Hit(id string, limit int, duration time.Duration) (Record, error) {
	record := c.createOrLoadRecord(id, limit, duration)
	return record.Hit()
}

func (c *controller) createOrLoadRecord(id string, limit int, duration time.Duration) *Record {
	record, _ := c.windowMap.LoadOrStore(id, NewThrottleRecord(id, limit, duration))
	return record.(*Record)
}

func (c *controller) Clear() {
	c.windowMap.Range(func(key, value interface{}) bool {
		c.windowMap.Delete(key)
		return true
	})
}
