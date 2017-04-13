package xray

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/goadesign/goa"
	"github.com/pkg/errors"
)

type (
	// Segment represents a AWS X-Ray segment document.
	Segment struct {
		// Mutex used to synchronize access to segment.
		*sync.Mutex
		// Name is the name of the service reported to X-Ray.
		Name string `json:"name"`
		// Namespace identifies the source that created the segment.
		Namespace string `json:"namespace"`
		// Type is either the empty string or "subsegment".
		Type string `json:"type,omitempty"`
		// ID is a unique ID for the segment.
		ID string `json:"id"`
		// TraceID is the ID of the root trace.
		TraceID string `json:"trace_id,omitempty"`
		// ParentID is the ID of the parent segment when it is from a
		// remote service. It is only initialized for the root segment.
		ParentID string `json:"parent_id,omitempty"`
		// StartTime is the segment start time.
		StartTime float64 `json:"start_time,omitempty"`
		// EndTime is the segment end time.
		EndTime float64 `json:"end_time,omitempty"`
		// InProgress is true if the segment hasn't completed yet.
		InProgress bool `json:"in_progress"`
		// HTTP contains the HTTP request and response information and is
		// only initialized for the root segment.
		HTTP *HTTP `json:"http,omitempty"`
		// Cause contains information about an error that occurred while
		// processing the request.
		Cause *Cause `json:"cause,omitempty"`
		// Error is true when a request causes an internal error. It is
		// automatically set by Close when the response status code is
		// 500 or more.
		Error bool `json:"error"`
		// Fault is true when a request results in an error that is due
		// to the user. Typically it should be set when the response
		// status code is between 400 and 500 (but not 429).
		Fault bool `json:"fault"`
		// Throttle is true when a request is throttled. It is set to
		// true when the segment closes and the response status code is
		// 429. Client code may set it to true manually as well.
		Throttle bool `json:"throttle"`
		// Annotations contains the segment annotations.
		Annotations map[string]interface{} `json:"annotations,omitempty"`
		// Metadata contains the segment metadata.
		Metadata map[string]map[string]interface{} `json:"metadata,omitempty"`
		// Subsegments contains all the subsegments.
		Subsegments []*Segment `json:"subsegments,omitempty"`
		// Parent is the subsegment parent, it's nil for the root
		// segment.
		Parent *Segment `json:"-"`
		// conn is the UDP client to the X-Ray daemon.
		conn net.Conn
		// counter keeps track of the number of subsegments that have not
		// completed yet.
		counter int
	}

	// HTTP describes a HTTP request.
	HTTP struct {
		// Request contains the data reported about the incoming request.
		Request *Request `json:"request,omitempty"`
		// Response contains the data reported about the HTTP response.
		Response *Response `json:"response,omitempty"`
	}

	// Request describes a HTTP request.
	Request struct {
		Method        string `json:"method,omitempty"`
		URL           string `json:"url,omitempty"`
		UserAgent     string `json:"user_agent,omitempty"`
		ClientIP      string `json:"client_ip,omitempty"`
		ContentLength int64  `json:"content_length"`
	}

	// Response describes a HTTP response.
	Response struct {
		Status        int   `json:"status"`
		ContentLength int64 `json:"content_length"`
	}

	// Cause list errors that happens during the request.
	Cause struct {
		// ID to segment where error originated, exclusive with other
		// fields.
		ID string `json:"id,omitempty"`
		// WorkingDirectory when error occurred. Exclusive with ID.
		WorkingDirectory string `json:"working_directory,omitempty"`
		// Exceptions contains the details on the error(s) that occurred
		// when the request as processing.
		Exceptions []*Exception `json:"exceptions,omitempty"`
	}

	// Exception describes an error.
	Exception struct {
		// Message contains the error message.
		Message string `json:"message"`
		// Stack is the error stack trace as initialized via the
		// github.com/pkg/errors package.
		Stack []*StackEntry `json:"stack"`
	}

	// StackEntry represents an entry in a error stacktrace.
	StackEntry struct {
		// Path to code file
		Path string `json:"path"`
		// Line number
		Line int `json:"line"`
		// Label is the line label if any
		Label string `json:"label,omitempty"`
	}

	// key is the type used for context keys.
	key int
)

const (
	// udpHeader is the header of each segment sent to the daemon.
	udpHeader = "{\"format\": \"json\", \"version\": 1}\n"

	// maxStackDepth is the maximum number of stack frames reported.
	maxStackDepth = 100
)

