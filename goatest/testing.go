package goatest

import (
	"bytes"
	"io"
	"log"

	"github.com/goadesign/goa"
	"github.com/goadesign/goa/middleware"
)

// ResponseSetterFunc func
type ResponseSetterFunc func(resp interface{})

// Encode implements a dummy encoder that returns the value being encoded
func (r ResponseSetterFunc) Encode(v interface{}) error {
	r(v)
	return nil
}

// Service provide a general goa.Service used for testing purposes
func Service(logBuf *bytes.Buffer, respSetter ResponseSetterFunc) *goa.Service {
	s := goa.New("test")
	logger := log.New(logBuf, "", log.Ltime)
	s.WithLogger(goa.NewLogger(logger))
	s.Use(middleware.LogRequest(true))
	s.Use(middleware.LogResponse())
	newEncoder := func(io.Writer) goa.Encoder {
		return respSetter
	}
	s.Decoder(goa.NewJSONDecoder, "*/*")
	s.Encoder(newEncoder, "*/*")
	return s
}
