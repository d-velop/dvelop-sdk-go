package otellogtest

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/d-velop/dvelop-sdk-go/otellog"
)

type testingT interface {
	Errorf(format string, args ...any)
	Helper()
}

type LogRecorder struct {
	Events []otellog.Event
	t      testingT
}

// NewLogRecorder creates a new LogRecorder that can be used to assert log messages.
// The given testingT is used to fail the test if an assertion fails.
// The default log output formatter is replaced with a formatter that records all log messages.
// The default time function is replaced with a function that always returns the same time (2022-01-01 01:02:03.000000004 UTC).
func NewLogRecorder(t testingT) *LogRecorder {
	otellog.Default().Reset()
	rec := &LogRecorder{[]otellog.Event{}, t}

	otellog.SetOutputFormatter(func(event *otellog.Event) ([]byte, error) {
		rec.Events = append(rec.Events, *event)
		return []byte{}, nil
	})

	otellog.SetTime(func() time.Time {
		return time.Date(2022, time.January, 01, 1, 2, 3, 4, time.UTC)
	})
	return rec
}

// ShouldHaveLogged asserts that the given log message was logged at some point.
// The log message can be an otellog.Severity, string or any other type that can be converted to a string.
// If multiple arguments are given, they are treated as a logical AND.
func (l *LogRecorder) ShouldHaveLogged(conditions ...any) {
	l.t.Helper()

	for _, event := range l.Events {
		if matches(event, conditions...) {
			return
		}
	}

	l.t.Errorf("no log found matching %v", conditions)
}

func matches(event otellog.Event, conditions ...any) bool {
	for _, condition := range conditions {
		switch condition := condition.(type) {
		case otellog.Severity:
			if event.Severity != condition {
				return false
			}
		case int:
			if int(event.Severity) != condition {
				return false
			}

		case func(event otellog.Event) bool:
			if !condition(event) {
				return false
			}

		default:
			bodyAsString := fmt.Sprint(event.Body)
			conditionAsString := fmt.Sprint(condition)
			if !strings.Contains(bodyAsString, conditionAsString) {
				return false
			}
		}
	}
	return true
}

func ContainsAttribute(key string, value any) func(event otellog.Event) bool {
	return func(event otellog.Event) bool {
		bytes, _ := event.Attributes.MarshalJSON()
		attributes := map[string]string{}
		_ = json.Unmarshal(bytes, &attributes)
		return attributes[key] == value
	}
}
