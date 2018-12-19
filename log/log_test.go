package log_test

import (
	"bytes"
	"context"
	"github.com/d-velop/dvelop-sdk-go/log"
	"regexp"
	"testing"
)

func TestMessageEndsWithNewline_Print_WritesMessageWithNewline(t *testing.T) {
	rec := newOutputRecorder(t)
	l := log.New(rec, messageOnlyWriteMsgFunc)

	l.Print(context.Background(), "message\n")

	rec.OutputShouldBe("message\n")
}

func TestMessageIsEmpty_Print_AppendsNewline(t *testing.T) {
	rec := newOutputRecorder(t)
	l := log.New(rec, messageOnlyWriteMsgFunc)

	l.Print(context.Background(), "")

	rec.OutputShouldBe("\n")
}

func TestMessageWithoutNewline_Print_AppendsNewline(t *testing.T) {
	rec := newOutputRecorder(t)
	l := log.New(rec, messageOnlyWriteMsgFunc)

	l.Print(context.Background(), "Message")

	rec.OutputShouldBe("Message\n")
}

func TestSetOutputCalled_Print_WritesToNewOutput(t *testing.T) {
	rec := newOutputRecorder(t)
	l := log.New(rec, messageOnlyWriteMsgFunc)
	newRec := newOutputRecorder(t)
	l.SetOutput(newRec)

	l.Print(context.Background(), "Message")

	rec.OutputShouldBe("")
	newRec.OutputShouldBe("Message\n")
}

func TestSetWriteMessageCalled_Print_InvokesNewWriteMessage(t *testing.T) {
	rec := newOutputRecorder(t)
	l := log.New(rec, messageOnlyWriteMsgFunc)

	type contextKey string
	const ctxKey = contextKey("ctxKey")
	fn := func(ctx context.Context, buf []byte, message string) []byte {
		ctxVal, _ := ctx.Value(ctxKey).(string)
		buf = append(buf, "Hello "...)
		buf = append(buf, ctxVal...)
		buf = append(buf, ' ')
		buf = append(buf, message...)
		return buf
	}
	l.SetWriteMessage(fn)
	ctx := context.WithValue(context.Background(), ctxKey, "World")

	l.Print(ctx, "Message")

	rec.OutputShouldBe("Hello World Message\n")
}

func TestMessageWithFormat_Printf_WritesFormatedMessage(t *testing.T) {
	rec := newOutputRecorder(t)
	l := log.New(rec, messageOnlyWriteMsgFunc)

	l.Printf(context.Background(), "Hello %v !", "world")

	rec.OutputShouldBe("Hello world !\n")
}

func TestWriteMessageDefined_Print_InvokesWriteMessage(t *testing.T) {
	rec := newOutputRecorder(t)
	type contextKey string
	const ctxKey = contextKey("ctxKey")
	fn := func(ctx context.Context, buf []byte, message string) []byte {
		ctxVal, _ := ctx.Value(ctxKey).(string)
		buf = append(buf, "Hello "...)
		buf = append(buf, ctxVal...)
		buf = append(buf, ' ')
		buf = append(buf, message...)
		return buf
	}
	l := log.New(rec, fn)
	ctx := context.WithValue(context.Background(), ctxKey, "World")

	l.Print(ctx, "Message")

	rec.OutputShouldBe("Hello World Message\n")
}

func TestMultipleWriteMessageFuncsDefined_Print_InvokesWriteMessageFuncsInOrder(t *testing.T) {
	rec := newOutputRecorder(t)
	type contextKey string
	const ctxKey = contextKey("ctxKey")
	fn1 := func(ctx context.Context, buf []byte, message string) []byte {
		ctxVal, _ := ctx.Value(ctxKey).(string)
		buf = append(buf, message...)
		buf = append(buf, ' ')
		buf = append(buf, ctxVal...)
		buf = append(buf, " writer 1"...)
		return buf
	}
	fn2 := func(ctx context.Context, buf []byte, message string) []byte {
		ctxVal, _ := ctx.Value(ctxKey).(string)
		buf = append(buf, message...)
		buf = append(buf, ' ')
		buf = append(buf, ctxVal...)
		buf = append(buf, " writer 2"...)
		return buf
	}
	l := log.New(rec, fn1, fn2)
	ctx := context.WithValue(context.Background(), ctxKey, "from")

	l.Print(ctx, "Hello")

	rec.OutputShouldBe("Hello from writer 1Hello from writer 2\n")
}

