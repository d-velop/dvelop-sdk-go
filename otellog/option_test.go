package otellog_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	log "github.com/d-velop/dvelop-sdk-go/otellog"
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

func TestLogMessageWithVisibilityIsFalseAndStructAsBody_Debug_WritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	body := &struct {
		Id string `json:"id,omitempty"`
		Toggle bool `json:"toggle,omitempty"`
		Counter int `json:"counter,omitempty"`
	}{
		Id: "id",
		Toggle: true,
		Counter: 5,
	}

	log.WithVisibility(false).Debug(context.Background(), body)

	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":5,\"body\":{\"id\":\"id\",\"toggle\":true,\"counter\":5},\"vis\":0}\n")
}

func TestLogMessageWithVisibilityIsFalseAndStructAsBody_Info_WritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	body := &struct {
		Id string `json:"id,omitempty"`
		Toggle bool `json:"toggle,omitempty"`
		Counter int `json:"counter,omitempty"`
	}{
		Id: "id",
		Toggle: true,
		Counter: 5,
	}

	log.WithVisibility(false).Info(context.Background(), body)

	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":9,\"body\":{\"id\":\"id\",\"toggle\":true,\"counter\":5},\"vis\":0}\n")
}

func TestLogMessageWithVisibilityIsFalseAndStructAsBody_Error_WritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	body := &struct {
		Id string `json:"id,omitempty"`
		Toggle bool `json:"toggle,omitempty"`
		Counter int `json:"counter,omitempty"`
	}{
		Id: "id",
		Toggle: true,
		Counter: 5,
	}

	log.WithVisibility(false).Error(context.Background(), body)

	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":17,\"body\":{\"id\":\"id\",\"toggle\":true,\"counter\":5},\"vis\":0}\n")
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

func TestLogMessageWithHttpRequest_Info_AddHttpPropertyAndWritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	req := httptest.NewRequest("GET", "https://www.example.com/path?q=param", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")

	log.WithHttpRequest(req).Info(context.Background(), "Log message")

	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":9,\"body\":\"Log message\",\"attr\":{\"http\":{\"method\":\"GET\",\"url\":\"https://www.example.com/path?q=param\",\"target\":\"/path?q=param\",\"host\":\"www.example.com\",\"scheme\":\"https\",\"route\":\"/path\",\"userAgent\":\"Mozilla/5.0\",\"clientIP\":\"192.0.2.1:1234\"}}}\n")
}

func TestLogMessageWithHttpRequestAndUserInUrl_Info_AddHttpPropertyHideUserAndWritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	req := httptest.NewRequest("GET", "https://username:password@www.example.com/path?q=param", nil)

	log.WithHttpRequest(req).Info(context.Background(), "Log message")

	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":9,\"body\":\"Log message\",\"attr\":{\"http\":{\"method\":\"GET\",\"url\":\"https://www.example.com/path?q=param\",\"target\":\"/path?q=param\",\"host\":\"www.example.com\",\"scheme\":\"https\",\"route\":\"/path\",\"clientIP\":\"192.0.2.1:1234\"}}}\n")
}

func TestLogMessageWithHttpRequestAndForwardedHeader_Info_AddHttpPropertyAndWritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	req := httptest.NewRequest("GET", "https://www.example.com/path?q=param", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("X-Forwarded-For", "client, proxy1, proxy2")

	log.WithHttpRequest(req).Info(context.Background(), "Log message")

	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":9,\"body\":\"Log message\",\"attr\":{\"http\":{\"method\":\"GET\",\"url\":\"https://www.example.com/path?q=param\",\"target\":\"/path?q=param\",\"host\":\"www.example.com\",\"scheme\":\"https\",\"route\":\"/path\",\"userAgent\":\"Mozilla/5.0\",\"clientIP\":\"client\"}}}\n")
}

func TestLogMessageWithHttpResponse_AddHttpPropertyAndWritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	req := httptest.NewRequest("GET", "https://www.example.com/path?q=param", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp := httptest.NewRecorder().Result()
	resp.Request = req

	log.WithHttpResponse(resp).Info(context.Background(), "Log message")

	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":9,\"body\":\"Log message\",\"attr\":{\"http\":{\"method\":\"GET\",\"statusCode\":200,\"url\":\"https://www.example.com/path?q=param\",\"target\":\"/path?q=param\",\"host\":\"www.example.com\",\"scheme\":\"https\",\"route\":\"/path\",\"userAgent\":\"Mozilla/5.0\",\"clientIP\":\"192.0.2.1:1234\"}}}\n")
}

