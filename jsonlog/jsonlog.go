package jsonlog

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type Logger struct {
	mu           sync.Mutex
	out          io.Writer
	writeMessage Writer
	time         Time
	hooks        []Hook
}

var StdLogger = New(os.Stdout)

func New(out io.Writer) *Logger {
	return &Logger{
		out:  out,
		time: time.Now,
		writeMessage: func(e *Event, msg string) ([]byte, error) {
			json, err := e.MarshalJSON()
			if err != nil {
				return nil, err
			}
			json = append(json, '\n')
			return json, nil
		}}
}

func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out = w
}

func (l *Logger) SetWriteMessage(writeMsgFunc Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.writeMessage = writeMsgFunc
}

func (l *Logger) SetTime(time Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.time = time
}

func (l *Logger) RegisterHook(h Hook) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.hooks = append(l.hooks, h)
}

func (l *Logger) Print(ctx context.Context, sev Severity, v ...interface{}) {

	var o []LogOption
	for i := len(v); i > 0; i-- {
		d, e := v[len(v)-1].(LogOption)
		if e {
			v = v[:len(v)-1]
			o = append(o, d)
		}
	}

	l.writeOutput(ctx, sev, fmt.Sprint(v...), o)
}

func (l *Logger) Printf(ctx context.Context, sev Severity, format string, v ...interface{}) {

	var o []LogOption
	for i := len(v); i > 0; i-- {
		d, e := v[len(v)-1].(LogOption)
		if e {
			v = v[:len(v)-1]
			o = append(o, d)
		}
	}

	l.writeOutput(ctx, sev, fmt.Sprintf(format, v...), o)
}

// ****************************

//Nächste Schritte:
//	Vermutlich macht es Sinn, ein neues Log Package zu bauen, welches die JSON Events definiert (Serialisierung/Deserialsisierung)
//	und ausschließlich nach FTS protokolliert. Wenn ich das bestehende erweitere, dann wird das intern zu unübersichtlich und
//	macht das Ding auch zu vieles gleichzeitig. Dennoch kann man das Package als Drop in Replacement bauen. D.h. der Konsument muss
//	lediglich den Import und die Konfiguration des Loggings ändern. Die Logstatements selbst könnten gleich oder zumindest ähnlich bleibem.
//	import log "github.com/d-velop/dvelop-sdk-go/structuredlog"
//	func (l *Logger) SetOutput(w io.Writer) bleibt, um den Output umzulenken
//	func (l *Logger) SetWriteMessage(writeMsgFunc ...func(ctx context.Context, buf []byte, message string) []byte) wird zu
//	func (l *Logger) TransformEvent(transformEventFunc ...func(ctx context.Context, e Event)) da SetWriteMessage sowieso nicht in der go log spec drin ist
//
//type Logdata struct {
//	Name       string
//	Visibility int
//	Attributes *log.Attributes
//}
//
//func Test_INTERFACE_PROTOTYPE(t *testing.T) {
//		logdata := Logdata{Name: "VacationRequested", Visibility: 1, Attributes: &log.Attributes{
//			Http: &log.Http{
//				Method: "Get",
//			},
//		}}
//		log.Error(context.Background(), "Message")
//		log.Error(context.Background(), "Message", logdata)
//		log.Errorf(context.Background(), "Message %v", 1, logdata)
//	}

//	Time       	-> writeOutput
//	Severity  	-> Logfunction
//	Name
//	Body
//	TenantId  	-> Hook (Tenant-Middleware)
//	TraceId     -> Hook (?-Middleware)
//	SpanId    	-> Hook
//	Resource   	-> Hook
//	Attributes
//		HTTP		-> Hook ?
//		DB
//		...
//	Visibility

// performance ??

// ****************************

func (l *Logger) writeOutput(ctx context.Context, sev Severity, msg string, options []LogOption) {

	l.mu.Lock()
	defer l.mu.Unlock()

	currentTime := l.time()
	e := Event{
		Time:     &currentTime,
		Severity: sev,
		Body:     msg,
	}

	for _, h := range l.hooks {
		h(ctx, &e)
	}

	for _, o := range options {
		o(&e)
	}

	o, err := l.writeMessage(&e, msg)
	if err == nil {
		l.out.Write(o)
	}
}

func Debug(ctx context.Context, v ...interface{}) {
	StdLogger.Print(ctx, SeverityDebug, v...)
}

func Debugf(ctx context.Context, format string, v ...interface{}) {
	StdLogger.Printf(ctx, SeverityDebug, format, v...)
}

func Info(ctx context.Context, v ...interface{}) {
	StdLogger.Print(ctx, SeverityInfo, v...)
}

func Infof(ctx context.Context, format string, v ...interface{}) {
	StdLogger.Printf(ctx, SeverityInfo, format, v...)
}

func Error(ctx context.Context, v ...interface{}) {
	StdLogger.Print(ctx, SeverityError, v...)
}

func Errorf(ctx context.Context, format string, v ...interface{}) {
	StdLogger.Printf(ctx, SeverityError, format, v...)
}
