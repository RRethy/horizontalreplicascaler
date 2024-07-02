package stabilization

import (
	"strings"
	"sync"
	"time"

	"k8s.io/utils/clock"
)

// RollingWindowType is the type of rolling window.
type RollingWindowType byte

const (
	// MaxRollingWindow is a rolling window that keeps the max value over the duration.
	MaxRollingWindow RollingWindowType = iota
	// MinRollingWindow is a rolling window that keeps the min value over the duration.
	MinRollingWindow
)

// KeyFor returns a key for the given strings.
// This provides a consistent way to generate keys for the RollingEvents map.
func KeyFor(s ...string) string {
	return strings.Join(s, "/")
}

// Option is a function that configures a Window.
type Option func(*Window)

// WithClock sets the clock used by the window.
// This is useful for testing where a fake clock is used.
func WithClock(clock clock.Clock) Option {
	return func(w *Window) {
		w.Clock = clock
	}
}

// Event represents a value in the rolling window at a given time.
type Event struct {
	Value     int32
	Timestamp time.Time
}

// Window is a thread-safe struct that implements keyed rolling windows.
// The window will only keep track of events that are within a given window duration.
type Window struct {
	// Clock is used to get the current time.
	// It is mocked in tests.
	Clock clock.Clock
	// Mutex is used to synchronize access to the RollingEvents map.
	Mutex sync.RWMutex
	// Type is the type of rolling window.
	// It can be either MaxRollingWindow or MinRollingWindow.
	Type RollingWindowType
	// RollingEvents is a map of keys to a list of events.
	// Only events that are within the window duration,
	// and can be the min/max are kept.
	RollingEvents map[string][]Event
}

// NewWindow creates a new Window with the given rolling window type and options.
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

// Stabilize is a thread-safe method which adds an event to the rolling window for the given key,
// and returns the stabilized value over the window duration.
// It runs in amortized O(1) time.
// TODO: pass in status
func (w *Window) Stabilize(key string, value int32, windowDuration time.Duration) (stabilized int32) {
	w.Mutex.Lock()
	defer w.Mutex.Unlock()

	window := w.RollingEvents[key]
	t := w.Clock.Now()

	for len(window) > 0 && window[0].Timestamp.Add(windowDuration).Before(t) {
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
	return window[0].Value
}
