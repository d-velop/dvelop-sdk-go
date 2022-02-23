package structuredlog

import (
	"context"
	"encoding/json"
	"fmt"
)

func writeMessage(sev Severity, msg string) {

	c := clock()
	e := Event{
		Severity: sev,
		Body:     msg,
		Time:     &c,

		//Visibility: 1,
		//Attr:       map[string]interface{}{},
	}

	json, err := json.Marshal(e)
	if err == nil {
		out.Write(json)
	}
}

func Debug(ctx context.Context, v ...interface{}) {
	writeMessage(SeverityDebug, fmt.Sprint(v...))
}

func Info(ctx context.Context, v ...interface{}) {
	writeMessage(SeverityInfo, fmt.Sprint(v...))
}

func Error(ctx context.Context, v ...interface{}) {
	writeMessage(SeverityError, fmt.Sprint(v...))
}

func Debugf(ctx context.Context, format string, v ...interface{}) {
	writeMessage(SeverityDebug, fmt.Sprintf(format, v...))
}

func Infof(ctx context.Context, format string, v ...interface{}) {
	writeMessage(SeverityInfo, fmt.Sprintf(format, v...))
}

func Errorf(ctx context.Context, format string, v ...interface{}) {
	writeMessage(SeverityError, fmt.Sprintf(format, v...))
}
