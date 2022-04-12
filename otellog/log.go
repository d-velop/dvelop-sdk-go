package otellog

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type Logger struct {
	mu              sync.Mutex
	out             io.Writer
	outputFormatter OutputFormatter
	time            Time
	hooks           []Hook
}

type Time func() time.Time

type Hook func(ctx context.Context, e *Event)

type OutputFormatter func(e *Event) ([]byte, error)

// New creates a new Logger.
func New() *Logger {
	logger := Logger{}
	logger.Reset()
	return &logger
}

var std = New()

// Default returns the standard logger used by the package-level output functions.
func Default() *Logger {
	return std
}

// Reset resets the logger to the default settings.
func (l *Logger) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.hooks = nil
	l.out = os.Stdout
	l.time = time.Now
	l.outputFormatter = func(e *Event) ([]byte, error) {
		return json.Marshal(e)
	}
}

// output writes the output for a logging event.
func (l *Logger) output(ctx context.Context, sev Severity, msg interface{}, options []Option) {
	l.mu.Lock()
	defer l.mu.Unlock()

	t := l.time()
	e := Event{
		Time:     &t,
		Severity: sev,
		Body:     msg,
	}

	for _, h := range l.hooks {
		h(ctx, &e)
	}

	for _, o := range options {
		o(&e)
	}

	s, err := l.outputFormatter(&e)
	if err == nil {
		if len(s) == 0 || s[len(s)-1] != '\n' {
			s = append(s, '\n')
		}
		l.out.Write(s)
	}
}

// SetOutput sets the output destination for the logger.
func SetOutput(w io.Writer) {
	std.mu.Lock()
	defer std.mu.Unlock()
	std.out = w
}

// SetTime sets the default clock for outputting the timestamp in the log statement.
func SetTime(time Time) {
	std.mu.Lock()
	defer std.mu.Unlock()
	std.time = time
}

// SetOutputFormatter sets a callback function that will be called when this logger writes a log statement.
func SetOutputFormatter(f OutputFormatter) {
	std.mu.Lock()
	defer std.mu.Unlock()
	std.outputFormatter = f
}

// RegisterHook adds a callback function that will be called before the logger writes the log statement.
// Inside the callback function the log event can be extended.
func RegisterHook(h Hook) {
	std.mu.Lock()
	defer std.mu.Unlock()
	std.hooks = append(std.hooks, h)
}

// Debug logs an event body according to the otel definition
func Debug(ctx context.Context, body interface{}) {
	std.output(ctx, SeverityDebug, body, nil)
}

// Debugf is equivalent to log.StdDebug.Printf()
func Debugf(ctx context.Context, format string, v ...interface{}) {
	std.output(ctx, SeverityDebug, fmt.Sprintf(format, v...), nil)
}

// Info logs an event body according to the otel definition
func Info(ctx context.Context, body interface{}) {
	std.output(ctx, SeverityInfo, body, nil)
}

// Infof is equivalent to log.StdInfo.Printf()
func Infof(ctx context.Context, format string, v ...interface{}) {
	std.output(ctx, SeverityInfo, fmt.Sprintf(format, v...), nil)
}

// Error logs an event body according to the otel definition
func Error(ctx context.Context, body interface{}) {
	std.output(ctx, SeverityError, body, nil)
}

// Errorf is equivalent to log.StdError.Printf()
func Errorf(ctx context.Context, format string, v ...interface{}) {
	std.output(ctx, SeverityError, fmt.Sprintf(format, v...), nil)
}