func TestLogMessageWithHttpResponseAndUserInUrl_Info_AddHttpPropertyHideUserAndWritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	req := httptest.NewRequest("GET", "https://username:password@www.example.com/", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp := httptest.NewRecorder().Result()
	resp.Request = req

	log.WithHttpResponse(resp).Info(context.Background(), "Log message")

	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":9,\"body\":\"Log message\",\"attr\":{\"http\":{\"method\":\"GET\",\"statusCode\":200,\"url\":\"https://www.example.com/\",\"target\":\"/\",\"host\":\"www.example.com\",\"scheme\":\"https\",\"route\":\"/\",\"userAgent\":\"Mozilla/5.0\",\"clientIP\":\"192.0.2.1:1234\"}}}\n")
}

func TestLogMessageWithHttpResponseAndForwardedHeader_Info_AddHttpPropertyAndWritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	req := httptest.NewRequest("GET", "https://www.example.com/path?q=param", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("X-Forwarded-For", "client, proxy1, proxy2")
	resp := httptest.NewRecorder().Result()
	resp.Request = req

	log.WithHttpResponse(resp).Info(context.Background(), "Log message")

	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":9,\"body\":\"Log message\",\"attr\":{\"http\":{\"method\":\"GET\",\"statusCode\":200,\"url\":\"https://www.example.com/path?q=param\",\"target\":\"/path?q=param\",\"host\":\"www.example.com\",\"scheme\":\"https\",\"route\":\"/path\",\"userAgent\":\"Mozilla/5.0\",\"clientIP\":\"client\"}}}\n")
}

func TestLogMessageWithHttpStatusCodeWithoutRequest_AddHttpPropertyAndWritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)

	log.WithHttpStatusCode(http.StatusMethodNotAllowed).Info(context.Background(), "Log message")

	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":9,\"body\":\"Log message\",\"attr\":{\"http\":{\"statusCode\":405}}}\n")
}

func TestLogMessageWithHttpStatusCodeWithRequest_AddHttpPropertyAndWritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	req := httptest.NewRequest("GET", "https://www.example.com/path?q=param", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")

	log.WithHttpStatusCode(http.StatusMethodNotAllowed).WithHttpRequest(req).Info(context.Background(), "Log message")

	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":9,\"body\":\"Log message\",\"attr\":{\"http\":{\"method\":\"GET\",\"statusCode\":405,\"url\":\"https://www.example.com/path?q=param\",\"target\":\"/path?q=param\",\"host\":\"www.example.com\",\"scheme\":\"https\",\"route\":\"/path\",\"userAgent\":\"Mozilla/5.0\",\"clientIP\":\"192.0.2.1:1234\"}}}\n")
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

func TestLogMessageWithAdditionalAttributes_Info_AddAdditionalAttributesPropertyAndWritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)
	type A struct {
		One int `json:"one"`
		Two int `json:"two"`
	}

	type B struct {
		One int `json:"one"`
		Two int `json:"two"`
	}
	type AdditionalAttributes struct {
		A A      `json:"a"`
		B B      `json:"b"`
		C string `json:"c"`
	}
	customAttr := AdditionalAttributes{
		A: A{One: 1, Two: 2},
		B: B{One: 1, Two: 2},
		C: "3",
	}

	log.WithAdditionalAttributes(customAttr).Info(context.Background(), "Log message")

	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":9,\"body\":\"Log message\",\"attr\":{\"a\":{\"one\":1,\"two\":2},\"b\":{\"one\":1,\"two\":2},\"c\":\"3\"}}\n")
}

func TestLogMessageWithEveryPossibleOption_Info_AddAllPropertiesAndWritesJSONToBuffer(t *testing.T) {
	rec := initializeLogger(t)

	type A struct {
		One int `json:"one"`
		Two int `json:"two"`
	}

	type B struct {
		One int `json:"one"`
		Two int `json:"two"`
	}
	type AdditionalAttributes struct {
		A A      `json:"a"`
		B B      `json:"b"`
		C string `json:"c"`
	}
	customAttr := AdditionalAttributes{
		A: A{One: 1, Two: 2},
		B: B{One: 1, Two: 2},
		C: "3",
	}

	log.WithName("Log message name").
		WithVisibility(false).
		WithHttp(log.Http{Method: "Get"}).
		WithDB(log.DB{Name: "CustomDb"}).
		WithException(log.Exception{Type: "CustomLogException"}).
		WithAdditionalAttributes(customAttr).
		Info(context.Background(), "Log message")

	rec.OutputShouldBe("{\"time\":\"2022-01-01T01:02:03.000000004Z\",\"sev\":9,\"name\":\"Log message name\",\"body\":\"Log message\",\"attr\":{\"a\":{\"one\":1,\"two\":2},\"b\":{\"one\":1,\"two\":2},\"c\":\"3\",\"db\":{\"name\":\"CustomDb\"},\"exception\":{\"type\":\"CustomLogException\"},\"http\":{\"method\":\"Get\"}},\"vis\":0}\n")
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
