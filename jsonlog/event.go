package jsonlog

import (
	"encoding/json"
	"time"
)

// An Event represents a structured logeevent inspired by the semantic model of OTEL (https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/logs/data-model.md)
type Event struct {
	Time       *time.Time  `json:"time,omitempty"`  // Time when the event occurred measured by the origin clock, normalized to UTC.
	Severity   Severity    `json:"sev,omitempty"`   // Numerical value of the severity cf. Severity constants like SeverityInfo for possible values and their semantics
	Name       string      `json:"name,omitempty"`  // Short event identifier that does not contain varying parts. Name describes what happened (e.g. "ProcessStarted"). Recommended to be no longer than 50 characters. Not guaranteed to be unique in any way. Typically used for filtering and grouping purposes in backends. Can be used to identify domain events like FeaturesRequested or UserLoggedIn (cf. example).
	Body       interface{} `json:"body,omitempty"`  // A value containing the body of the log record. Can be for example a human-readable string message (including multi-line) describing the event in a free form or it can be a structured data composed of arrays and maps of other values. Can vary for each occurrence of the event coming from the same source.
	TenantId   string      `json:"tn,omitempty"`    // ID of the tenant to which this event belongs.
	TraceId    string      `json:"trace,omitempty"` // Request trace-id as defined in W3C Trace Context (https://www.w3.org/TR/trace-context/#trace-id) specification. That is the ID of the whole trace forest used to uniquely identify a distributed trace through a system.
	SpanId     string      `json:"span,omitempty"`  // span-id. Can be set for logs that are part of a particular processing span. A span (https://opentracing.io/docs/overview/spans/) is the primary building block of a distributed trace, representing an individual unit of work done in a distributed system.
	Resource   *Resource   `json:"res,omitempty"`   // Describes the source of the log. Multiple occurrences of events coming from the same event source can happen across time and they all have the same value of Resource. Can contain for example information about the application that emits the record or about the infrastructure where the application runs.
	Attributes *Attributes `json:"attr,omitempty"`  // Additional information about the specific event occurrence. Unlike the Resource field, which is fixed for a particular source, Attributes can vary for each occurrence of the event coming from the same source. Can contain information about the request context (other than TraceId/SpanId).
	Visibility *int        `json:"vis,omitempty"`   // Specifies if the logstatement is visible for tenant owner / customer. For now possible values are 1: true 0: false	1 is the default value, that is statements are visible if not explicitly denied by setting this value to 0
}

// A Resource describes the source of the log. Multiple occurrences of events coming from the same event source can happen across time and they all have the same value of res. Can contain for example information about the application that emits the record or about the infrastructure where the application runs.
type Resource struct {
	Service *Service `json:"svc,omitempty"`
}

// A Service describes a service instance
type Service struct {
	Name     string `json:"name,omitempty"` // Logical name of the service. MUST be the same for all instances of horizontally scaled services. If necessary dots can be used to denote subcomponents like myservice.syncworker or myservice.scheduler.
	Version  string `json:"ver,omitempty"`  // The version string of the service API or implementation.
	Instance string `json:"inst,omitempty"` // The ID of the service instance. MUST be unique for each instance of the same service. The ID helps to distinguish instances of the same service that exist at the same time (e.g. instances of a horizontally scaled service). It is preferable for the ID to be persistent and stay the same for the lifetime of the service instance, however it is acceptable that the ID is ephemeral and changes during important lifetime events for the service (e.g. service restarts). If the service has no inherent unique ID that can be used as the value of this attribute it is recommended to generate a random Version 1 or Version 4 RFC 4122 UUID (services aiming for reproducible UUIDs may also use Version 5, see RFC 4122 for more recommendations).
}

// Attributes contains additional information about the specific event occurrence. Unlike the res field, which is fixed for a particular source, attr can vary for each occurrence of the event coming from the same source. Can contain information about the request context (other than TraceId/SpanId).
type Attributes struct {
	Http      *Http      `json:"http,omitempty"`      // Information about outbound or inbound http requests.
	DB        *DB        `json:"db,omitempty"`        // Information about outbound db requests.
	Exception *Exception `json:"exception,omitempty"` // Information about an exception
}

// Http contains information about outbound or inbound HTTP requests
type Http struct {
	Method     string  `json:"method,omitempty"`     // HTTP request method in upper case. For example GET, POST, DELETE
	StatusCode uint16  `json:"statusCode,omitempty"` // HTTP response status code
	URL        string  `json:"url,omitempty"`        // Full HTTP request URL in the form scheme://host[:port]/path?query[#fragment]. Usually the fragment is not transmitted over HTTP, but if it is known, it should be included nevertheless.	MUST NOT contain credentials passed via URL in form of https://username:password@www.example.com/. In such case the attribute's value should be https://www.example.com/.
	Target     string  `json:"target,omitempty"`     // The full request target as passed in a HTTP request line or equivalent. For example /path/12314/?q=ddds#123
	Host       string  `json:"host,omitempty"`       // The value of the HTTP host header. For example www.example.org
	Scheme     string  `json:"scheme,omitempty"`     // The URI scheme identifying the used protocol. For example http or https
	Route      string  `json:"route,omitempty"`      // The matched route (path template). For example /users/:userID?
	UserAgent  string  `json:"userAgent,omitempty"`  // Value of the HTTP User-Agent header sent by the client.
	ClientIP   string  `json:"clientIP,omitempty"`   // The IP address of the original client behind all proxies, if known (e.g. from X-Forwarded-For).
	Server     *Server `json:"server,omitempty"`     // Specific values for inbound HTTP requests
	Client     *Client `json:"client,omitempty"`     // Specific values for outbound HTTP requests
}

