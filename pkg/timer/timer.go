package timer

import (
	"errors"
	"sync"
	"time"
)

type Timer interface {
	Start(duration time.Duration, task func()) error
	Stop() error
}

type PeriodicTimer struct {
	ticker *time.Ticker
	stopCh chan struct{}
	mu     sync.Mutex
	active bool
}

func NewPeriodicTimer() *PeriodicTimer {
	return &PeriodicTimer{
		stopCh: make(chan struct{}),
	}
}

func (pt *PeriodicTimer) Start(duration time.Duration, task func()) error {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	if pt.active {
		return errors.New("timer is already running")
	}

	pt.ticker = time.NewTicker(duration)
	pt.active = true

	go func() {
		for {
			select {
			case <-pt.ticker.C:
				task()
			case <-pt.stopCh:
				pt.ticker.Stop()
				return
			}
		}
	}()

	return nil
}

func (pt *PeriodicTimer) Stop() error {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	if !pt.active {
		return errors.New("timer is not running")
	}

	close(pt.stopCh)
	pt.active = false
	return nil
}

type DelayedExecutionTimer struct {
	timer *time.Timer
	mu    sync.Mutex
}

func NewDelayedExecutionTimer() *DelayedExecutionTimer {
	return &DelayedExecutionTimer{}
}

func (det *DelayedExecutionTimer) Start(duration time.Duration, task func()) error {
	det.mu.Lock()
	defer det.mu.Unlock()

	if det.timer != nil {
		return errors.New("timer is already scheduled")
	}

	det.timer = time.AfterFunc(duration, task)
	return nil
}

func (det *DelayedExecutionTimer) Stop() error {
	det.mu.Lock()
	defer det.mu.Unlock()

	if det.timer == nil {
		return errors.New("no timer is scheduled")
	}

	if !det.timer.Stop() {
		return errors.New("timer already expired")
	}

	det.timer = nil
	return nil
}
