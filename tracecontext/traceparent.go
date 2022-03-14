package tracecontext

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
)

// TraceParent defines the header used for distributed tracing. A Traceparent is
// based on the W3C trace-context specification available at
// https://w3c.github.io/trace-context/#version.
type TraceParent struct {
	version    byte
	traceId    [16]byte
	parentId   [8]byte
	traceFlags byte
}

// TraceId returns the string representation of the trace-id.
func (t *TraceParent) TraceId() string {
	return hex.EncodeToString(t.traceId[:])
}

// ParentId returns the string representation of the parent-id.
func (t *TraceParent) ParentId() string {
	return hex.EncodeToString(t.parentId[:])
}

// String returns the string representation of the traceparent.
func (t *TraceParent) String() string {
	return hex.EncodeToString([]byte{t.version}) + "-" +
		hex.EncodeToString(t.traceId[:]) + "-" +
		hex.EncodeToString(t.parentId[:]) + "-" +
		hex.EncodeToString([]byte{t.traceFlags})
}

// NewTraceId generates a new trace-id and returns the string representation.
func NewTraceId() (string, error) {
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		return "", err
	}
	return hex.EncodeToString(id[:]), nil
}

// NewSpanId generates a new span-id and returns the string representation.
func NewSpanId() (string, error) {
	id := make([]byte, 8)
	if _, err := rand.Read(id); err != nil {
		return "", err
	}
	return hex.EncodeToString(id[:]), nil
}

// NewTraceParent creates a new TraceParent with given trace-id and parent-id.
func NewTraceParent(traceId string, parentId string) (*TraceParent, error) {
	t := &TraceParent{
		version:    0,
		traceFlags: 1,
	}

	ti, err := hex.DecodeString(traceId)
	if err != nil || len(ti) != 16 {
		return nil, errors.New("invalid trace-id")
	}

	pi, err := hex.DecodeString(parentId)
	if err != nil || len(pi) != 8 {
		return nil, errors.New("invalid parent-id")
	}

	copy(t.traceId[:], ti)
	copy(t.parentId[:], pi)
	return t, nil
}

// ParseTraceParent parses a traceparent string and returns a TraceParent.
func ParseTraceParent(traceParent string) (*TraceParent, error) {
	parts := strings.Split(traceParent, "-")

	if len(parts) != 4 {
		return nil, errors.New("missing parts")
	}

	t, err := NewTraceParent(parts[1], parts[2])
	if err != nil {
		return nil, err
	}

	version, err := hex.DecodeString(parts[0])
	if err != nil || len(version) != 1 {
		return nil, errors.New("invalid version")
	}

	traceFlags, err := hex.DecodeString(parts[3])
	if err != nil || len(traceFlags) != 1 {
		return nil, errors.New("invalid trace-flags")
	}

	t.version = version[0]
	t.traceFlags = traceFlags[0]
	return t, nil
}
