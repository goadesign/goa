package http

import (
	"bytes"
	"fmt"
	"testing"
)

var (
	testString = "test string"
)

func TestTextPlainEncoder_Encode(t *testing.T) {
	cases := []struct {
		name  string
		value interface{}
		error error
	}{
		{"string", testString, nil},
		{"*string", &testString, nil},
		{"[]byte", []byte(testString), nil},
		{"other", 123, fmt.Errorf("can't encode int as text/plain")},
	}

	buffer := bytes.Buffer{}
	encoder := textPlainEncoder{w: &buffer}

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
	buffer := bytes.Buffer{}
	buffer.WriteString(testString)
	encoder := textPlainDecoder{r: &buffer}

	var value string
	err := encoder.Decode(&value)
	if err != nil {
		t.Errorf("got error %q, expected <nil>", err)
	}
	if testString != value {
		t.Errorf("got string %s, expected %s", value, testString)
	}
}

func TestTextPlainDecoder_Decode_Bytes(t *testing.T) {
	buffer := bytes.Buffer{}
	buffer.WriteString(testString)
	encoder := textPlainDecoder{r: &buffer}

	var value []byte
	err := encoder.Decode(&value)
	if err != nil {
		t.Errorf("got error %q, expected <nil>", err)
	}
	if testString != string(value) {
		t.Errorf("got string %s, expected %s", value, testString)
	}
}

func TestTextPlainDecoder_Decode_Other(t *testing.T) {
	buffer := bytes.Buffer{}
	buffer.WriteString(testString)
	encoder := textPlainDecoder{r: &buffer}

	expected := fmt.Errorf("can't decode text/plain to *int")

	var value int
	err := encoder.Decode(&value)
	if err == nil || err.Error() != expected.Error() {
		t.Errorf("got error %q, expected %q", err, expected)
	}
}
