package syslog_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/d-velop/dvelop-sdk-go/log/syslog"
)

func TestWriteHeaderInfoFunc_WritesAppnameAndInfoSeverity(t *testing.T) {
	f := syslog.NewWriteHeaderFunc("myapp", syslog.INFO)

	var buf []byte
	buf = f(context.Background(), buf, "message")

	hostname, _ := os.Hostname()
	// Priority value is calculated as follows
	// FACILITY * 8 + SEVERITY
	// FACILITY is 16 (local use 0)
	// SEVERITY is 6 (Info)
	// 16*8+6=134
	// https://tools.ietf.org/html/rfc5424#section-6.2.1
	regex := fmt.Sprintf(`<134>1 \d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z %v myapp %v - `, hostname, os.Getpid())
	r, _ := regexp.Compile(regex)
	actual := string(buf)
	if !r.MatchString(actual) {
		t.Errorf("\n'%v' doesn't match the pattern\n'%v'", actual, regex)
	}
}

func TestWriteHeaderDebugFunc_WritesAppnameAndDebugSeverity(t *testing.T) {
	f := syslog.NewWriteHeaderFunc("myapp", syslog.DEBUG)

	var buf []byte
	buf = f(context.Background(), buf, "message")

	hostname, _ := os.Hostname()
	// Priority value is calculated as follows
	// FACILITY * 8 + SEVERITY
	// FACILITY is 16 (local use 0)
	// SEVERITY is 7 (Debug)
	// 16*8+7=135
	// https://tools.ietf.org/html/rfc5424#section-6.2.1
	regex := fmt.Sprintf(`<135>1 \d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z %v myapp %v - `, hostname, os.Getpid())
	r, _ := regexp.Compile(regex)
	actual := string(buf)
	if !r.MatchString(actual) {
		t.Errorf("\n'%v' doesn't match the pattern\n'%v'", actual, regex)
	}
}

func TestWriteHeaderDebugFunc_WritesAppnameAndWarnSeverity(t *testing.T) {
	f := syslog.NewWriteHeaderFunc("myapp", syslog.WARN)

	var buf []byte
	buf = f(context.Background(), buf, "message")

	hostname, _ := os.Hostname()
	// Priority value is calculated as follows
	// FACILITY * 8 + SEVERITY
	// FACILITY is 16 (local use 0)
	// SEVERITY is 4 (Warn)
	// 16*8+4=132
	// https://tools.ietf.org/html/rfc5424#section-6.2.1
	regex := fmt.Sprintf(`<132>1 \d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z %v myapp %v - `, hostname, os.Getpid())
	r, _ := regexp.Compile(regex)
	actual := string(buf)
	if !r.MatchString(actual) {
		t.Errorf("\n'%v' doesn't match the pattern\n'%v'", actual, regex)
	}
}

func TestWriteHeaderDebugFunc_WritesAppnameAndErrorSeverity(t *testing.T) {
	f := syslog.NewWriteHeaderFunc("myapp", syslog.ERROR)

	var buf []byte
	buf = f(context.Background(), buf, "message")

	hostname, _ := os.Hostname()
	// Priority value is calculated as follows
	// FACILITY * 8 + SEVERITY
	// FACILITY is 16 (local use 0)
	// SEVERITY is 3 (Error)
	// 16*8+3=131
	// https://tools.ietf.org/html/rfc5424#section-6.2.1
	regex := fmt.Sprintf(`<131>1 \d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z %v myapp %v - `, hostname, os.Getpid())
	r, _ := regexp.Compile(regex)
	actual := string(buf)
	if !r.MatchString(actual) {
		t.Errorf("\n'%v' doesn't match the pattern\n'%v'", actual, regex)
	}
}
