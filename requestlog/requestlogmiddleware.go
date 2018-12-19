// Package requestlog provides a http middleware which logs http requests.
//
// Information about the request and the response like http method, url, status code, elapsed time in ms and so on
// are logged as structured syslog data (cf. https://tools.ietf.org/html/rfc5424#section-6.3)
// which doesn't mean a syslog server has to be used. It's just a way of formatting structured data
// in strings. So any log destination which accepts strings can be used.
//
// http@49610 is used as syslog SD-ID (cf. https://tools.ietf.org/html/rfc5424#section-6.3.2).
// 49610 is the private enterprise number officially reserved for d.velop
// (cf. https://www.iana.org/assignments/enterprise-numbers/enterprise-numbers)
//
//
// Example:
//	func main() {
//		mux := http.NewServeMux()
//		mux.Handle("/hello", requestlog.Log(func(ctx context.Context, logmessage string) { log.Info(ctx, logmessage) })(helloHandler()))
//	}
//
//	func helloHandler() http.Handler {
//		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//			// ...
//		})
//	}
package requestlog

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// Log logs information about the request and response using the provided log function
func Log(log func(ctx context.Context, logmessage string)) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			start := time.Now()
			log(req.Context(), logBegin(req))
			lrw := newLogResponseWriter(rw)
			next.ServeHTTP(lrw, req)
			log(req.Context(), logEnd(req, lrw, time.Since(start)))
		})
	}
}

type logResponseWriter struct {
	http.ResponseWriter
	statusCode int
	header     http.Header
}

func newLogResponseWriter(rw http.ResponseWriter) *logResponseWriter {
	return &logResponseWriter{rw, http.StatusOK, nil}
}

func (lrw *logResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.header = lrw.ResponseWriter.Header()
	lrw.ResponseWriter.WriteHeader(code)
}

func logBegin(r *http.Request) string {
	return fmt.Sprintf("[http@49610 method=\"%v\" url=\"%v\"] BEGIN request %v", r.Method, r.URL.Path, logHeader(r.Header))
}

func logEnd(r *http.Request, lrw *logResponseWriter, t time.Duration) string {
	return fmt.Sprintf("[http@49610 method=\"%v\" url=\"%v\" millis=\"%d\" status=\"%v\"] END request %v", r.Method, r.URL.Path, int64(t/time.Millisecond), lrw.statusCode, logHeader(lrw.Header()))
}

var authSessionIdRegEx = regexp.MustCompile(`AuthSessionId=[^;\s]+`)
var authorizationHeaderValueRegEx = regexp.MustCompile(`(\S*) (\S*)`)

func logHeader(m map[string][]string) string {
	var buf []byte
	for key, value := range m {
		buf = append(buf, key...)
		buf = append(buf, ":"...)
		for _, v := range value {
			val := authSessionIdRegEx.ReplaceAll([]byte(fmt.Sprintf("%v", v)), []byte("AuthSessionId=***"))
			if strings.ToLower(key) == "authorization" {
				val = authorizationHeaderValueRegEx.ReplaceAll(val, []byte("$1 ***"))
			}
			buf = append(buf, val...)
		}
		buf = append(buf, ' ')
	}
	return string(buf)
}
