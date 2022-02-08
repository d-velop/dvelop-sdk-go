package jsonlog_test

import (
	"bytes"
	"context"
	log "github.com/d-velop/dvelop-sdk-go/jsonlog"
	"testing"
	"time"
)

func TestLogger_Print(t *testing.T) {
	testcases := []struct {
		msg      string
		sev      log.Severity
		expected string
	}{
		{
			"Log debug",
			log.SeverityDebug,
			"{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":5,\"body\":\"Log debug\"}\n",
		},
		{
			"Log info",
			log.SeverityInfo,
			"{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":9,\"body\":\"Log info\"}\n",
		},
		{
			"Log error",
			log.SeverityError,
			"{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":17,\"body\":\"Log error\"}\n",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.msg, func(t *testing.T) {
			rec := newOutputRecorder(t)
			l := log.New(rec)
			l.SetTime(func() time.Time {
				return time.Date(2022, time.January, 01, 1, 2, 3, 4, time.UTC)
			})
			l.Print(context.Background(), tc.sev, tc.msg)
			rec.OutputShouldBe(tc.expected)
		})
	}
}

func TestLogger_Print_With_LogOptionName(t *testing.T) {
	rec := newOutputRecorder(t)
	l := log.New(rec)
	l.SetTime(func() time.Time {
		return time.Date(2022, time.January, 01, 1, 2, 3, 4, time.UTC)
	})

	l.Print(context.Background(), log.SeverityDebug, "Log message",
		log.WithName("CustomLogEvent"))

	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":5,\"name\":\"CustomLogEvent\",\"body\":\"Log message\"}\n")
}

func TestLogger_Print_With_LogOptionVisibility(t *testing.T) {
	rec := newOutputRecorder(t)
	l := log.New(rec)
	l.SetTime(func() time.Time {
		return time.Date(2022, time.January, 01, 1, 2, 3, 4, time.UTC)
	})

	l.Print(context.Background(), log.SeverityDebug, "Log message",
		log.WithVisibility(false))

	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":5,\"body\":\"Log message\",\"vis\":0}\n")
}

func TestLogger_Print_With_LogOptionHttp(t *testing.T) {
	rec := newOutputRecorder(t)
	l := log.New(rec)
	l.SetTime(func() time.Time {
		return time.Date(2022, time.January, 01, 1, 2, 3, 4, time.UTC)
	})

	l.Print(context.Background(), log.SeverityDebug, "Log message",
		log.WithHttp(log.Http{Method: "Get"}))

	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":5,\"body\":\"Log message\",\"attr\":{\"http\":{\"method\":\"Get\"}}}\n")
}

func TestLogger_Print_With_LogOptionException(t *testing.T) {
	rec := newOutputRecorder(t)
	l := log.New(rec)
	l.SetTime(func() time.Time {
		return time.Date(2022, time.January, 01, 1, 2, 3, 4, time.UTC)
	})

	l.Print(context.Background(), log.SeverityDebug, "Log message",
		log.WithException(log.Exception{Type: "CustomLogException"}))

	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":5,\"body\":\"Log message\",\"attr\":{\"exception\":{\"type\":\"CustomLogException\"}}}\n")
}

func TestLogger_Print_With_LogOptionDB(t *testing.T) {
	rec := newOutputRecorder(t)
	l := log.New(rec)
	l.SetTime(func() time.Time {
		return time.Date(2022, time.January, 01, 1, 2, 3, 4, time.UTC)
	})

	l.Print(context.Background(), log.SeverityDebug, "Log message",
		log.WithDB(log.DB{Name: "CustomDb"}))

	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":5,\"body\":\"Log message\",\"attr\":{\"db\":{\"name\":\"CustomDb\"}}}\n")
}

func TestLogger_Print_With_LogOptions(t *testing.T) {
	rec := newOutputRecorder(t)
	l := log.New(rec)
	l.SetTime(func() time.Time {
		return time.Date(2022, time.January, 01, 1, 2, 3, 4, time.UTC)
	})

	l.Print(context.Background(), log.SeverityDebug, "Log", " message",
		log.WithName("CustomLogEvent"),
		log.WithVisibility(false),
		log.WithHttp(log.Http{Method: "Get"}),
		log.WithException(log.Exception{Type: "CustomLogException"}),
		log.WithDB(log.DB{Name: "CustomDb"}))

	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":5,\"name\":\"CustomLogEvent\",\"body\":\"Log message\",\"attr\":{\"http\":{\"method\":\"Get\"},\"db\":{\"name\":\"CustomDb\"},\"exception\":{\"type\":\"CustomLogException\"}},\"vis\":0}\n")
}


func TestLogger_Printf(t *testing.T) {
	testcases := []struct {
		msg      string
		sev      log.Severity
		expected string
	}{
		{
			"Log %s debug",
			log.SeverityDebug,
			"{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":5,\"body\":\"Log format debug\"}\n",
		},
		{
			"Log %s info",
			log.SeverityInfo,
			"{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":9,\"body\":\"Log format info\"}\n",
		},
		{
			"Log %s error",
			log.SeverityError,
			"{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":17,\"body\":\"Log format error\"}\n",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.msg, func(t *testing.T) {
			rec := newOutputRecorder(t)
			l := log.New(rec)
			l.SetTime(func() time.Time {
				return time.Date(2022, time.January, 01, 1, 2, 3, 4, time.UTC)
			})
			l.Printf(context.Background(), tc.sev, tc.msg, "format")
			rec.OutputShouldBe(tc.expected)
		})
	}
}