type (
	causer interface {
		Cause() error
	}
	stackTracer interface {
		StackTrace() errors.StackTrace
	}
)

// NewSegment creates a new segment that gets written to the given connection
// on close.
func NewSegment(name, traceID, spanID string, conn net.Conn) *Segment {
	return &Segment{
		Mutex:      &sync.Mutex{},
		Name:       name,
		TraceID:    traceID,
		ID:         spanID,
		StartTime:  now(),
		InProgress: true,
		conn:       conn,
	}
}

// RecordRequest traces a request.
//
// It sets Http.Request & Namespace (ex: "remote")
func (s *Segment) RecordRequest(req *http.Request, namespace string) {
	s.Lock()
	defer s.Unlock()

	if s.HTTP == nil {
		s.HTTP = &HTTP{}
	}

	s.Namespace = namespace
	s.HTTP.Request = requestData(req)
}

// RecordResponse traces a response.
//
// It sets Throttle, Fault, Error and HTTP.Response
func (s *Segment) RecordResponse(resp *http.Response) {
	s.Lock()
	defer s.Unlock()

	if s.HTTP == nil {
		s.HTTP = &HTTP{}
	}

	s.recordStatusCode(resp.StatusCode)

	s.HTTP.Response = responseData(resp)
}

// RecordContextResponse traces a context response if present in the context
//
// It sets Throttle, Fault, Error and HTTP.Response
func (s *Segment) RecordContextResponse(ctx context.Context) {
	resp := goa.ContextResponse(ctx)
	if resp == nil {
		return
	}

	s.Lock()
	defer s.Unlock()

	if s.HTTP == nil {
		s.HTTP = &HTTP{}
	}

	s.recordStatusCode(resp.Status)
	s.HTTP.Response = &Response{resp.Status, int64(resp.Length)}
}

// RecordError traces an error. The client may also want to initialize the
// fault field of s.
//
// The trace contains a stack trace and a cause for the error if the argument
// was created using one of the New, Errorf, Wrap or Wrapf functions of the
// github.com/pkg/errors package. Otherwise the Stack and Cause fields are empty.
func (s *Segment) RecordError(e error) {
	xerr := exceptionData(e)

	s.Lock()
	defer s.Unlock()

	if s.Cause == nil {
		wd, _ := os.Getwd()
		s.Cause = &Cause{WorkingDirectory: wd}
	}
	s.Cause.Exceptions = append(s.Cause.Exceptions, xerr)
	p := s.Parent
	for p != nil {
		if p.Cause == nil {
			p.Cause = &Cause{ID: s.ID}
		}
		p = p.Parent
	}
}

// NewSubsegment creates a subsegment of s.
func (s *Segment) NewSubsegment(name string) *Segment {
	s.Lock()
	defer s.Unlock()

	sub := &Segment{
		Mutex:      &sync.Mutex{},
		ID:         NewID(),
		TraceID:    s.TraceID,
		ParentID:   s.ID,
		Type:       "subsegment",
		Name:       name,
		StartTime:  now(),
		InProgress: true,
		Parent:     s,
		conn:       s.conn,
	}
	s.Subsegments = append(s.Subsegments, sub)
	s.counter++
	return sub
}

// Capture creates a subsegment to record the execution of the given function.
// Usage:
//
//     s := xray.ContextSegment(ctx)
//     s.Capture("slow-func", func() {
//         // ... some long executing code
//     })
//
func (s *Segment) Capture(name string, fn func()) {
	sub := s.NewSubsegment(name)
	defer sub.Close()
	fn()
}

// AddAnnotation adds a key-value pair that can be queried by AWS X-Ray.
func (s *Segment) AddAnnotation(key string, value string) {
	s.addAnnotation(key, value)
}

// AddInt64Annotation adds a key-value pair that can be queried by AWS X-Ray.
func (s *Segment) AddInt64Annotation(key string, value int64) {
	s.addAnnotation(key, value)
}

// AddBoolAnnotation adds a key-value pair that can be queried by AWS X-Ray.
func (s *Segment) AddBoolAnnotation(key string, value bool) {
	s.addAnnotation(key, value)
}

// addAnnotation adds a key-value pair that can be queried by AWS X-Ray.
// AWS X-Ray only supports annotations of type string, integer or boolean.
func (s *Segment) addAnnotation(key string, value interface{}) {
	s.Lock()
	defer s.Unlock()

	if s.Annotations == nil {
		s.Annotations = make(map[string]interface{})
	}
	s.Annotations[key] = value
}

