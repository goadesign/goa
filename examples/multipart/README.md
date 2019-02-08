# Resume Service

This example illustrates encoding and decoding multipart requests in goa v2.

## Design

The design describes two methods `list` and `add`. The `add` method receives
the resumes to store as a multipart HTTP request.

```
  Method("add", func() {
    Payload(ArrayOf(Resume))
    HTTP(func() {
      POST("/")
      MultipartRequest()
    })
  })
```

## Code Generation

`goa example` command creates a file called `multipart.go` in the top-level
directory which contains a dummy implementation of multipart decoder and
encoder functions for the `add` method. The multipart decoder function must 
populate the payload instance passed as a parameter. Application developers
must implement these functions or can provide their own encoders/decoders
that satisfy the function signatures and pass them as arguments in client and
server initializations as shown below.

```
// cmd/resume/http.go

  var (
    resumeServer *resumesvr.Server
  )
  {
    eh := errorHandler(logger)
    resumeServer = resumesvr.New(resumeEndpoints, mux, dec, enc, eh, api.ResumeAddDecoderFunc)
  }

// cmd/resume-cli/http.go

  return cli.ParseEndpoint(
    scheme,
    host,
    doer,
    goahttp.RequestEncoder,
    goahttp.ResponseDecoder,
    debug,
    api.ResumeAddEncoderFunc,
  )
```
