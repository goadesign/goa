package http

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"
)

var (
	testString = "test string"
)

func TestResponseDecoder(t *testing.T) {
	cases := []struct {
		contentType string
		decoderType string
	}{
		{"application/json", "*json.Decoder"},
		{"+json", "*json.Decoder"},
		{"application/xml", "*xml.Decoder"},
		{"+xml", "*xml.Decoder"},
		{"application/gob", "*gob.Decoder"},
		{"+gob", "*gob.Decoder"},
		{"text/html", "*http.textDecoder"},
		{"+html", "*http.textDecoder"},
		{"text/plain", "*http.textDecoder"},
		{"+txt", "*http.textDecoder"},
	}

	for _, c := range cases {
		t.Run(c.contentType, func(t *testing.T) {
			r := &http.Response{
				Header: map[string][]string{
					"Content-Type": {c.contentType},
				},
			}
			decoder := ResponseDecoder(r)
			if c.decoderType != fmt.Sprintf("%T", decoder) {
				t.Errorf("got decoder type %s, expected %s", fmt.Sprintf("%T", decoder), c.decoderType)
			}
		})
	}
}

func TestTextEncoder_Encode(t *testing.T) {
	cases := []struct {
		name  string
		value interface{}
		error error
	}{
		{"string", testString, nil},
		{"*string", &testString, nil},
		{"[]byte", []byte(testString), nil},
		{"other", 123, fmt.Errorf("can't encode int as content/type")},
	}

	buffer := bytes.Buffer{}
	encoder := newTextEncoder(&buffer, "content/type")

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			buffer.Reset()
			err := encoder.Encode(c.value)
			if c.error != nil {
				if err == nil || c.error.Error() != err.Error() {
					t.Errorf("got error %q, expected %q", err, c.error)
				}
			} else {
				if err != nil {
					t.Errorf("got error %q, expected <nil>", err)
				}
				if buffer.String() != testString {
					t.Errorf("got string %s, expected %s", buffer.String(), testString)
				}
			}
		})
	}
}

func TestTextPlainDecoder_Decode_String(t *testing.T) {
	decoder := makeTextDecoder()

	var value string
	err := decoder.Decode(&value)
	if err != nil {
		t.Errorf("got error %q, expected <nil>", err)
	}
	if testString != value {
		t.Errorf("got string %s, expected %s", value, testString)
	}
}

func TestTextPlainDecoder_Decode_Bytes(t *testing.T) {
	decoder := makeTextDecoder()

	var value []byte
	err := decoder.Decode(&value)
	if err != nil {
		t.Errorf("got error %q, expected <nil>", err)
	}
	if testString != string(value) {
		t.Errorf("got string %s, expected %s", value, testString)
	}
}

func TestTextPlainDecoder_Decode_Other(t *testing.T) {
	decoder := makeTextDecoder()

	expectedErr := fmt.Errorf("can't decode content/type to *int")

	var value int
	err := decoder.Decode(&value)
	if err == nil || err.Error() != expectedErr.Error() {
		t.Errorf("got error %q, expectedErr %q", err, expectedErr)
	}
}

func makeTextDecoder() Decoder {
	buffer := bytes.Buffer{}
	buffer.WriteString(testString)
	return newTextDecoder(&buffer, "content/type")
}
