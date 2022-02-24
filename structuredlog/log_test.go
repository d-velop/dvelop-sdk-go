package structuredlog_test

import (
	"bytes"
	"context"
	"fmt"
	log "github.com/d-velop/dvelop-sdk-go/structuredlog"
	"sync"
	"testing"
	"time"
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

func initializeLogger(t *testing.T) *outputRecorder {
	log.Default().Init()
	rec := &outputRecorder{&bytes.Buffer{}, t}
	log.SetOutput(rec)
	log.SetTime(func() time.Time {
		return time.Date(2022, time.January, 01, 1, 2, 3, 4, time.UTC)
	})
	return rec
}

func TestLogMessageWithSimpleString_Debug_WritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.Debug(context.Background(), "Log message")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":5,\"body\":\"Log message\"}\n")
}

func TestLogMessageWithSimpleString_Info_WritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.Info(context.Background(), "Log message")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":9,\"body\":\"Log message\"}\n")
}

func TestLogMessageWithSimpleString_Error_WritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.Error(context.Background(), "Log message")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":17,\"body\":\"Log message\"}\n")
}

func TestLogMessageWithMultipleStringParts_Debug_WritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.Debug(context.Background(), "Log ", "message")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":5,\"body\":\"Log message\"}\n")
}

func TestLogMessageWithMultipleStringParts_Info_WritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.Info(context.Background(), "Log ", "message")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":9,\"body\":\"Log message\"}\n")
}

func TestLogMessageWithMultipleStringParts_Error_WritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.Error(context.Background(), "Log ", "message")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":17,\"body\":\"Log message\"}\n")
}

func TestLogMessageIsFormatted_Debug_WritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.Debugf(context.Background(), "This is a %s log message", "formatted")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":5,\"body\":\"This is a formatted log message\"}\n")
}

func TestLogMessageIsFormatted_Info_WritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.Infof(context.Background(), "This is a %s log message", "formatted")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":9,\"body\":\"This is a formatted log message\"}\n")
}

func TestLogMessageIsFormatted_Error_WritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.Errorf(context.Background(), "This is a %s log message", "formatted")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":17,\"body\":\"This is a formatted log message\"}\n")
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

func TestLogMessageWithVisibilityIsFalse_Debug_AddVisPropertyAndWritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.WithVisibility(false).Debug(context.Background(), "Log message")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":5,\"body\":\"Log message\",\"vis\":0}\n")
}

func TestLogMessageWithVisibilityIsTrue_Debug_DoNotAddVisPropertyAndWritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.WithVisibility(true).Debug(context.Background(), "Log message")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":5,\"body\":\"Log message\"}\n")
}

func TestLogMessageWithVisibilityIsFalse_Info_AddVisPropertyAndWritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.WithVisibility(false).Info(context.Background(), "Log message")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":9,\"body\":\"Log message\",\"vis\":0}\n")
}

func TestLogMessageWithVisibilityIsFalse_Error_AddVisPropertyAndWritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.WithVisibility(false).Error(context.Background(), "Log message")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":17,\"body\":\"Log message\",\"vis\":0}\n")
}

func TestLogMessageWithVisibilityIsFalseAndMultipleStringParts_Debug_WritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.WithVisibility(false).Debug(context.Background(), "Log ", "message")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":5,\"body\":\"Log message\",\"vis\":0}\n")
}

func TestLogMessageWithVisibilityIsFalseAndMultipleStringParts_Info_WritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.WithVisibility(false).Info(context.Background(), "Log ", "message")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":9,\"body\":\"Log message\",\"vis\":0}\n")
}

func TestLogMessageWithVisibilityIsFalseAndMultipleStringParts_Error_WritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.WithVisibility(false).Error(context.Background(), "Log ", "message")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":17,\"body\":\"Log message\",\"vis\":0}\n")
}

