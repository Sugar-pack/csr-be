package timer

import "time"

type Timer interface {
	Start(duration time.Duration, task func())
	Stop()
}

type PeriodicTimer struct {
	ticker *time.Ticker
	stopCh chan struct{}
}

func NewPeriodicTimer() *PeriodicTimer {
	return &PeriodicTimer{
		stopCh: make(chan struct{}),
	}
}

func (pt *PeriodicTimer) Start(duration time.Duration, task func()) {
	pt.ticker = time.NewTicker(duration)
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
}

func (pt *PeriodicTimer) Stop() {
	close(pt.stopCh)
}

type DelayedExecutionTimer struct {
	timer *time.Timer
}

func NewDelayedExecutionTimer() *DelayedExecutionTimer {
	return &DelayedExecutionTimer{}
}

func (det *DelayedExecutionTimer) Schedule(duration time.Duration, task func()) {
	det.timer = time.AfterFunc(duration, task)
}

func (det *DelayedExecutionTimer) Stop() {
	if det.timer != nil {
		det.timer.Stop()
	}
}