func TestStdInfoWriteMessageUnchanged_Debug_WritesMessageWithTimestampAndInfo(t *testing.T) {
	rec := newOutputRecorder(t)
	log.StdDebug.SetOutput(rec)

	log.Debug(context.Background(), "Message")

	r, _ := regexp.Compile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z DEBUG Message\n`)
	actual := rec.String()
	if !r.MatchString(actual) {
		t.Errorf("'%v' doesn't match the pattern '<RFC3339 Timestamp> DEBUG Message\n'", actual)
	}
}

func TestStdInfoWriteMessageUnchanged_Debugf_WritesMessageWithTimestampAndInfo(t *testing.T) {
	rec := newOutputRecorder(t)
	log.StdDebug.SetOutput(rec)

	log.Debugf(context.Background(), "Message %v", 1)

	r, _ := regexp.Compile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z DEBUG Message 1\n`)
	actual := rec.String()
	if !r.MatchString(actual) {
		t.Errorf("'%v' doesn't match the pattern '<RFC3339 Timestamp> DEBUG Message\n'", actual)
	}
}

func TestStdInfoWriteMessageUnchanged_Info_WritesMessageWithTimestampAndInfo(t *testing.T) {
	rec := newOutputRecorder(t)
	log.StdInfo.SetOutput(rec)

	log.Info(context.Background(), "Message")

	r, _ := regexp.Compile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z INFO Message\n`)
	actual := rec.String()
	if !r.MatchString(actual) {
		t.Errorf("'%v' doesn't match the pattern '<RFC3339 Timestamp> INFO Message\n'", actual)
	}
}

func TestStdInfoWriteMessageUnchanged_Infof_WritesMessageWithTimestampAndInfo(t *testing.T) {
	rec := newOutputRecorder(t)
	log.StdInfo.SetOutput(rec)

	log.Infof(context.Background(), "Message %v", 1)

	r, _ := regexp.Compile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z INFO Message 1\n`)
	actual := rec.String()
	if !r.MatchString(actual) {
		t.Errorf("'%v' doesn't match the pattern '<RFC3339 Timestamp> INFO Message\n'", actual)
	}
}

func TestStdInfoWriteMessageUnchanged_Error_WritesMessageWithTimestampAndInfo(t *testing.T) {
	rec := newOutputRecorder(t)
	log.StdError.SetOutput(rec)

	log.Error(context.Background(), "Message")

	r, _ := regexp.Compile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z ERROR Message\n`)
	actual := rec.String()
	if !r.MatchString(actual) {
		t.Errorf("'%v' doesn't match the pattern '<RFC3339 Timestamp> ERROR Message\n'", actual)
	}
}

func TestStdInfoWriteMessageUnchanged_Errorf_WritesMessageWithTimestampAndInfo(t *testing.T) {
	rec := newOutputRecorder(t)
	log.StdError.SetOutput(rec)

	log.Errorf(context.Background(), "Message %v", 1)

	r, _ := regexp.Compile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z ERROR Message 1\n`)
	actual := rec.String()
	if !r.MatchString(actual) {
		t.Errorf("'%v' doesn't match the pattern '<RFC3339 Timestamp> ERROR Message\n'", actual)
	}
}

func messageOnlyWriteMsgFunc(ctx context.Context, buf []byte, message string) []byte {
	buf = append(buf, message...)
	_ = ctx // avoid Unused parameter warning
	return buf
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
