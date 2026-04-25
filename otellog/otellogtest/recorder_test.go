package otellogtest_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/d-velop/dvelop-sdk-go/otellog"
	"github.com/d-velop/dvelop-sdk-go/otellog/otellogtest"
)

func TestLogRecorder_givenSomeLog_whenAsserting_thenShouldMatch(t *testing.T) {
	recorder := otellogtest.NewLogRecorder(t)
	otellog.Info(context.Background(), "foo")

	recorder.ShouldHaveLogged("foo")
}

func TestLogRecorder_givenSomeLog_whenAsserting_thenShouldMatchPartially(t *testing.T) {
	recorder := otellogtest.NewLogRecorder(t)
	otellog.Info(context.Background(), "foo bar")

	recorder.ShouldHaveLogged("foo")
}

func TestLogRecorder_givenSomeLog_whenAsserting_thenShouldMatchWithSeverity(t *testing.T) {
	recorder := otellogtest.NewLogRecorder(t)
	otellog.Info(context.Background(), "foo")

	recorder.ShouldHaveLogged(otellog.SeverityInfo, "foo")
}

func TestLogRecorder_givenFunctionMatcher_whenAsserting_thenShouldMatch(t *testing.T) {
	recorder := otellogtest.NewLogRecorder(t)
	otellog.Info(context.Background(), "test")

	recorder.ShouldHaveLogged(func(event otellog.Event) bool {
		return true
	})
}

func TestLogRecorder_givenFunctionMatcherNotMatching_whenAsserting_thenShouldNotMatch(t *testing.T) {
	ft := &fakeTestingT{}
	recorder := otellogtest.NewLogRecorder(ft)
	otellog.Info(context.Background(), "test")

	recorder.ShouldHaveLogged(func(event otellog.Event) bool {
		return false
	})

	if !ft.failed {
		t.Error("expected test to fail")
	}
}

func TestLogRecorder_givenAttributeMatcher_whenAsserting_thenShouldMatch(t *testing.T) {
	recorder := otellogtest.NewLogRecorder(t)
	otellog.WithAdditionalAttributes(map[string]any{"foo": "bar"}).Info(context.Background(), "test")

	recorder.ShouldHaveLogged("test", otellogtest.ContainsAttribute("foo", "bar"))
}

func TestLogRecorder_givenAttributeMatcherNotMatching_whenAsserting_thenShouldNotMatch(t *testing.T) {
	ft := &fakeTestingT{}
	recorder := otellogtest.NewLogRecorder(ft)
	otellog.WithAdditionalAttributes(map[string]any{"foo": "bar"}).Info(context.Background(), "test")

	recorder.ShouldHaveLogged("test", otellogtest.ContainsAttribute("hello", "world"))

	if !ft.failed {
		t.Error("expected test to fail")
	}
}

func TestLogRecorder_givenOnlySeverity_whenAsserting_thenShouldMatch(t *testing.T) {
	recorder := otellogtest.NewLogRecorder(t)
	otellog.Info(context.Background(), "foo")
	var severity otellog.Severity
	{
		severity = otellog.SeverityInfo
	}

	recorder.ShouldHaveLogged(severity)
}

func TestLogRecorder_givenSomeLog_whenAsserting_thenShouldNotMatch(t *testing.T) {
	ft := &fakeTestingT{}
	recorder := otellogtest.NewLogRecorder(ft)
	otellog.Info(context.Background(), "foo")

	recorder.ShouldHaveLogged("bar")
	if !ft.failed {
		t.Error("expected test to fail")
	}
}

func TestLogRecorder_givenSomeLog_whenAsserting_thenShouldNotMatchWithSeverity(t *testing.T) {
	ft := &fakeTestingT{}
	recorder := otellogtest.NewLogRecorder(ft)
	otellog.Info(context.Background(), "foo")
	var severity otellog.Severity
	{
		severity = otellog.SeverityError
	}

	recorder.ShouldHaveLogged(severity, "foo")
	if !ft.failed {
		t.Error("expected test to fail")
	}
}

func TestLogRecorder_givenSomeLog_whenAsserting_thenShouldNotMatchWithSeverityOnly(t *testing.T) {
	ft := &fakeTestingT{}
	recorder := otellogtest.NewLogRecorder(ft)
	otellog.Info(context.Background(), "foo")

	recorder.ShouldHaveLogged(otellog.SeverityError)
	if !ft.failed {
		t.Error("expected test to fail")
	}
}

func TestLogRecorder_givenSomeLog_whenAsserting_thenShouldMatchWithMultipleConditions(t *testing.T) {
	recorder := otellogtest.NewLogRecorder(t)
	otellog.Info(context.Background(), "foo bar")

	recorder.ShouldHaveLogged("foo", "bar")
}

func TestLogRecorder_givenSomeLog_whenAsserting_thenShouldNotMatchWithMultipleConditions(t *testing.T) {
	ft := &fakeTestingT{}
	recorder := otellogtest.NewLogRecorder(ft)
	otellog.Info(context.Background(), "foo bar")

	recorder.ShouldHaveLogged("foo", "baz")
	if !ft.failed {
		t.Error("expected test to fail")
	}
}

func TestLogRecorder_givenSomeLog_whenAsserting_thenShouldMatchWithStringerThing(t *testing.T) {
	recorder := otellogtest.NewLogRecorder(t)
	otellog.Info(context.Background(), thing{"foo"})

	recorder.ShouldHaveLogged(thing{"foo"})
}

func TestLogRecorder_givenSomeLog_whenAsserting_thenFailsWithCorrectMessage(t *testing.T) {
	ft := &fakeTestingT{}
	recorder := otellogtest.NewLogRecorder(ft)
	otellog.Info(context.Background(), "foo")

	recorder.ShouldHaveLogged("bar")
	if !ft.failed {
		t.Error("expected test to fail")
	}
	if ft.m != "no log found matching [bar]" {
		t.Errorf("expected error message to be 'no log found matching [bar]', but was '%v'", ft.m)
	}
}

func TestLogRecorder_givenSomeLog_whenAsserting_thenFailsWithCorrectMessageAndSeverity(t *testing.T) {
	ft := &fakeTestingT{}
	recorder := otellogtest.NewLogRecorder(ft)
	otellog.Info(context.Background(), "foo")

	recorder.ShouldHaveLogged(otellog.SeverityError, "bar")
	if !ft.failed {
		t.Error("expected test to fail")
	}
	if ft.m != "no log found matching [17 bar]" {
		t.Errorf("expected error message to be 'no log found matching [17 bar]', but was '%v'", ft.m)
	}
}

type thing struct {
	content string
}

func (t thing) String() string {
	return t.content
}

type fakeTestingT struct {
	failed bool
	m      string
}

func (f *fakeTestingT) Errorf(format string, args ...any) {
	f.failed = true
	f.m = fmt.Sprintf(format, args...)
}

func (f *fakeTestingT) Helper() {
}
