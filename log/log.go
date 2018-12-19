// Package log implements a simple replacement for the standard go log package.
// It provides 3 predefined loggers DEBUG, INFO and ERROR and supports the standard go context.Context
// to include context specific information in every logstatement
package log

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type Logger struct {
	mu           sync.Mutex // ensures atomic writes; protects the following fields
	out          io.Writer
	buf          []byte // for accumulating text to write
	writeMessage []func(ctx context.Context, buf []byte, message string) []byte
}

// New creates an new Logger.
// The out variable determines the destination for logstatements.
// The writeMessage variable determines callback functions which are invoked in the given order when writing the logstatement.
func New(out io.Writer, writeMessage ...func(ctx context.Context, buf []byte, message string) []byte) *Logger {
	return &Logger{out: out, writeMessage: writeMessage}
}

// SetOuput sets the output destination for the logger
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out = w
}

// SetWriteMessage set callback functions which are invoked in the given oder when this logger writes a logstatement
func (l *Logger) SetWriteMessage(writeMsgFunc ...func(ctx context.Context, buf []byte, message string) []byte) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.writeMessage = writeMsgFunc
}

// Print writes v to the log.
// Arguments are handled in the same manner as fmt.Print.
func (l *Logger) Print(ctx context.Context, v ...interface{}) {
	l.writeOutput(ctx, fmt.Sprint(v...))
}

// Print writes v to the log.
// Arguments are handled in the same manner as fmt.Printf.
func (l *Logger) Printf(ctx context.Context, format string, v ...interface{}) {
	l.writeOutput(ctx, fmt.Sprintf(format, v...))
}

func (l *Logger) writeOutput(ctx context.Context, message string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.buf = l.buf[:0]
	for _, f := range l.writeMessage {
		l.buf = f(ctx, l.buf, message)
	}

	if len(l.buf) == 0 || l.buf[len(l.buf)-1] != '\n' {
		l.buf = append(l.buf, '\n')
	}

	_, _ = l.out.Write(l.buf)
}

// StdDebug is the standard logger for debug messages
var StdDebug = New(os.Stderr, newWriteMessageFunc("DEBUG"))

// StdInfo is the standard logger for info messages
var StdInfo = New(os.Stderr, newWriteMessageFunc("INFO"))

// StdError is the standard logger for error messages
var StdError = New(os.Stderr, newWriteMessageFunc("ERROR"))

func newWriteMessageFunc(severity string) func(ctx context.Context, buf []byte, message string) []byte {
	return func(ctx context.Context, buf []byte, message string) []byte {
		buf = append(buf, time.Now().UTC().Format(time.RFC3339)...)
		buf = append(buf, ' ')
		buf = append(buf, severity...)
		buf = append(buf, ' ')
		buf = append(buf, message...)
		return buf
	}
}

// Debug is equivalent to log.StdDebug.Print()
func Debug(ctx context.Context, v ...interface{}) {
	StdDebug.Print(ctx, v...)
}

// Debug is equivalent to log.StdDebug.Printf()
func Debugf(ctx context.Context, format string, v ...interface{}) {
	StdDebug.Printf(ctx, format, v...)
}

// Info is equivalent to log.StdInfo.Print()
func Info(ctx context.Context, v ...interface{}) {
	StdInfo.Print(ctx, v...)
}

// Info is equivalent to log.StdInfo.Printf()
func Infof(ctx context.Context, format string, v ...interface{}) {
	StdInfo.Printf(ctx, format, v...)
}

// Error is equivalent to log.StdError.Print()
func Error(ctx context.Context, v ...interface{}) {
	StdError.Print(ctx, v...)
}

// Error is equivalent to log.StdError.Printf()
func Errorf(ctx context.Context, format string, v ...interface{}) {
	StdError.Printf(ctx, format, v...)
}
