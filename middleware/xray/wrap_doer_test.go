package xray

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

func TestWrapDoer(t *testing.T) {
	RegisterTestingT(t)

	var (
		doer = NewMockDoer()
		ctx  = context.Background()
	)
	req, err := http.NewRequest("GET", "http://somehost:80/path", nil)
	Expect(err).To(Succeed())

	t.Run(`no segment in context; success`, func(t *testing.T) {
		RegisterTestingT(t)

		doer.Expect("Do", func(c context.Context, r *http.Request) (*http.Response, error) {
			Expect(r).To(Equal(req))
			Expect(c).To(Equal(ctx))
			return &http.Response{StatusCode: 123}, nil
		})
		resp, err := WrapDoer(doer).Do(ctx, req)
		Expect(err).To(Succeed())
		Expect(resp.StatusCode).To(Equal(123))
		Expect(doer.MetExpectations()).To(Succeed())
	})

	const (
		segmentName = "segmentName1"
		traceID     = "traceID1"
		spanID      = "spanID1"
	)

	t.Run(`with a segment in context - successful request`, func(t *testing.T) {
		RegisterTestingT(t)

		// add an xray segment to the context
		xrayConn := NewTestNetConn()
		segment := NewSegment(segmentName, traceID, spanID, xrayConn)
		ctx = WithSegment(ctx, segment)

		doer.Expect("Do", func(c context.Context, r *http.Request) (*http.Response, error) {
			Expect(r).To(Equal(req))
			Expect(ContextSegment(c).ParentID).To(Equal(segment.ID))
			return &http.Response{StatusCode: 123}, nil
		})

		xrayConn.Expect("Write", func(b []byte) (int, error) {
			lines := strings.Split(string(b), "\n")
			Expect(lines).To(HaveLen(2))
			Expect(lines[0]).To(Equal(`{"format": "json", "version": 1}`))

			var s Segment
			err := json.Unmarshal([]byte(lines[1]), &s)
			Expect(err).To(Succeed())
			Expect(s).To(MatchFields(IgnoreMissing|IgnoreExtras, Fields{
				"Name":      Equal("somehost:80"),
				"Namespace": Equal("remote"),
				"Type":      Equal("subsegment"),
				"ID":        And(Not(BeEmpty()), Not(Equal(segment.ID))), // randomly generated
				"TraceID":   Equal(traceID),
				"ParentID":  Equal(spanID),
				"Error":     BeFalse(),
				"HTTP": PointTo(MatchAllFields(Fields{
					"Request":  Equal(&Request{Method: "GET", URL: "http://somehost:80/path"}),
					"Response": Equal(&Response{Status: 123}),
				})),
			}))
			return len(b), nil
		})
		resp, err := WrapDoer(doer).Do(ctx, req)
		Expect(err).To(Succeed())
		Expect(resp.StatusCode).To(Equal(123))
		Expect(doer.MetExpectations()).To(Succeed())
		Expect(xrayConn.MetExpectations()).To(Succeed())
	})

	t.Run(`with a segment in context - failed request`, func(t *testing.T) {
		RegisterTestingT(t)

		// add an xray segment to the context
		xrayConn := NewTestNetConn()
		segment := NewSegment(segmentName, traceID, spanID, xrayConn)
		ctx = WithSegment(ctx, segment)

		var (
			requestErr = errors.New("some request error")
		)
		doer.Expect("Do", func(c context.Context, r *http.Request) (*http.Response, error) {
			Expect(ContextSegment(c).ParentID).To(Equal(segment.ID))
			return nil, requestErr
		})

		xrayConn.Expect("Write", func(b []byte) (int, error) {
			lines := strings.Split(string(b), "\n")
			Expect(lines).To(HaveLen(2))
			Expect(lines[0]).To(Equal(`{"format": "json", "version": 1}`))

			var s Segment
			err := json.Unmarshal([]byte(lines[1]), &s)
			Expect(err).To(Succeed())
			Expect(s).To(MatchFields(IgnoreMissing|IgnoreExtras, Fields{
				"Name":      Equal("somehost:80"),
				"Namespace": Equal("remote"),
				"Type":      Equal("subsegment"),
				"ID":        And(Not(BeEmpty()), Not(Equal(segment.ID))), // randomly generated
				"TraceID":   Equal(traceID),
				"ParentID":  Equal(spanID),
				"Error":     BeTrue(),
				"HTTP": PointTo(MatchAllFields(Fields{
					"Request":  Equal(&Request{Method: "GET", URL: "http://somehost:80/path"}),
					"Response": BeNil(),
				})),
			}))
			return len(b), nil
		})
		_, err := WrapDoer(doer).Do(ctx, req)
		Expect(err).To(MatchError(requestErr))

		Expect(doer.MetExpectations()).To(Succeed())
		Expect(xrayConn.MetExpectations()).To(Succeed())
	})
}

type MockDoer struct {
	*TestClientExpectation
}

func NewMockDoer() *MockDoer {
	return &MockDoer{NewTestClientExpectation()}
}

func (m *MockDoer) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	if e := m.Expectation("Do"); e != nil {
		return e.(func(context.Context, *http.Request) (*http.Response, error))(ctx, req)
	}
	return nil, nil
}
