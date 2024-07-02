package stabilization

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	clock "k8s.io/utils/clock/testing"
)

var initialTime = time.Date(1997, time.November, 7, 0, 0, 0, 0, time.UTC)

func TestWindow_AddEvent(t *testing.T) {
	tests := []struct {
		testName          string
		rollingWindowType RollingWindowType
		initialEvents     map[string][]Event
		currentTime       time.Time
		key               string
		value             int32
		windowDuration    time.Duration
		expectedEvents    map[string][]Event
	}{
		{
			testName:          "max rolling window has max event at head",
			rollingWindowType: MaxRollingWindow,
			initialEvents: map[string][]Event{"foobar": {
				{Value: 4, Timestamp: initialTime},
				{Value: 3, Timestamp: initialTime},
				{Value: 1, Timestamp: initialTime},
			}},
			currentTime:    initialTime.Add(1 * time.Millisecond),
			key:            "foobar",
			value:          2,
			windowDuration: 10000 * time.Second,
			expectedEvents: map[string][]Event{"foobar": {
				{Value: 4, Timestamp: initialTime},
				{Value: 3, Timestamp: initialTime},
				{Value: 2, Timestamp: initialTime.Add(1 * time.Millisecond)},
			}},
		},
		{
			testName:          "min rolling window has min event at head",
			rollingWindowType: MinRollingWindow,
			initialEvents: map[string][]Event{"foobar": {
				{Value: 1, Timestamp: initialTime},
				{Value: 3, Timestamp: initialTime},
				{Value: 7, Timestamp: initialTime},
			}},
			currentTime:    initialTime.Add(1 * time.Millisecond),
			key:            "foobar",
			value:          5,
			windowDuration: 1 * time.Second,
			expectedEvents: map[string][]Event{"foobar": {
				{Value: 1, Timestamp: initialTime},
				{Value: 3, Timestamp: initialTime},
				{Value: 5, Timestamp: initialTime.Add(1 * time.Millisecond)},
			}},
		},
		{
			testName:          "values outside the window are removed",
			rollingWindowType: MaxRollingWindow,
			initialEvents: map[string][]Event{"foobar": {
				{Value: 6, Timestamp: initialTime},
				{Value: 4, Timestamp: initialTime.Add(5 * time.Second)},
				{Value: 3, Timestamp: initialTime.Add(10 * time.Second)},
			}},
			currentTime:    initialTime.Add(11 * time.Second),
			key:            "foobar",
			value:          1,
			windowDuration: 10 * time.Second,
			expectedEvents: map[string][]Event{"foobar": {
				{Value: 4, Timestamp: initialTime.Add(5 * time.Second)},
				{Value: 3, Timestamp: initialTime.Add(10 * time.Second)},
				{Value: 1, Timestamp: initialTime.Add(11 * time.Second)},
			}},
		},
		{
			testName:          "values inside the window are kept",
			rollingWindowType: MaxRollingWindow,
			initialEvents: map[string][]Event{"foobar": {
				{Value: 6, Timestamp: initialTime},
				{Value: 4, Timestamp: initialTime.Add(5 * time.Second)},
				{Value: 3, Timestamp: initialTime.Add(10 * time.Second)},
			}},
			currentTime:    initialTime.Add(11 * time.Second),
			key:            "foobar",
			value:          1,
			windowDuration: 20 * time.Second,
			expectedEvents: map[string][]Event{"foobar": {
				{Value: 6, Timestamp: initialTime},
				{Value: 4, Timestamp: initialTime.Add(5 * time.Second)},
				{Value: 3, Timestamp: initialTime.Add(10 * time.Second)},
				{Value: 1, Timestamp: initialTime.Add(11 * time.Second)},
			}},
		},
		{
			testName:          "stabilization window of 0 seconds keep single value",
			rollingWindowType: MaxRollingWindow,
			initialEvents: map[string][]Event{"foobar": {
				{Value: 6, Timestamp: initialTime},
				{Value: 4, Timestamp: initialTime.Add(1 * time.Second)},
				{Value: 3, Timestamp: initialTime.Add(2 * time.Second)},
			}},
			currentTime:    initialTime.Add(3 * time.Second),
			key:            "foobar",
			value:          1,
			windowDuration: 0 * time.Second,
			expectedEvents: map[string][]Event{"foobar": {
				{Value: 1, Timestamp: initialTime.Add(3 * time.Second)},
			}},
		},
		{
			testName:          "value greater than all others results in single value",
			rollingWindowType: MaxRollingWindow,
			initialEvents: map[string][]Event{"foobar": {
				{Value: 6, Timestamp: initialTime},
				{Value: 4, Timestamp: initialTime.Add(1 * time.Second)},
				{Value: 3, Timestamp: initialTime.Add(2 * time.Second)},
			}},
			currentTime:    initialTime.Add(3 * time.Second),
			key:            "foobar",
			value:          10,
			windowDuration: 20 * time.Second,
			expectedEvents: map[string][]Event{"foobar": {
				{Value: 10, Timestamp: initialTime.Add(3 * time.Second)},
			}},
		},
		{
			testName:          "max rolling keeps duplicate values",
			rollingWindowType: MaxRollingWindow,
			initialEvents: map[string][]Event{"foobar": {
				{Value: 6, Timestamp: initialTime},
				{Value: 4, Timestamp: initialTime.Add(1 * time.Second)},
				{Value: 3, Timestamp: initialTime.Add(2 * time.Second)},
			}},
			currentTime:    initialTime.Add(3 * time.Second),
			key:            "foobar",
			value:          4,
			windowDuration: 20 * time.Second,
			expectedEvents: map[string][]Event{"foobar": {
				{Value: 6, Timestamp: initialTime},
				{Value: 4, Timestamp: initialTime.Add(1 * time.Second)},
				{Value: 4, Timestamp: initialTime.Add(3 * time.Second)},
			}},
		},
		{
			testName:          "min rolling keeps duplicate values",
			rollingWindowType: MinRollingWindow,
			initialEvents: map[string][]Event{"foobar": {
				{Value: 2, Timestamp: initialTime},
				{Value: 4, Timestamp: initialTime.Add(1 * time.Second)},
				{Value: 5, Timestamp: initialTime.Add(2 * time.Second)},
			}},
			currentTime:    initialTime.Add(3 * time.Second),
			key:            "foobar",
			value:          4,
			windowDuration: 20 * time.Second,
			expectedEvents: map[string][]Event{"foobar": {
				{Value: 2, Timestamp: initialTime},
				{Value: 4, Timestamp: initialTime.Add(1 * time.Second)},
				{Value: 4, Timestamp: initialTime.Add(3 * time.Second)},
			}},
		},
		{
			testName:          "multiple keys are handled independently",
			rollingWindowType: MinRollingWindow,
			initialEvents: map[string][]Event{
				"foobar": {
					{Value: 2, Timestamp: initialTime},
					{Value: 4, Timestamp: initialTime.Add(1 * time.Second)},
					{Value: 5, Timestamp: initialTime.Add(2 * time.Second)},
				},
				"barfoo": {
					{Value: 2, Timestamp: initialTime},
					{Value: 4, Timestamp: initialTime.Add(1 * time.Second)},
					{Value: 5, Timestamp: initialTime.Add(2 * time.Second)},
				},
			},
			currentTime:    initialTime.Add(3 * time.Second),
			key:            "barfoo",
			value:          4,
			windowDuration: 20 * time.Second,
			expectedEvents: map[string][]Event{
				"foobar": {
					{Value: 2, Timestamp: initialTime},
					{Value: 4, Timestamp: initialTime.Add(1 * time.Second)},
					{Value: 5, Timestamp: initialTime.Add(2 * time.Second)},
				},
				"barfoo": {
					{Value: 2, Timestamp: initialTime},
					{Value: 4, Timestamp: initialTime.Add(1 * time.Second)},
					{Value: 4, Timestamp: initialTime.Add(3 * time.Second)},
				},
			},
		},
		{
			testName:          "window is inclusive",
			rollingWindowType: MaxRollingWindow,
			initialEvents: map[string][]Event{"foobar": {
				{Value: 5, Timestamp: initialTime},
				{Value: 4, Timestamp: initialTime.Add(1 * time.Second)},
				{Value: 3, Timestamp: initialTime.Add(2 * time.Second)},
			}},
			currentTime:    initialTime.Add(3 * time.Second),
			key:            "foobar",
			value:          2,
			windowDuration: 2 * time.Second,
			expectedEvents: map[string][]Event{"foobar": {
				{Value: 4, Timestamp: initialTime.Add(1 * time.Second)},
				{Value: 3, Timestamp: initialTime.Add(2 * time.Second)},
				{Value: 2, Timestamp: initialTime.Add(3 * time.Second)},
			}},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			w := NewWindow(test.rollingWindowType, WithClock(clock.NewFakeClock(test.currentTime)))
			w.RollingEvents = test.initialEvents
			w.AddEvent(test.key, test.value, test.windowDuration)
			require.Equal(t, test.expectedEvents, w.RollingEvents)
		})
	}
}
