//go:generate moq -out mocks/scheduler_moq.go -pkg mocks . Scheduler

package authtoken

import "time"

type Scheduler interface {
	AfterFunc(d time.Duration, f func()) *time.Timer
}

type SchedulerFunc func(d time.Duration, f func()) *time.Timer

func (s SchedulerFunc) AfterFunc(d time.Duration, f func()) *time.Timer {
	return s(d, f)
}
