package tracecontext

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
)

// Traceparent defines the header used for distributed tracing. A Traceparent is
// based on the W3C trace-context specification available at
// https://w3c.github.io/trace-context/#traceparent-header.
type Traceparent struct {
	version    byte
	traceId    [16]byte
	parentId   [8]byte
	traceFlags byte
}

// TraceId returns the string representation of the trace-id.
func (t *Traceparent) TraceId() string {
	return hex.EncodeToString(t.traceId[:])
}

// ParentId returns the string representation of the parent-id.
func (t *Traceparent) ParentId() string {
	return hex.EncodeToString(t.parentId[:])
}

// String returns the string representation of the traceparent.
func (t *Traceparent) String() string {
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

// NewTraceparent creates a new Traceparent with given trace-id and parent-id.
func NewTraceparent(traceId string, parentId string) (*Traceparent, error) {
	t := &Traceparent{
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

// ParseTraceparent parses a traceparent string and returns a Traceparent.
func ParseTraceparent(traceparent string) (*Traceparent, error) {
	parts := strings.Split(traceparent, "-")

	if len(parts) != 4 {
		return nil, errors.New("missing parts")
	}

	t, err := NewTraceparent(parts[1], parts[2])
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
