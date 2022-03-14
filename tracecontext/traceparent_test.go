package tracecontext_test

import (
	"github.com/d-velop/dvelop-sdk-go/tracecontext"
	"testing"
)

func TestValidTraceIdAndValidParentId_NewTraceParent_ReturnsTraceParentWithVersion0AndFlags1(t *testing.T) {
	tp, err := tracecontext.NewTraceParent("4bf92f3577b34da6a3ce929d0e0e4736", "00f067aa0ba902b7")

	if err != nil {
		t.Errorf("TraceParent cannot be created: %v", err)
	}

	assertString(t, tp.TraceId(), "4bf92f3577b34da6a3ce929d0e0e4736")
	assertString(t, tp.ParentId(), "00f067aa0ba902b7")
	assertString(t, tp.String(), "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
}

func TestInvalidTraceIdAndValidParentId_NewTraceParent_ReturnsTraceIdError(t *testing.T) {
	tp, err := tracecontext.NewTraceParent("00f067aa0ba902b7", "00f067aa0ba902b7")

	if err == nil {
		t.Errorf("TraceParent should throw an error")
	}

	if tp != nil {
		t.Errorf("TraceParent should be null")
	}

	assertString(t, err.Error(), "invalid trace-id")
}

func TestValidTraceIdAndInvalidParentId_NewTraceParent_ReturnsParentIdError(t *testing.T) {
	tp, err := tracecontext.NewTraceParent("4bf92f3577b34da6a3ce929d0e0e4736", "4bf92f3577b34da6a3ce929d0e0e4736")

	if err == nil {
		t.Errorf("TraceParent should throw an error")
	}

	if tp != nil {
		t.Errorf("TraceParent should be null")
	}

	assertString(t, err.Error(), "invalid parent-id")
}

func TestValidTraceParent_ParseTraceParent_ReturnsTraceParentWithVersion0AndFlags1(t *testing.T) {
	tp, err := tracecontext.ParseTraceParent("00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")

	if err != nil {
		t.Errorf("TraceParent cannot be created: %v", err)
	}

	assertString(t, tp.TraceId(), "4bf92f3577b34da6a3ce929d0e0e4736")
	assertString(t, tp.ParentId(), "00f067aa0ba902b7")
	assertString(t, tp.String(), "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
}

func TestTraceParentWithInvalidVersion_ParseTraceParent_ReturnsVersionError(t *testing.T) {
	tp, err := tracecontext.ParseTraceParent("0-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")

	if err == nil {
		t.Errorf("TraceParent should throw an error")
	}

	if tp != nil {
		t.Errorf("TraceParent should be null")
	}

	assertString(t, err.Error(), "invalid version")
}

func TestTraceParentWithInvalidTraceId_ParseTraceParent_ReturnsTraceIdError(t *testing.T) {
	tp, err := tracecontext.ParseTraceParent("00-00f067aa0ba902b7-00f067aa0ba902b7-01")

	if err == nil {
		t.Errorf("TraceParent should throw an error")
	}

	if tp != nil {
		t.Errorf("TraceParent should be null")
	}

	assertString(t, err.Error(), "invalid trace-id")
}

func TestTraceParentWithInvalidParentId_ParseTraceParent_ReturnsParentIdError(t *testing.T) {
	tp, err := tracecontext.ParseTraceParent("00-4bf92f3577b34da6a3ce929d0e0e4736-4bf92f3577b34da6a3ce929d0e0e4736-01")

	if err == nil {
		t.Errorf("TraceParent should throw an error")
	}

	if tp != nil {
		t.Errorf("TraceParent should be null")
	}

	assertString(t, err.Error(), "invalid parent-id")
}

func TestTraceParentWithInvalidFlags_ParseTraceParent_ReturnsFlagsError(t *testing.T) {
	tp, err := tracecontext.ParseTraceParent("00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-0")

	if err == nil {
		t.Errorf("TraceParent should throw an error")
	}

	if tp != nil {
		t.Errorf("TraceParent should be null")
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