func TestLogger_Printf_With_LogOptionName(t *testing.T) {
	rec := newOutputRecorder(t)
	l := log.New(rec)
	l.SetTime(func() time.Time {
		return time.Date(2022, time.January, 01, 1, 2, 3, 4, time.UTC)
	})

	l.Printf(context.Background(), log.SeverityDebug, "Log %s message", "format",
		log.WithName("CustomLogEvent"))

	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":5,\"name\":\"CustomLogEvent\",\"body\":\"Log format message\"}\n")
}

func TestLogger_Printf_With_LogOptionVisibility(t *testing.T) {
	rec := newOutputRecorder(t)
	l := log.New(rec)
	l.SetTime(func() time.Time {
		return time.Date(2022, time.January, 01, 1, 2, 3, 4, time.UTC)
	})

	l.Printf(context.Background(), log.SeverityDebug, "Log %s message", "format",
		log.WithVisibility(false))

	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":5,\"body\":\"Log format message\",\"vis\":0}\n")
}

func TestLogger_Printf_With_LogOptionHttp(t *testing.T) {
	rec := newOutputRecorder(t)
	l := log.New(rec)
	l.SetTime(func() time.Time {
		return time.Date(2022, time.January, 01, 1, 2, 3, 4, time.UTC)
	})

	l.Printf(context.Background(), log.SeverityDebug, "Log %s message", "format",
		log.WithHttp(log.Http{Method: "Get"}))

	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":5,\"body\":\"Log format message\",\"attr\":{\"http\":{\"method\":\"Get\"}}}\n")
}

func TestLogger_Printf_With_LogOptionException(t *testing.T) {
	rec := newOutputRecorder(t)
	l := log.New(rec)
	l.SetTime(func() time.Time {
		return time.Date(2022, time.January, 01, 1, 2, 3, 4, time.UTC)
	})

	l.Printf(context.Background(), log.SeverityDebug, "Log %s message", "format",
		log.WithException(log.Exception{Type: "CustomLogException"}))

	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":5,\"body\":\"Log format message\",\"attr\":{\"exception\":{\"type\":\"CustomLogException\"}}}\n")
}

func TestLogger_Printf_With_LogOptionDB(t *testing.T) {
	rec := newOutputRecorder(t)
	l := log.New(rec)
	l.SetTime(func() time.Time {
		return time.Date(2022, time.January, 01, 1, 2, 3, 4, time.UTC)
	})

	l.Printf(context.Background(), log.SeverityDebug, "Log %s message", "format",
		log.WithDB(log.DB{Name: "CustomDb"}))

	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":5,\"body\":\"Log format message\",\"attr\":{\"db\":{\"name\":\"CustomDb\"}}}\n")
}

func TestLogger_Printf_With_LogOptions(t *testing.T) {
	rec := newOutputRecorder(t)
	l := log.New(rec)
	l.SetTime(func() time.Time {
		return time.Date(2022, time.January, 01, 1, 2, 3, 4, time.UTC)
	})

	l.Printf(context.Background(), log.SeverityDebug, "Log %s message", "format",
		log.WithName("CustomLogEvent"),
		log.WithVisibility(false),
		log.WithHttp(log.Http{Method: "Get"}),
		log.WithException(log.Exception{Type: "CustomLogException"}),
		log.WithDB(log.DB{Name: "CustomDb"}))

	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":5,\"name\":\"CustomLogEvent\",\"body\":\"Log format message\",\"attr\":{\"http\":{\"method\":\"Get\"},\"db\":{\"name\":\"CustomDb\"},\"exception\":{\"type\":\"CustomLogException\"}},\"vis\":0}\n")
}

func TestLogger_RegisterHook(t *testing.T) {
	rec := newOutputRecorder(t)
	l := log.New(rec)
	l.SetTime(func() time.Time {
		return time.Date(2022, time.January, 01, 1, 2, 3, 4, time.UTC)
	})
	l.RegisterHook(func(ctx context.Context, event *log.Event) {
		event.TenantId = "tnId"
	})
	l.Print(context.Background(), log.SeverityDebug, "Log message")
	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":5,\"body\":\"Log message\",\"tn\":\"tnId\"}\n")
}

func TestLogger_SetWriteMessage(t *testing.T) {
	rec := newOutputRecorder(t)
	l := log.New(rec)
	l.SetTime(func() time.Time {
		return time.Date(2022, time.January, 01, 1, 2, 3, 4, time.UTC)
	})
	l.SetWriteMessage(func(event *log.Event) ([]byte, error) {
		return []byte("Use a custom writer"), nil
	})
	l.Print(context.Background(), log.SeverityDebug, "Log message")
	rec.OutputShouldBe("Use a custom writer")
}

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
