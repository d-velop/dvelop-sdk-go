package structuredlog

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
	mu           	sync.Mutex
	out             io.Writer
	outputFormatter OutputFormatterFunc
	time            Time
	hooks           []Hook
}

type Time func() time.Time

type Hook func(ctx context.Context, e *Event)

type OutputFormatterFunc func(e *Event, msg string) ([]byte, error)

func New() *Logger {
	logger := Logger{}
	logger.Init()
	return &logger
}

var std = New()

func Default() *Logger {
	return std
}

func (l *Logger) Init() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.hooks = nil
	l.out = os.Stdout
	l.time = time.Now
	l.outputFormatter = func(e *Event, msg string) ([]byte, error) {
		return json.Marshal(e)
	}
}

func (l *Logger) output(ctx context.Context, sev Severity, msg string, options []Option) {
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

	s, err := l.outputFormatter(&e, msg)
	if err == nil {
		if len(s) == 0 || s[len(s)-1] != '\n' {
			s = append(s, '\n')
		}
		l.out.Write(s)
	}
}

func SetOutput(w io.Writer) {
	std.mu.Lock()
	defer std.mu.Unlock()
	std.out = w
}

func SetTime(time Time) {
	std.mu.Lock()
	defer std.mu.Unlock()
	std.time = time
}

func SetOutputFormatter(f OutputFormatterFunc) {
	std.mu.Lock()
	defer std.mu.Unlock()
	std.outputFormatter = f
}

func RegisterHook(h Hook) {
	std.mu.Lock()
	defer std.mu.Unlock()
	std.hooks = append(std.hooks, h)
}

func Debug(ctx context.Context, v ...interface{}) {
	std.output(ctx, SeverityDebug, fmt.Sprint(v...), nil)
}

func Info(ctx context.Context, v ...interface{}) {
	std.output(ctx, SeverityInfo, fmt.Sprint(v...), nil)
}

func Error(ctx context.Context, v ...interface{}) {
	std.output(ctx, SeverityError, fmt.Sprint(v...), nil)
}

func Debugf(ctx context.Context, format string, v ...interface{}) {
	std.output(ctx, SeverityDebug, fmt.Sprintf(format, v...), nil)
}

func Infof(ctx context.Context, format string, v ...interface{}) {
	std.output(ctx, SeverityInfo, fmt.Sprintf(format, v...), nil)
}

func Errorf(ctx context.Context, format string, v ...interface{}) {
	std.output(ctx, SeverityError, fmt.Sprintf(format, v...), nil)
}
