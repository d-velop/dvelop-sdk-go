package structuredlog_test

import (
	"bytes"
	"context"
	log "github.com/d-velop/dvelop-sdk-go/structuredlog"
	"testing"
	"time"
)

type outputRecorder struct {
	*bytes.Buffer
	t *testing.T
}

func newOutputRecorder(t *testing.T) *outputRecorder {
	return &outputRecorder{&bytes.Buffer{}, t}
}

func (o *outputRecorder) OutputShouldBe(expected string) {
	actual := o.String()
	if actual != expected {
		o.t.Errorf("\ngot   :'%v'\nwanted:'%v'", actual, expected)
	}
}

func initializeLogger(t *testing.T) *outputRecorder {
	rec := newOutputRecorder(t)
	log.SetWriter(rec)
	log.SetClock(func() time.Time {
		return time.Date(2022, time.January, 01, 1, 2, 3, 4, time.UTC)
	})
	return rec
}

func TestSimpleMessageRedirectedToBuffer_Debug_WritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.Debug(context.Background(), "Message")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":5,\"body\":\"Message\"}")
}

func TestSimpleMessageRedirectedToBuffer_Info_WritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.Info(context.Background(), "Message")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":9,\"body\":\"Message\"}")
}

func TestSimpleMessageRedirectedToBuffer_Error_WritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.Error(context.Background(), "Message")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":17,\"body\":\"Message\"}")
}

func TestMultiPartMessageRedirectedToBuffer_Debug_WritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.Debug(context.Background(), "This", " is a ", "log ", "message")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":5,\"body\":\"This is a log message\"}")
}

func TestMultiPartMessageRedirectedToBuffer_Info_WritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.Info(context.Background(), "This", " is a ", "log ", "message")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":9,\"body\":\"This is a log message\"}")
}

func TestMultiPartMessageRedirectedToBuffer_Error_WritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.Error(context.Background(), "This", " is a ", "log ", "message")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":17,\"body\":\"This is a log message\"}")
}

func TestFormattedMessageRedirectedToBuffer_Debug_WritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.Debugf(context.Background(), "This is a %v log message", "formatted")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":5,\"body\":\"This is a formatted log message\"}")
}

func TestFormattedMessageRedirectedToBuffer_Info_WritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.Infof(context.Background(), "This is a %v log message", "formatted")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":9,\"body\":\"This is a formatted log message\"}")
}

func TestFormattedMessageRedirectedToBuffer_Error_WritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.Errorf(context.Background(), "This is a %v log message", "formatted")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":17,\"body\":\"This is a formatted log message\"}")
}
