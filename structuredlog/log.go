package structuredlog

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

type Logger struct {
	//mu           sync.Mutex
	out io.Writer
	outputFormatter OutputFormatterFunc
	time  Time
	hooks []Hook
}

type Time func() time.Time

type Hook func(ctx context.Context, e *Event)

type OutputFormatterFunc func(e *Event, msg string) ([]byte, error)

func New(out io.Writer) *Logger {
	return &Logger{
		out:  out,
		time: time.Now,
		outputFormatter: func(e *Event, msg string) ([]byte, error) {
			return json.Marshal(e)
		},
	}
}

var std = New(os.Stdout)

func Default() *Logger {
	return std
}

func (l *Logger) Output(ctx context.Context, sev Severity, msg string) {

	t := l.time()
	e := Event{
		Time:     &t,
		Severity: sev,
		Body:     msg,
	}

	for _, h := range l.hooks {
		h(ctx, &e)
	}

	json, err := l.outputFormatter(&e, msg)
	if err == nil {
		l.out.Write(json)
	}
}

func SetOutput(w io.Writer) {
	//l.mu.Lock()
	//defer l.mu.Unlock()
	std.out = w
}

func SetTime(time Time) {
	//l.mu.Lock()
	//defer l.mu.Unlock()
	std.time = time
}

func SetOutputFormatter(f OutputFormatterFunc) {
	//l.mu.Lock()
	//defer l.mu.Unlock()
	std.outputFormatter = f
}

func RegisterHook(h Hook) {
	//l.mu.Lock()
	//defer l.mu.Unlock()
	std.hooks = append(std.hooks, h)
}

func Debug(ctx context.Context, v ...interface{}) {
	std.Output(ctx, SeverityDebug, fmt.Sprint(v...))
}

func Info(ctx context.Context, v ...interface{}) {
	std.Output(ctx, SeverityInfo, fmt.Sprint(v...))
}

func Error(ctx context.Context, v ...interface{}) {
	std.Output(ctx, SeverityError, fmt.Sprint(v...))
}

func Debugf(ctx context.Context, format string, v ...interface{}) {
	std.Output(ctx, SeverityDebug, fmt.Sprintf(format, v...))
}

func Infof(ctx context.Context, format string, v ...interface{}) {
	std.Output(ctx, SeverityInfo, fmt.Sprintf(format, v...))
}

func Errorf(ctx context.Context, format string, v ...interface{}) {
	std.Output(ctx, SeverityError, fmt.Sprintf(format, v...))
}
