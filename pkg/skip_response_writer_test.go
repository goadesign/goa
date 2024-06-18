package goa

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestSkipResponseWriter(t *testing.T) {
	const input = "Hello, World!"
	var responseWriter io.ReadCloser

	responseWriter = SkipResponseWriter(strings.NewReader(input))
	defer func() {
		err := responseWriter.Close()
		if err != nil {
			t.Error(err)
		}
	}()
	_, ok := responseWriter.(io.WriterTo)
	if !ok {
		t.Errorf("SkipResponseWriter's result must implement io.WriterTo")
	}

	var writerToBuffer bytes.Buffer
	io.Copy(&writerToBuffer, responseWriter) // io.Copy uses WriterTo if implemented
	if writerToBuffer.String() != input {
		t.Errorf("WriteTo: expected=%q actual=%q", input, writerToBuffer.String())
	}

	responseWriter = SkipResponseWriter(strings.NewReader(input))
	defer func() {
		err := responseWriter.Close()
		if err != nil {
			t.Error(err)
		}
	}()
	readBytes, err := io.ReadAll(responseWriter) // io.ReadAll ignores WriterTo and calls Read
	if err != nil {
		t.Fatal(err)
	}
	if string(readBytes) != input {
		t.Errorf("Read: expected=%q actual=%q", input, string(readBytes))
	}
}
