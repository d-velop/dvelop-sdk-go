package tracecontext_test

import (
	"github.com/d-velop/dvelop-sdk-go/tracecontext"
	"testing"
)

func TestValidTraceIdAndValidParentId_NewTraceparent_ReturnsTraceparentWithVersion0AndFlags1(t *testing.T) {
	tp, err := tracecontext.NewTraceparent("4bf92f3577b34da6a3ce929d0e0e4736", "00f067aa0ba902b7")

	if err != nil {
		t.Errorf("Traceparent cannot be created: %v", err)
	}

	assertString(t, tp.TraceId(), "4bf92f3577b34da6a3ce929d0e0e4736")
	assertString(t, tp.ParentId(), "00f067aa0ba902b7")
	assertString(t, tp.String(), "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
}

func TestInvalidTraceIdAndValidParentId_NewTraceparent_ReturnsTraceIdError(t *testing.T) {
	tp, err := tracecontext.NewTraceparent("00f067aa0ba902b7", "00f067aa0ba902b7")

	if err == nil {
		t.Errorf("Traceparent should throw an error")
	}

	if tp != nil {
		t.Errorf("Traceparent should be null")
	}

	assertString(t, err.Error(), "invalid trace-id")
}

func TestValidTraceIdAndInvalidParentId_NewTraceparent_ReturnsParentIdError(t *testing.T) {
	tp, err := tracecontext.NewTraceparent("4bf92f3577b34da6a3ce929d0e0e4736", "4bf92f3577b34da6a3ce929d0e0e4736")

	if err == nil {
		t.Errorf("Traceparent should throw an error")
	}

	if tp != nil {
		t.Errorf("Traceparent should be null")
	}

	assertString(t, err.Error(), "invalid parent-id")
}

func TestValidTraceparent_ParseTraceparent_ReturnsTraceparentWithVersion0AndFlags1(t *testing.T) {
	tp, err := tracecontext.ParseTraceparent("00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")

	if err != nil {
		t.Errorf("Traceparent cannot be created: %v", err)
	}

	assertString(t, tp.TraceId(), "4bf92f3577b34da6a3ce929d0e0e4736")
	assertString(t, tp.ParentId(), "00f067aa0ba902b7")
	assertString(t, tp.String(), "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
}

func TestTraceparentWithInvalidVersion_ParseTraceparent_ReturnsVersionError(t *testing.T) {
	tp, err := tracecontext.ParseTraceparent("0-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")

	if err == nil {
		t.Errorf("Traceparent should throw an error")
	}

	if tp != nil {
		t.Errorf("Traceparent should be null")
	}

	assertString(t, err.Error(), "invalid version")
}

func TestTraceparentWithInvalidTraceId_ParseTraceparent_ReturnsTraceIdError(t *testing.T) {
	tp, err := tracecontext.ParseTraceparent("00-00f067aa0ba902b7-00f067aa0ba902b7-01")

	if err == nil {
		t.Errorf("Traceparent should throw an error")
	}

	if tp != nil {
		t.Errorf("Traceparent should be null")
	}

	assertString(t, err.Error(), "invalid trace-id")
}

func TestTraceparentWithInvalidParentId_ParseTraceparent_ReturnsParentIdError(t *testing.T) {
	tp, err := tracecontext.ParseTraceparent("00-4bf92f3577b34da6a3ce929d0e0e4736-4bf92f3577b34da6a3ce929d0e0e4736-01")

	if err == nil {
		t.Errorf("Traceparent should throw an error")
	}

	if tp != nil {
		t.Errorf("Traceparent should be null")
	}

	assertString(t, err.Error(), "invalid parent-id")
}

func TestTraceparentWithInvalidFlags_ParseTraceparent_ReturnsFlagsError(t *testing.T) {
	tp, err := tracecontext.ParseTraceparent("00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-0")

	if err == nil {
		t.Errorf("Traceparent should throw an error")
	}

	if tp != nil {
		t.Errorf("Traceparent should be null")
	}

	assertString(t, err.Error(), "invalid trace-flags")
}

func TestNewTraceId_ReturnsRandom16ByteString(t *testing.T) {
	traceId, err := tracecontext.NewTraceId()

	if err != nil {
		t.Errorf("TraceId cannot be generated: %v", err)
	}

	if traceId == "" || len(traceId) != 32 {
		t.Errorf("TraceId was generated incorrectly: %v", traceId)
	}
}

func TestNewSpanId_ReturnsRandom8ByteString(t *testing.T) {
	spanId, err := tracecontext.NewSpanId()

	if err != nil {
		t.Errorf("SpanId cannot be generated: %v", err)
	}

	if spanId == "" || len(spanId) != 16 {
		t.Errorf("SpanId was generated incorrectly: %v", spanId)
	}
}

func assertString(t *testing.T, actual string, expected string) {
	if actual != expected {
		t.Errorf("\ngot   :'%v'\nwanted:'%v'", actual, expected)
	}
}
