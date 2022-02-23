package structuredlog_test

import (
	"bytes"
	"context"
	"fmt"
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
	log.SetOutput(rec)
	log.SetTime(func() time.Time {
		return time.Date(2022, time.January, 01, 1, 2, 3, 4, time.UTC)
	})
	return rec
}

func TestLogMessageWithSimpleString_Debug_WritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.Debug(context.Background(), "Log message")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":5,\"body\":\"Log message\"}")
}

func TestLogMessageWithSimpleString_Info_WritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.Info(context.Background(), "Log message")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":9,\"body\":\"Log message\"}")
}

func TestLogMessageWithSimpleString_Error_WritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.Error(context.Background(), "Log message")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":17,\"body\":\"Log message\"}")
}

func TestLogMessageWithMultipleStringParts_Debug_WritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.Debug(context.Background(), "Log ", "message")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":5,\"body\":\"Log message\"}")
}

func TestLogMessageWithMultipleStringParts_Info_WritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.Info(context.Background(), "Log ", "message")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":9,\"body\":\"Log message\"}")
}

func TestLogMessageWithMultipleStringParts_Error_WritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.Error(context.Background(), "Log ", "message")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":17,\"body\":\"Log message\"}")
}

func TestLogMessageIsFormatted_Debug_WritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.Debugf(context.Background(), "This is a %v log message", "formatted")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":5,\"body\":\"This is a formatted log message\"}")
}

func TestLogMessageIsFormatted_Info_WritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.Infof(context.Background(), "This is a %v log message", "formatted")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":9,\"body\":\"This is a formatted log message\"}")
}

func TestLogMessageIsFormatted_Error_WritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.Errorf(context.Background(), "This is a %v log message", "formatted")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":17,\"body\":\"This is a formatted log message\"}")
}

func TestLogMessageWithRegisteredHook_Output_AddServiceAndWritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)

	log.RegisterHook(func(ctx context.Context, e *log.Event) {
		e.Resource = &log.Resource{
			Service: &log.Service{
				Name:     "GoApplication",
				Version:  "1.0.0",
				Instance: "instanceId",
			},
		}
	})

	log.Default().Output(context.Background(), log.SeverityDebug, "Log message")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":5,\"body\":\"Log message\",\"res\":{\"svc\":{\"name\":\"GoApplication\",\"ver\":\"1.0.0\",\"inst\":\"instanceId\"}}}")
}

func TestLogMessageWithCustomOutputFormatter_Output_WritesCustomFormatToBuffer(t *testing.T) {
	rec := initializeLogger(t)

	log.SetOutputFormatter(func(e *log.Event, msg string) ([]byte, error) {
		return []byte(fmt.Sprintf("This is a %s with severity level %d.", msg, e.Severity)), nil
	})

	log.Default().Output(context.Background(), log.SeverityDebug, "Log message")
	rec.OutputShouldBe("This is a Log message with severity level 5.")
}
