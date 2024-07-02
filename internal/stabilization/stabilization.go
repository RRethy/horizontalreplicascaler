package stabilization

import (
	"fmt"
	"sync"
	"time"
)

type RollingWindowType byte

const (
	MaxRollingWindow RollingWindowType = iota
	MinRollingWindow
)

type Event struct {
	Value     int32
	Timestamp time.Time
}

type Window struct {
	Mutex         sync.RWMutex
	Type          RollingWindowType
	RollingEvents map[string][]Event
}

func NewWindow(rollingWindowType RollingWindowType) *Window {
	return &Window{
		Mutex:         sync.RWMutex{},
		Type:          rollingWindowType,
		RollingEvents: make(map[string][]Event),
	}
}

// TODO: pass in status
func (w *Window) AddEvent(name, namespace string, value, windowSeconds int32) {
	key := w.keyFor(name, namespace)

	w.Mutex.Lock()
	defer w.Mutex.Unlock()

	window := w.RollingEvents[key]
	t := time.Now()

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

func (w *Window) GetStabilizedValue(name, namespace string) (value int32, ok bool) {
	key := w.keyFor(name, namespace)

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

func (w *Window) keyFor(name, namespace string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}
