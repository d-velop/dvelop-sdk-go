package otellog_test

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	log "github.com/d-velop/dvelop-sdk-go/otellog"
)

type outputRecorder struct {
	*bytes.Buffer
	t *testing.T
}

type outputRecorderSlow struct {
	*outputRecorder
}

func (r *outputRecorderSlow) Write(p []byte) (n int, err error) {
	for _, b := range p {
		time.Sleep(5 * time.Millisecond)
		r.Buffer.Write([]byte{b})
	}
	return 0, err
}

func (o *outputRecorder) OutputShouldBe(expected string) {
	actual := o.String()
	if actual != expected {
		o.t.Errorf("\ngot   :'%v'\nwanted:'%v'", actual, expected)
	}
}

var severities = []struct {
	name  string
	level int
	log   func(ctx context.Context, v ...interface{})
	logf  func(ctx context.Context, format string, v ...interface{})
}{
	{"Info", 9, log.Info, log.Infof},
	{"Debug", 5, log.Debug, log.Debugf},
	{"Error", 17, log.Error, log.Errorf},
}

func initializeLogger(t *testing.T) *outputRecorder {
	log.Default().Reset()
	rec := &outputRecorder{&bytes.Buffer{}, t}
	log.SetOutput(rec)
	log.SetTime(func() time.Time {
		return time.Date(2022, time.January, 01, 1, 2, 3, 4, time.UTC)
	})
	return rec
}

func TestLogMessageWithSimpleString_SeverityLevel_WritesJSONToBuffer(t *testing.T) {
	for _, sev := range severities {
		t.Run(sev.name, func(t *testing.T) {
			rec := initializeLogger(t)

			sev.log(context.Background(), "Log message")

			rec.OutputShouldBe(fmt.Sprintf("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":%d,\"body\":\"Log message\"}\n", sev.level))
		})
	}
}

func TestLogMessageWithMultipleStringParts_SeverityLevel_WritesJSONToBuffer(t *testing.T) {
	for _, sev := range severities {
		t.Run(sev.name, func(t *testing.T) {
			rec := initializeLogger(t)

			sev.log(context.Background(), "Log ", "message")

			rec.OutputShouldBe(fmt.Sprintf("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":%d,\"body\":\"Log message\"}\n", sev.level))
		})
	}
}

func TestLogMessageIsFormatted_SeverityLevel_WritesJSONToBuffer(t *testing.T) {
	for _, sev := range severities {
		t.Run(sev.name, func(t *testing.T) {
			rec := initializeLogger(t)

			sev.logf(context.Background(), "This is a %s log message", "formatted")

			rec.OutputShouldBe(fmt.Sprintf("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":%d,\"body\":\"This is a formatted log message\"}\n", sev.level))
		})
	}
}

func TestLogMessageWithRegisteredHook_Info_AddServiceAndWritesJSONToBuffer(t *testing.T) {
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

	log.Info(context.Background(), "Log message")

	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":9,\"body\":\"Log message\",\"res\":{\"svc\":{\"name\":\"GoApplication\",\"ver\":\"1.0.0\",\"inst\":\"instanceId\"}}}\n")
}

func TestLogMessageWithCustomOutputFormatter_Info_WritesCustomFormatToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.SetOutputFormatter(func(e *log.Event, msg string) ([]byte, error) {
		return []byte(fmt.Sprintf("This is a %s with severity level %d.", msg, e.Severity)), nil
	})

	log.Info(context.Background(), "Log message")

	rec.OutputShouldBe("This is a Log message with severity level 9.\n")
}

func TestSeveralLogMessagesAtTheSameTime_Info_WritesJSONToBuffer(t *testing.T) {
	rec := &outputRecorderSlow{initializeLogger(t)}
	log.SetOutput(rec)

	var wg sync.WaitGroup
	wg.Add(3)
	for i := 0; i < 3; i++ {
		go func() {
			defer wg.Done()
			log.Info(context.Background(), "Log message")
		}()
	}
	wg.Wait()

	rec.OutputShouldBe(fmt.Sprint("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":9,\"body\":\"Log message\"}\n",
		"{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":9,\"body\":\"Log message\"}\n",
		"{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":9,\"body\":\"Log message\"}\n"))
}