func TestLogMessageWithVisibilityIsFalseAndIsFormatted_Debug_WritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.WithVisibility(false).Debugf(context.Background(), "This is a %s log message", "formatted")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":5,\"body\":\"This is a formatted log message\",\"vis\":0}\n")
}

func TestLogMessageWithVisibilityIsFalseAndIsFormatted_Info_WritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.WithVisibility(false).Infof(context.Background(), "This is a %s log message", "formatted")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":9,\"body\":\"This is a formatted log message\",\"vis\":0}\n")
}

func TestLogMessageWithVisibilityIsFalseAndIsFormatted_Error_WritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.WithVisibility(false).Errorf(context.Background(), "This is a %s log message", "formatted")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":17,\"body\":\"This is a formatted log message\",\"vis\":0}\n")
}

func TestLogMessageWithName_Info_AddNamePropertyAndWritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.WithName("Log message name").Info(context.Background(), "Log message")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":9,\"name\":\"Log message name\",\"body\":\"Log message\"}\n")
}

func TestLogMessageWithHttp_Info_AddHttpPropertyAndWritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.WithHttp(log.Http{Method: "Get"}).Info(context.Background(), "Log message")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":9,\"body\":\"Log message\",\"attr\":{\"http\":{\"method\":\"Get\"}}}\n")
}

func TestLogMessageWithDB_Info_AddDBPropertyAndWritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.WithDB(log.DB{Name: "CustomDb"}).Info(context.Background(), "Log message")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":9,\"body\":\"Log message\",\"attr\":{\"db\":{\"name\":\"CustomDb\"}}}\n")
}

func TestLogMessageWithException_Info_AddExceptionPropertyAndWritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.WithException(log.Exception{Type: "CustomLogException"}).Info(context.Background(), "Log message")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":9,\"body\":\"Log message\",\"attr\":{\"exception\":{\"type\":\"CustomLogException\"}}}\n")
}

func TestLogMessageWithEveryPossibleOption_Info_AddAllPropertiesAndWritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	log.WithName("Log message name").
		WithVisibility(false).
		WithHttp(log.Http{Method: "Get"}).
		WithDB(log.DB{Name: "CustomDb"}).
		WithException(log.Exception{Type: "CustomLogException"}).
		Info(context.Background(), "Log message")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":9,\"name\":\"Log message name\",\"body\":\"Log message\",\"attr\":{\"http\":{\"method\":\"Get\"},\"db\":{\"name\":\"CustomDb\"},\"exception\":{\"type\":\"CustomLogException\"}},\"vis\":0}\n")
}

func TestLogMessageWithRegisteredHookAndOtherService_Info_OverrideServicePropertyAndWritesJSONToBuffer(t *testing.T) {
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

	log.With(func(e *log.Event) {
		e.Resource = &log.Resource{
			Service: &log.Service{
				Name:     "OtherGoApplication",
				Version:  "2.0.0",
				Instance: "instanceId",
			},
		}
	}).Info(context.Background(), "Log message")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":9,\"body\":\"Log message\",\"res\":{\"svc\":{\"name\":\"OtherGoApplication\",\"ver\":\"2.0.0\",\"inst\":\"instanceId\"}}}\n")
}

func logAsync(wg *sync.WaitGroup) {
	defer wg.Done()
	log.Info(context.Background(), "Log message")
}

func TestSeveralLogMessagesAtTheSameTime_Info_WritesJSONToBuffer(t *testing.T) {
	rec := &outputRecorderSlow{initializeLogger(t)}
	log.SetOutput(rec)
	var wg sync.WaitGroup
	wg.Add(3)
	go logAsync(&wg)
	go logAsync(&wg)
	go logAsync(&wg)
	wg.Wait()
	rec.OutputShouldBe(fmt.Sprint("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":9,\"body\":\"Log message\"}\n",
		"{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":9,\"body\":\"Log message\"}\n",
		"{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":9,\"body\":\"Log message\"}\n"))
}
