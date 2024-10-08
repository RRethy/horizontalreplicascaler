package stabilization

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clock "k8s.io/utils/clock/testing"

	rrethyv1 "github.com/RRethy/horizontalreplicascaler/api/v1"
)

var initialTime = metav1.NewTime(time.Date(1997, time.November, 7, 0, 0, 0, 0, time.UTC))

func TestWindow_Stabilize(t *testing.T) {
	tests := []struct {
		testName           string
		rollingWindowType  RollingWindowType
		initialEvents      map[string][]rrethyv1.ScaleEvent
		currentTime        time.Time
		key                string
		value              int32
		windowDuration     time.Duration
		expectedEvents     map[string][]rrethyv1.ScaleEvent
		expectedStabilized int32
		expectedOk         bool
	}{
		{
			testName:          "max rolling window has max event at head",
			rollingWindowType: MaxRollingWindow,
			initialEvents: map[string][]rrethyv1.ScaleEvent{"foobar": {
				{Value: 5, Timestamp: metav1.NewTime(initialTime.Add(-10000 * time.Second))}, // we need an event outside the window
				{Value: 4, Timestamp: initialTime},
				{Value: 3, Timestamp: initialTime},
				{Value: 1, Timestamp: initialTime},
			}},
			currentTime:    initialTime.Add(1 * time.Millisecond),
			key:            "foobar",
			value:          2,
			windowDuration: 10000 * time.Second,
			expectedEvents: map[string][]rrethyv1.ScaleEvent{"foobar": {
				{Value: 4, Timestamp: initialTime},
				{Value: 3, Timestamp: initialTime},
				{Value: 2, Timestamp: metav1.NewTime(initialTime.Add(1 * time.Millisecond))},
			}},
			expectedStabilized: 4,
			expectedOk:         true,
		},
		{
			testName:          "min rolling window has min event at head",
			rollingWindowType: MinRollingWindow,
			initialEvents: map[string][]rrethyv1.ScaleEvent{"foobar": {
				{Value: 0, Timestamp: metav1.NewTime(initialTime.Add(-1 * time.Second))}, // we need an event outside the window
				{Value: 1, Timestamp: initialTime},
				{Value: 3, Timestamp: initialTime},
				{Value: 7, Timestamp: initialTime},
			}},
			currentTime:    initialTime.Add(1 * time.Millisecond),
			key:            "foobar",
			value:          5,
			windowDuration: 1 * time.Second,
			expectedEvents: map[string][]rrethyv1.ScaleEvent{"foobar": {
				{Value: 1, Timestamp: initialTime},
				{Value: 3, Timestamp: initialTime},
				{Value: 5, Timestamp: metav1.NewTime(initialTime.Add(1 * time.Millisecond))},
			}},
			expectedStabilized: 1,
			expectedOk:         true,
		},
		{
			testName:          "values outside the window are removed",
			rollingWindowType: MaxRollingWindow,
			initialEvents: map[string][]rrethyv1.ScaleEvent{"foobar": {
				{Value: 6, Timestamp: initialTime},
				{Value: 4, Timestamp: metav1.NewTime(initialTime.Add(5 * time.Second))},
				{Value: 3, Timestamp: metav1.NewTime(initialTime.Add(10 * time.Second))},
			}},
			currentTime:    initialTime.Add(11 * time.Second),
			key:            "foobar",
			value:          1,
			windowDuration: 10 * time.Second,
			expectedEvents: map[string][]rrethyv1.ScaleEvent{"foobar": {
				{Value: 4, Timestamp: metav1.NewTime(initialTime.Add(5 * time.Second))},
				{Value: 3, Timestamp: metav1.NewTime(initialTime.Add(10 * time.Second))},
				{Value: 1, Timestamp: metav1.NewTime(initialTime.Add(11 * time.Second))},
			}},
			expectedStabilized: 4,
			expectedOk:         true,
		},
		{
			testName:          "values inside the window are kept",
			rollingWindowType: MaxRollingWindow,
			initialEvents: map[string][]rrethyv1.ScaleEvent{"foobar": {
				{Value: 6, Timestamp: metav1.NewTime(initialTime.Add(-10 * time.Second))}, // we need an event outside the window
				{Value: 6, Timestamp: initialTime},
				{Value: 4, Timestamp: metav1.NewTime(initialTime.Add(5 * time.Second))},
				{Value: 3, Timestamp: metav1.NewTime(initialTime.Add(10 * time.Second))},
			}},
			currentTime:    initialTime.Add(11 * time.Second),
			key:            "foobar",
			value:          1,
			windowDuration: 20 * time.Second,
			expectedEvents: map[string][]rrethyv1.ScaleEvent{"foobar": {
				{Value: 6, Timestamp: initialTime},
				{Value: 4, Timestamp: metav1.NewTime(initialTime.Add(5 * time.Second))},
				{Value: 3, Timestamp: metav1.NewTime(initialTime.Add(10 * time.Second))},
				{Value: 1, Timestamp: metav1.NewTime(initialTime.Add(11 * time.Second))},
			}},
			expectedStabilized: 6,
			expectedOk:         true,
		},
		{
			testName:          "max stabilization window of 0 seconds keep single value",
			rollingWindowType: MaxRollingWindow,
			initialEvents: map[string][]rrethyv1.ScaleEvent{"foobar": {
				{Value: 6, Timestamp: initialTime},
				{Value: 4, Timestamp: metav1.NewTime(initialTime.Add(1 * time.Second))},
				{Value: 3, Timestamp: metav1.NewTime(initialTime.Add(2 * time.Second))},
			}},
			currentTime:    initialTime.Add(3 * time.Second),
			key:            "foobar",
			value:          1,
			windowDuration: 0 * time.Second,
			expectedEvents: map[string][]rrethyv1.ScaleEvent{"foobar": {
				{Value: 1, Timestamp: metav1.NewTime(initialTime.Add(3 * time.Second))},
			}},
			expectedStabilized: 1,
			expectedOk:         true,
		},
		{
			testName:          "min stabilization window of 0 seconds keep single value",
			rollingWindowType: MinRollingWindow,
			initialEvents: map[string][]rrethyv1.ScaleEvent{"foobar": {
				{Value: 3, Timestamp: metav1.NewTime(initialTime.Add(2 * time.Second))},
				{Value: 4, Timestamp: metav1.NewTime(initialTime.Add(1 * time.Second))},
				{Value: 6, Timestamp: initialTime},
			}},
			currentTime:    initialTime.Add(3 * time.Second),
			key:            "foobar",
			value:          1,
			windowDuration: 0 * time.Second,
			expectedEvents: map[string][]rrethyv1.ScaleEvent{"foobar": {
				{Value: 1, Timestamp: metav1.NewTime(initialTime.Add(3 * time.Second))},
			}},
			expectedStabilized: 1,
			expectedOk:         true,
		},
		{
			testName:          "value greater than all others results in single value",
			rollingWindowType: MaxRollingWindow,
			initialEvents: map[string][]rrethyv1.ScaleEvent{"foobar": {
				{Value: 6, Timestamp: metav1.NewTime(initialTime.Add(-20 * time.Second))}, // we need an event outside the window
				{Value: 6, Timestamp: initialTime},
				{Value: 4, Timestamp: metav1.NewTime(initialTime.Add(1 * time.Second))},
				{Value: 3, Timestamp: metav1.NewTime(initialTime.Add(2 * time.Second))},
			}},
			currentTime:    initialTime.Add(3 * time.Second),
			key:            "foobar",
			value:          10,
			windowDuration: 20 * time.Second,
			expectedEvents: map[string][]rrethyv1.ScaleEvent{"foobar": {
				{Value: 10, Timestamp: metav1.NewTime(initialTime.Add(3 * time.Second))},
			}},
			expectedStabilized: 10,
			expectedOk:         true,
		},
		{
			testName:          "max rolling keeps duplicate values",
			rollingWindowType: MaxRollingWindow,
			initialEvents: map[string][]rrethyv1.ScaleEvent{"foobar": {
				{Value: 6, Timestamp: metav1.NewTime(initialTime.Add(-20 * time.Second))}, // we need an event outside the window
				{Value: 6, Timestamp: initialTime},
				{Value: 4, Timestamp: metav1.NewTime(initialTime.Add(1 * time.Second))},
				{Value: 3, Timestamp: metav1.NewTime(initialTime.Add(2 * time.Second))},
			}},
			currentTime:    initialTime.Add(3 * time.Second),
			key:            "foobar",
			value:          4,
			windowDuration: 20 * time.Second,
			expectedEvents: map[string][]rrethyv1.ScaleEvent{"foobar": {
				{Value: 6, Timestamp: initialTime},
				{Value: 4, Timestamp: metav1.NewTime(initialTime.Add(1 * time.Second))},
				{Value: 4, Timestamp: metav1.NewTime(initialTime.Add(3 * time.Second))},
			}},
			expectedStabilized: 6,
			expectedOk:         true,
		},
		{
			testName:          "min rolling keeps duplicate values",
			rollingWindowType: MinRollingWindow,
			initialEvents: map[string][]rrethyv1.ScaleEvent{"foobar": {
				{Value: 2, Timestamp: metav1.NewTime(initialTime.Add(-20 * time.Second))}, // we need an event outside the window
				{Value: 2, Timestamp: initialTime},
				{Value: 4, Timestamp: metav1.NewTime(initialTime.Add(1 * time.Second))},
				{Value: 5, Timestamp: metav1.NewTime(initialTime.Add(2 * time.Second))},
			}},
			currentTime:    initialTime.Add(3 * time.Second),
			key:            "foobar",
			value:          4,
			windowDuration: 20 * time.Second,
			expectedEvents: map[string][]rrethyv1.ScaleEvent{"foobar": {
				{Value: 2, Timestamp: initialTime},
				{Value: 4, Timestamp: metav1.NewTime(initialTime.Add(1 * time.Second))},
				{Value: 4, Timestamp: metav1.NewTime(initialTime.Add(3 * time.Second))},
			}},
			expectedStabilized: 2,
			expectedOk:         true,
		},
		{
			testName:          "multiple keys are handled independently",
			rollingWindowType: MinRollingWindow,
			initialEvents: map[string][]rrethyv1.ScaleEvent{
				"foobar": {
					{Value: 3, Timestamp: initialTime},
					{Value: 4, Timestamp: metav1.NewTime(initialTime.Add(1 * time.Second))},
					{Value: 5, Timestamp: metav1.NewTime(initialTime.Add(2 * time.Second))},
				},
				"barfoo": {
					{Value: 2, Timestamp: metav1.NewTime(initialTime.Add(-20 * time.Second))}, // we need an event outside the window
					{Value: 2, Timestamp: initialTime},
					{Value: 4, Timestamp: metav1.NewTime(initialTime.Add(1 * time.Second))},
					{Value: 5, Timestamp: metav1.NewTime(initialTime.Add(2 * time.Second))},
				},
			},
			currentTime:    initialTime.Add(3 * time.Second),
			key:            "barfoo",
			value:          4,
			windowDuration: 20 * time.Second,
			expectedEvents: map[string][]rrethyv1.ScaleEvent{
				"foobar": {
					{Value: 3, Timestamp: initialTime},
					{Value: 4, Timestamp: metav1.NewTime(initialTime.Add(1 * time.Second))},
					{Value: 5, Timestamp: metav1.NewTime(initialTime.Add(2 * time.Second))},
				},
				"barfoo": {
					{Value: 2, Timestamp: initialTime},
					{Value: 4, Timestamp: metav1.NewTime(initialTime.Add(1 * time.Second))},
					{Value: 4, Timestamp: metav1.NewTime(initialTime.Add(3 * time.Second))},
				},
			},
			expectedStabilized: 2,
			expectedOk:         true,
		},
		{
			testName:          "window is inclusive",
			rollingWindowType: MaxRollingWindow,
			initialEvents: map[string][]rrethyv1.ScaleEvent{"foobar": {
				{Value: 5, Timestamp: initialTime},
				{Value: 4, Timestamp: metav1.NewTime(initialTime.Add(1 * time.Second))},
				{Value: 3, Timestamp: metav1.NewTime(initialTime.Add(2 * time.Second))},
			}},
			currentTime:    initialTime.Add(3 * time.Second),
			key:            "foobar",
			value:          2,
			windowDuration: 2 * time.Second,
			expectedEvents: map[string][]rrethyv1.ScaleEvent{"foobar": {
				{Value: 4, Timestamp: metav1.NewTime(initialTime.Add(1 * time.Second))},
				{Value: 3, Timestamp: metav1.NewTime(initialTime.Add(2 * time.Second))},
				{Value: 2, Timestamp: metav1.NewTime(initialTime.Add(3 * time.Second))},
			}},
			expectedStabilized: 4,
			expectedOk:         true,
		},
		{
			testName:           "empty window with 0 duration returns event and true",
			rollingWindowType:  MaxRollingWindow,
			initialEvents:      map[string][]rrethyv1.ScaleEvent{},
			currentTime:        initialTime.Time,
			key:                "foobar",
			value:              1,
			windowDuration:     0,
			expectedEvents:     map[string][]rrethyv1.ScaleEvent{"foobar": {{Value: 1, Timestamp: initialTime}}},
			expectedStabilized: 1,
			expectedOk:         true,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			w := NewWindow(test.rollingWindowType, WithClock(clock.NewFakeClock(test.currentTime)))
			w.RollingEvents = test.initialEvents
			stabilized, ok := w.Stabilize(test.key, test.value, test.windowDuration)
			assert.Equal(t, test.expectedOk, ok)
			assert.Equal(t, test.expectedStabilized, stabilized)
			assert.Equal(t, test.expectedEvents, w.RollingEvents)
		})
	}
}
