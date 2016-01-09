package goa

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"gopkg.in/inconshreveable/log15.v2"

	. "github.com/smartystreets/goconvey/convey"
)

func ExampleFilterHeaders() {
	headers := http.Header{}
	for header := range sensitiveHeaders {
		headers.Set(header, "")
	}
	headers.Set("header-not-sensitive-1", "") // header key will be convent to canonical format
	headers.Set("header-not-sensitive-2", "")

	iterator := func(name string, value []string) {
		fmt.Println(name == "Header-Not-Sensitive-1" || name == "Header-Not-Sensitive-2")
	}
	filterHeaders(headers, iterator)

	// Output:
	// true
	// true
}

func ExampleWriteHeaders() {
	headers := http.Header{}
	headers.Add("key", "val1")
	headers.Add("key", "val2")

	buffer := bytes.NewBuffer(nil)
	writeHeaders(buffer, headers)
	fmt.Println(buffer.String())

	// Output:
	// Key: val1, val2
}

// FIXME:
// Cannot find a good way to test Client
type clientFakeLogger struct {
	log15.Logger
	records []string
}

func (l *clientFakeLogger) Info(name string, _ ...interface{}) {
	l.records = append(l.records, name)
}

func TestClient(t *testing.T) {
	Convey("Given a new client", t, func() {
		c := NewClient()
		l := &clientFakeLogger{}
		c.Logger = l

		Convey("When request", func() {
			req, _ := http.NewRequest("", "", nil)
			c.Do(req)

			Convey(`Log records should resemble []string{"started"}`, func() {
				So(l.records, ShouldResemble, []string{"started"})
			})
		})
	})
}
