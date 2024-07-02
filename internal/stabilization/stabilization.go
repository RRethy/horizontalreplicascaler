package stabilization

import (
	"strings"
	"sync"
	"time"

	"k8s.io/utils/clock"
)

type RollingWindowType byte

const (
	MaxRollingWindow RollingWindowType = iota
	MinRollingWindow
)

func KeyFor(s ...string) string {
	return strings.Join(s, "/")
}

type Option func(*Window)

func WithClock(clock clock.Clock) Option {
	return func(w *Window) {
		w.Clock = clock
	}
}

type Event struct {
	Value     int32
	Timestamp time.Time
}

type Window struct {
	Clock         clock.Clock
	Mutex         sync.RWMutex
	Type          RollingWindowType
	RollingEvents map[string][]Event
}

func NewWindow(rollingWindowType RollingWindowType, options ...Option) *Window {
	w := &Window{
		Clock:         clock.RealClock{},
		Mutex:         sync.RWMutex{},
		Type:          rollingWindowType,
		RollingEvents: make(map[string][]Event),
	}

	for _, option := range options {
		option(w)
	}

	return w
}

// TODO: pass in status
// TODO: don't use seconds
func (w *Window) AddEvent(key string, value, windowSeconds int32) {
	w.Mutex.Lock()
	defer w.Mutex.Unlock()

	window := w.RollingEvents[key]
	t := w.Clock.Now()

	for len(window) > 0 && window[0].Timestamp.Add(time.Duration(windowSeconds)*time.Second).Before(t) {
		window = window[1:]
	}

	switch w.Type {
	case MaxRollingWindow:
		for len(window) > 0 && window[len(window)-1].Value < value {
			window = window[:len(window)-1]
		}
	case MinRollingWindow:
		for len(window) > 0 && window[len(window)-1].Value > value {
			window = window[:len(window)-1]
		}
	default:
		panic("invalid rolling window type")
	}

	window = append(window, Event{Value: value, Timestamp: t})
	w.RollingEvents[key] = window
}

func (w *Window) GetStabilizedValue(key string) (value int32, ok bool) {
	w.Mutex.RLock()
	defer w.Mutex.RUnlock()

	window, ok := w.RollingEvents[key]
	if !ok {
		return 0, false
	}

	if len(window) == 0 {
		return 0, false
	}

	return window[len(window)-1].Value, true
}
