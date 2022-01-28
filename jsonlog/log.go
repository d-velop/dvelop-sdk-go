package jsonlog

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type Logger struct {
	mu  sync.Mutex
	out io.Writer
}

var StdLogger = New(os.Stdout)

func New(out io.Writer) *Logger {
	return &Logger{out: out}
}

func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out = w
}

func (l *Logger) Print(ctx context.Context, sev Severity, v ...interface{}) {

	l.writeOutput(ctx, sev, fmt.Sprint(v...))
}

func (l *Logger) Printf(ctx context.Context, sev Severity, format string, v ...interface{}) {
	l.writeOutput(ctx, sev, fmt.Sprintf(format, v...))
}

//	Time       	-> writeOutput
//	Severity  	-> Logfunction
//	Name
//	Body
//	TenantId  	-> Hook (Tenant-Middleware)
//	TraceId     	-> Hook (?-Middleware)
//	SpanId    	-> Hook
//	Resource   	-> Hook
//	Attributes
//		HTTP		-> Hook ?
//		DB
//		...
//	Visibility

func (l *Logger) writeOutput(ctx context.Context, sev Severity, msg string) {

	l.mu.Lock()
	defer l.mu.Unlock()

	currentTime := time.Now()
	e := Event{
		Time:     &currentTime,
		Severity: sev,
		Body:     msg,
		//Attributes:
	}

	for _, h := range hooks {
		h(ctx, &e)
	}

	json, err := e.MarshalJSON()
	if err == nil {
		json = append(json, '\n')
		l.out.Write(json)
	}
}

func Debug(ctx context.Context, v ...interface{}) {
	StdLogger.Print(ctx, SeverityDebug, v...)
}

func Debugf(ctx context.Context, format string, v ...interface{}) {
	StdLogger.Printf(ctx, SeverityDebug, format, v...)
}

func Info(ctx context.Context, v ...interface{}) {
	StdLogger.Print(ctx, SeverityInfo, v...)
}

func Infof(ctx context.Context, format string, v ...interface{}) {
	StdLogger.Printf(ctx, SeverityInfo, format, v...)
}

func Error(ctx context.Context, v ...interface{}) {
	StdLogger.Print(ctx, SeverityError, v...)
}

func Errorf(ctx context.Context, format string, v ...interface{}) {
	StdLogger.Printf(ctx, SeverityError, format, v...)
}
