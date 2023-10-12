// Package syslog provides an io.Writer for a syslog server and
// WriteMessageFunctions which produce RFC5424 (https://tools.ietf.org/html/rfc5424)
// compliant logmessages
package syslog

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"time"
)

// cf. https://tools.ietf.org/html/rfc5424#section-6.2.1

const DEBUG = 7
const INFO = 6
const WARN = 4
const ERROR = 3

// NewWriter creates a new syslogwriter which writes to the given endpoint
//
// For example
//	syslog.NewWriter (localhost:514)
// writes to local syslog server listening on port 514
func NewWriter(endpoint string) (io.Writer, error) {
	conn, err := net.Dial("udp4", endpoint)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// NewWriteHeaderFunc returns a function which writes a syslog compliant header to a buffer.
// Note: The header is terminated with a space
//
// For example
//	w := syslog.NewWriteHeaderFunc("myapp", syslog.INFO)
// assigns an Info writer to w.
func NewWriteHeaderFunc(appname string, severity int) func(ctx context.Context, buf []byte, message string) []byte {
	prival := fmt.Sprintf("<%v>", 16*8+severity) // Facility 16 = local use 0
	version := strconv.Itoa(1)
	pri_version_sp := prival + version + " "
	const msgid = "-"
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "UNKNOWN"
	}
	var procid = strconv.Itoa(os.Getpid())
	hostname_sp_appname_sp_procid_sp_msgid_sp := hostname + " " + appname + " " + procid + " " + msgid + " "

	return func(ctx context.Context, buf []byte, message string) []byte {
		buf = append(buf, pri_version_sp...)
		buf = append(buf, time.Now().UTC().Format(time.RFC3339)...)
		buf = append(buf, ' ')
		buf = append(buf, hostname_sp_appname_sp_procid_sp_msgid_sp...)
		return buf
	}
}