// AddMetadata adds a key-value pair to the metadata.default attribute.
// Metadata is not queryable, but is recorded.
func (s *Segment) AddMetadata(key string, value string) {
	s.addMetadata(key, value)
}

// AddInt64Metadata adds a key-value pair that can be queried by AWS X-Ray.
func (s *Segment) AddInt64Metadata(key string, value int64) {
	s.addMetadata(key, value)
}

// AddBoolMetadata adds a key-value pair that can be queried by AWS X-Ray.
func (s *Segment) AddBoolMetadata(key string, value bool) {
	s.addMetadata(key, value)
}

// addMetadata adds a key-value pair that can be queried by AWS X-Ray.
// AWS X-Ray only supports annotations of type string, integer or boolean.
func (s *Segment) addMetadata(key string, value interface{}) {
	s.Lock()
	defer s.Unlock()

	if s.Metadata == nil {
		s.Metadata = make(map[string]map[string]interface{})
		s.Metadata["default"] = make(map[string]interface{})
	}
	s.Metadata["default"][key] = value
}

// Close closes the segment by setting its EndTime.
func (s *Segment) Close() {
	s.Lock()
	defer s.Unlock()

	s.EndTime = now()
	s.InProgress = false
	if s.Parent != nil {
		s.Parent.decrementCounter()
	}
	if s.counter <= 0 {
		s.flush()
	}
}

// flush sends the segment to the AWS X-Ray daemon.
func (s *Segment) flush() {
	b, _ := json.Marshal(s)
	// append so we make only one call to Write to be goroutine-safe
	s.conn.Write(append([]byte(udpHeader), b...))
}

// recordStatusCode sets Throttle, Fault, Error
//
// It is expected that the mutex has already been locked when calling this method.
func (s *Segment) recordStatusCode(statusCode int) {
	switch {
	case statusCode == http.StatusTooManyRequests:
		s.Throttle = true
	case statusCode >= 400 && statusCode < 500:
		s.Fault = true
	case statusCode >= 500:
		s.Error = true
	}
}

// decrementCounter decrements the segment counter and flushes it if it's 0.
func (s *Segment) decrementCounter() {
	s.Lock()
	defer s.Unlock()

	s.counter--
	if s.counter <= 0 && s.EndTime != 0 {
		// Segment is closed and last subsegment closed, flush it
		s.flush()
	}
}

// exceptionData creates an Exception from an error.
func exceptionData(e error) *Exception {
	var xerr *Exception
	if c, ok := e.(causer); ok {
		xerr = &Exception{Message: c.Cause().Error()}
	} else {
		xerr = &Exception{Message: e.Error()}
	}
	if s, ok := e.(stackTracer); ok {
		st := s.StackTrace()
		ln := len(st)
		if ln > maxStackDepth {
			ln = maxStackDepth
		}
		frames := make([]*StackEntry, ln)
		for i := 0; i < ln; i++ {
			f := st[i]
			line, _ := strconv.Atoi(fmt.Sprintf("%d", f))
			frames[i] = &StackEntry{
				Path:  fmt.Sprintf("%s", f),
				Line:  line,
				Label: fmt.Sprintf("%n", f),
			}
		}
		xerr.Stack = frames
	}

	return xerr
}

// requestData creates a Request from a http.Request.
func requestData(req *http.Request) *Request {
	var (
		scheme = "http"
		host   = req.Host
	)
	if len(req.URL.Scheme) > 0 {
		scheme = req.URL.Scheme
	}
	if len(req.URL.Host) > 0 {
		host = req.URL.Host
	}

	return &Request{
		Method:        req.Method,
		URL:           fmt.Sprintf("%s://%s%s", scheme, host, req.URL.Path),
		ClientIP:      getIP(req),
		UserAgent:     req.UserAgent(),
		ContentLength: req.ContentLength,
	}
}

// responseData creates a Response from a http.Response.
func responseData(resp *http.Response) *Response {
	return &Response{
		Status:        resp.StatusCode,
		ContentLength: resp.ContentLength,
	}
}

// getIP implements a heuristic that returns an origin IP address for a request.
func getIP(req *http.Request) string {
	for _, h := range []string{"X-Forwarded-For", "X-Real-Ip"} {
		for _, ip := range strings.Split(req.Header.Get(h), ",") {
			if len(ip) == 0 {
				continue
			}
			realIP := net.ParseIP(strings.Replace(ip, " ", "", -1))
			return realIP.String()
		}
	}

	// not found in header
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return req.RemoteAddr
	}
	return host
}
