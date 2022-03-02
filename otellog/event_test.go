package otellog_test

import (
	"encoding/json"
	log "github.com/d-velop/dvelop-sdk-go/otellog"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestEventWithStringBody_Marshal_BodyPropertyIsString(t *testing.T) {
	e := log.Event{Body: "A normal message"}

	b, _ := json.Marshal(e)
	actual := string(b)

	expected := `{"body":"A normal message"}`
	if actual != expected {
		t.Errorf("\ngot   :'%v'\nwanted:'%v'", actual, expected)
	}
}

func TestEventWithStructBody_Marshal_BodyPropertyIsJsonObject(t *testing.T) {
	body := struct {
		Message string `json:"message"`
		Count   int    `json:"count"`
	}{
		"A normal message",
		2,
	}
	e := log.Event{Body: body}

	b, _ := json.Marshal(e)
	actual := string(b)

	expected := `{"body":{"message":"A normal message","count":2}}`
	if actual != expected {
		t.Errorf("\ngot   :'%v'\nwanted:'%v'", actual, expected)
	}
}

func Int(v int) *int {
	return &v
}

func TestEventWithDefaultVisibility_Marshal_OmitsVisProperty(t *testing.T) {
	e := log.Event{
		Body:       "A message",
		Visibility: Int(1),
	}

	b, _ := json.Marshal(e)
	actual := string(b)

	expected := `{"body":"A message"}`
	if actual != expected {
		t.Errorf("\ngot   :'%v'\nwanted:'%v'", actual, expected)
	}
}

func TestEventWithNonUTCTime_Marshal_TimePropertyIsRFCFormattedUTCTimeAndOriginalEventTimeIsUnchanged(t *testing.T) {
	originalTime := time.Date(2021, 03, 03, 15, 22, 04, 1001001, time.FixedZone("UTC-2", -2*60*60))
	eventTime := originalTime
	e := log.Event{
		Time: &eventTime,
	}

	b, _ := json.Marshal(e)
	actual := string(b)

	expected := `{"time":"2021-03-03T17:22:04.001001001Z"}`
	if actual != expected {
		t.Errorf("\ngot   :'%v'\nwanted:'%v'", actual, expected)
	}
	if eventTime != originalTime {
		t.Errorf("\ngot   :'%v'\nwanted:'%v'", eventTime, originalTime)
	}
}

func TestJsonStringWithoutVisProperty_Unmarshal_ProducesEventWithDefaultVisibility(t *testing.T) {
	var e log.Event
	_ = json.Unmarshal([]byte(`{"body":"A message"}`), &e)
	if e.Body == nil {
		t.Errorf("Body has wrong value\ngot   :'%v'\nwanted:'%v'", e.Body, "A message")
	}
	if e.Body != "A message" {
		t.Errorf("Body has wrong value\ngot   :'%v'\nwanted:'%v'", e.Body, "A message")
	}
	if e.Visibility == nil {
		t.Errorf("Visibility has wrong value\ngot   :'%v'\nwanted:'%v'", e.Visibility, 1)
	}
	if *e.Visibility != 1 {
		t.Errorf("Visibility has wrong value\ngot   :'%v'\nwanted:'%v'", *e.Visibility, 1)
	}
}

func TestJsonStringWithDurationInMilliseconds_Unmarshal_ProducesCorrectDuration(t *testing.T) {
	var e log.Event
	_ = json.Unmarshal([]byte(`{
		"body":"A normal message",
	    "attr":{
    	    "http":{
				"server":{
					"duration":5000
				},
				"client":{
					"duration":2000
				}
			}
       	}
	}`), &e)
	if e.Attributes.Http.Server.Duration != time.Millisecond*5000 {
		t.Errorf("Duration has wrong value\ngot   :'%v'\nwanted:'%v'", e.Attributes.Http.Server.Duration, time.Millisecond*5000)
	}
	if e.Attributes.Http.Client.Duration != time.Millisecond*2000 {
		t.Errorf("Duration has wrong value\ngot   :'%v'\nwanted:'%v'", e.Attributes.Http.Client.Duration, time.Millisecond*2000)
	}
}

func TestEventWithAllProperties_Marshal_JsonObjectWithAllProperties(t *testing.T) {
	ti := time.Date(2021, 06, 04, 07, 22, 48, 0, time.UTC)
	e := log.Event{
		Time:     &ti,
		Severity: log.SeverityInfo,
		Name:     "VacationRequested",
		Body:     "A normal message",
		TenantId: "45f",
		TraceId:  "f4dbb3edd765f620",
		SpanId:   "14dbb3edd765f650",
		Resource: &log.Resource{
			Service: &log.Service{
				Name:     "vacationprocessapp",
				Version:  "2.0.0",
				Instance: "f23c72a0497e4b9d8ab786a28b6ab37e",
			},
		},
		Attributes: &log.Attributes{
			Http: &log.Http{
				Method:     http.MethodPost,
				StatusCode: http.StatusInternalServerError,
				URL:        "http://acme.server.invalid/vacationprocess/vacations/1",
				Target:     "/vacationprocess/vacations/1",
				Host:       "acme.server.invalid",
				Scheme:     "https",
				Route:      "/vacationprocess/vacations/:vacationId",
				UserAgent:  "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:72.0) Gecko/20100101 Firefox/72.0",
				ClientIP:   "192.0.2.4",
				Server: &log.Server{
					Duration: 5000 * time.Millisecond,
				},
				Client: &log.Client{
					Duration: 2000 * time.Millisecond,
				},
			},
			DB: &log.DB{
				Name:      "customerdb",
				Statement: "SELECT * FROM wuser_table; SET mykey 'WuValue'",
				Operation: "SELECT",
			},
			Exception: &log.Exception{
				Type:       "java.net.ConnectException",
				Message:    "Division by zero",
				Stacktrace: "a stacktrace",
			},
		},
		Visibility: Int(0),
	}

	b, _ := json.Marshal(e)
	actualJsonString := normalizeJsonString(string(b))

	expectedJsonString := normalizeJsonString(`{
		"time":"2021-06-04T07:22:48Z",
		"sev":9,
		"name":"VacationRequested",
		"body":"A normal message",
		"tn":"45f",
		"trace":"f4dbb3edd765f620",
		"span": "14dbb3edd765f650",
		"res": {   
	        "svc":{
    	        "name": "vacationprocessapp",
        	    "ver": "2.0.0",
            	"inst": "f23c72a0497e4b9d8ab786a28b6ab37e"
        	}
    	},
	    "attr":{
    	    "http":{
        	    "method": "POST",
            	"statusCode": 500,
				"url":"http://acme.server.invalid/vacationprocess/vacations/1",
            	"target": "/vacationprocess/vacations/1",
				"host":"acme.server.invalid",
            	"scheme": "https",
				"route":"/vacationprocess/vacations/:vacationId",
            	"userAgent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:72.0) Gecko/20100101 Firefox/72.0",
            	"clientIP": "192.0.2.4",
				"server":{
					"duration":5000
				},
				"client":{
					"duration":2000
				}
			},
			"db":{
				"name":"customerdb",
				"statement":"SELECT * FROM wuser_table; SET mykey 'WuValue'",
				"operation":"SELECT"
			},
			"exception":{
				"type":"java.net.ConnectException",
				"message":"Division by zero",
				"stacktrace":"a stacktrace"
			}
       	},
		"vis":0
	}`)

	if actualJsonString != expectedJsonString {
		t.Errorf("\ngot   :'%v'\nwanted:'%v'", actualJsonString, expectedJsonString)
	}
}

func normalizeJsonString(s string) string {
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "\t", "")
	return s
}
