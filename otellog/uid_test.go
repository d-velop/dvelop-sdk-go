package otellog_test

import (
	"context"
	"errors"
	log "github.com/d-velop/dvelop-sdk-go/otellog"
	"testing"
)

func TestAddUidToLogEvent_givenUidFnReturnsUid_whenLogWritten_thenAddsUidToLog(t *testing.T) {

	uidFn := log.UidFromContextFn(func(ctx context.Context) (string, error) {
		return "a63554a8-6044-417b-a0aa-37a3b7d24e82", nil
	})

	hook := log.AddUidToLogEvent(uidFn)

	rec := initializeLogger(t)

	log.RegisterHook(hook)
	log.Info(context.Background(), "Testmessage")

	rec.OutputShouldBe(`{"time":"2022-01-01T01:02:03.000000004Z","sev":9,"body":"Testmessage","uid":"89757a7742548835532f1809558e0ab24eb39057966ad1630f1c493d94c3aec1"}` + "\n")
}

func TestAddUidToLogEvent_givenUidFnReturnsEmptyUid_whenLogWritten_thenAddsNothingToLog(t *testing.T) {

	uidFn := log.UidFromContextFn(func(ctx context.Context) (string, error) {
		return "", nil
	})

	hook := log.AddUidToLogEvent(uidFn)

	rec := initializeLogger(t)

	log.RegisterHook(hook)
	log.Info(context.Background(), "Testmessage")

	rec.OutputShouldBe(`{"time":"2022-01-01T01:02:03.000000004Z","sev":9,"body":"Testmessage"}` + "\n")
}

func TestAddUidToLogEvent_givenUidFnReturnsError_whenLogWritten_thenAddsNothingToLog(t *testing.T) {

	uidFn := log.UidFromContextFn(func(ctx context.Context) (string, error) {
		return "", errors.New("some error")
	})

	hook := log.AddUidToLogEvent(uidFn)

	rec := initializeLogger(t)

	log.RegisterHook(hook)
	log.Info(context.Background(), "Testmessage")

	rec.OutputShouldBe(`{"time":"2022-01-01T01:02:03.000000004Z","sev":9,"body":"Testmessage"}` + "\n")
}
