package structuredlog_test

import (
	"context"
	log "github.com/d-velop/dvelop-sdk-go/structuredlog"
	"testing"
)

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