// Server contains specific values for inbound HTTP requests
type Server struct {
	Duration time.Duration `json:"duration,omitempty"` // Measures the duration of the inbound HTTP request in ms
}

// Client contains specific values for outbound HTTP requests
type Client struct {
	Duration time.Duration `json:"duration,omitempty"` // Measures the duration of the outbound HTTP request in ms
}

// DB contains information about outbound db requests.
type DB struct {
	Name      string `json:"name,omitempty"`      // This attribute is used to report the name of the database being accessed. For example customers oder main
	Statement string `json:"statement,omitempty"` // The database statement being executed. Must be sanitized to exclude sensitive information. For example SELECT * FROM wuser_table; SET mykey "WuValue"
	Operation string `json:"operation,omitempty"` // The name of the operation being executed, e.g. the MongoDB command name such as findAndModify, or the SQL keyword. For example findAndModify; HMSET; SELECT
}

// Exception contains information about an exception
type Exception struct {
	Type       string `json:"type,omitempty"`       // The type of the exception (its fully-qualified class name, if applicable). The dynamic type of the exception should be preferred over the static type in languages that support it. For example java.net.ConnectException; OSError
	Message    string `json:"message,omitempty"`    // The exception message. For example Division by zero
	Stacktrace string `json:"stacktrace,omitempty"` // A stacktrace as a string in the natural representation for the language runtime.
}

// MarshalJSON customizes the JSON Representation of the Event type
func (e Event) MarshalJSON() ([]byte, error) {
	type Alias Event // type alias to prevent infinite recursion
	ev := Alias(e)

	if ev.Visibility != nil && *ev.Visibility == 1 {
		cp := *ev.Visibility
		ev.Visibility = &cp
		ev.Visibility = nil // don't serialize default value
	}
	if ev.Time != nil && ev.Time.Location() != time.UTC {
		cp := *ev.Time // copy prevents changing the original event
		ev.Time = &cp
		*ev.Time = ev.Time.UTC() // normalize to UTC
	}

	return json.Marshal(ev)
}

// UnmarshalJSON customizes the JSON Deserialization of the Event type
func (e *Event) UnmarshalJSON(data []byte) error {
	type Alias Event // type alias to prevent infinite recursion
	var ev Alias
	if err := json.Unmarshal(data, &ev); err != nil {
		return err
	}
	if ev.Visibility == nil {
		var v = 1
		ev.Visibility = &v
	}
	*e = Event(ev)
	return nil
}

// MarshalJSON customizes the JSON Representation of the Server type
func (s Server) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Duration int64 `json:"duration,omitempty"` // Measures the duration of the inbound HTTP request in ms
	}{
		s.Duration.Milliseconds(),
	})
}

// UnmarshalJSON customizes the JSON Deserialization of the Event type
func (s *Server) UnmarshalJSON(data []byte) error {
	str := struct {
		Duration int64 `json:"duration,omitempty"` // Measures the duration of the inbound HTTP request in ms
	}{}

	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	s.Duration = time.Millisecond * time.Duration(str.Duration)
	return nil
}

// MarshalJSON customizes the JSON Representation of the Client type
func (s Client) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Duration int64 `json:"duration,omitempty"` // Measures the duration of the inbound HTTP request in ms
	}{
		s.Duration.Milliseconds(),
	})
}

// UnmarshalJSON customizes the JSON Deserialization of the Event type
func (s *Client) UnmarshalJSON(data []byte) error {
	str := struct {
		Duration int64 `json:"duration,omitempty"` // Measures the duration of the inbound HTTP request in ms
	}{}

	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	s.Duration = time.Millisecond * time.Duration(str.Duration)
	return nil
}

type Severity uint8

const (
	SeverityDebug = 5  // The information is meant for the developer of the app or component. The purpose it to follow the execution path while explicitly debugging a certain problem.
	SeverityInfo  = 9  // The information is meant for the developer or operator of the own or other teams. In contrast to SeverityError this severity is used to emit events which work as designed.
	SeverityError = 17 // The information is meant for the developer or operator of the own or other teams. In contrast to SeverityInfo this severity number is used to emit events which denote that something unexpected happened. Which can't be compensated in the own component. For example a failed outbound http request which can be compensated with a subsequent retry must not be logged as SeverityError. SeverityInfo must be used because from the outside everything works as expected. Whereas an inbound request which yields a 500 - internal server error must use SeverityError because there is no chance for the own component to recover.
)
