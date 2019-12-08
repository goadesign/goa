package testdata

var ServerMultiEndpointsConstructorCode = `// New instantiates HTTP handlers for all the ServiceMultiEndpoints service
// endpoints using the provided encoder and decoder. The handlers are mounted
// on the given mux using the HTTP verb and path defined in the design.
// errhandler is called whenever a response fails to be encoded. formatter is
// used to format errors returned by the service methods prior to encoding.
// Both errhandler and formatter are optional and can be nil.
func New(
	e *servicemultiendpoints.Endpoints,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(err error) goahttp.Statuser,
) *Server {
	return &Server{
		Mounts: []*MountPoint{
			{"MethodMultiEndpoints1", "GET", "/server_multi_endpoints/{id}"},
			{"MethodMultiEndpoints2", "POST", "/server_multi_endpoints"},
		},
		MethodMultiEndpoints1: NewMethodMultiEndpoints1Handler(e.MethodMultiEndpoints1, mux, decoder, encoder, errhandler, formatter),
		MethodMultiEndpoints2: NewMethodMultiEndpoints2Handler(e.MethodMultiEndpoints2, mux, decoder, encoder, errhandler, formatter),
	}
}
`

var ServerMultiBasesConstructorCode = `// New instantiates HTTP handlers for all the ServiceMultiBases service
// endpoints using the provided encoder and decoder. The handlers are mounted
// on the given mux using the HTTP verb and path defined in the design.
// errhandler is called whenever a response fails to be encoded. formatter is
// used to format errors returned by the service methods prior to encoding.
// Both errhandler and formatter are optional and can be nil.
func New(
	e *servicemultibases.Endpoints,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(err error) goahttp.Statuser,
) *Server {
	return &Server{
		Mounts: []*MountPoint{
			{"MethodMultiBases", "GET", "/base_1/{id}"},
			{"MethodMultiBases", "GET", "/base_2/{id}"},
		},
		MethodMultiBases: NewMethodMultiBasesHandler(e.MethodMultiBases, mux, decoder, encoder, errhandler, formatter),
	}
}
`

var ServerFileServerConstructorCode = `// New instantiates HTTP handlers for all the ServiceFileServer service
// endpoints using the provided encoder and decoder. The handlers are mounted
// on the given mux using the HTTP verb and path defined in the design.
// errhandler is called whenever a response fails to be encoded. formatter is
// used to format errors returned by the service methods prior to encoding.
// Both errhandler and formatter are optional and can be nil.
func New(
	e *servicefileserver.Endpoints,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(err error) goahttp.Statuser,
) *Server {
	return &Server{
		Mounts: []*MountPoint{
			{"/path/to/file1.json", "GET", "/server_file_server/file1.json"},
			{"/path/to/file2.json", "GET", "/server_file_server/file2.json"},
			{"/path/to/file3.json", "GET", "/server_file_server/file3.json"},
		},
	}
}
`

var ServerMixedConstructorCode = `// New instantiates HTTP handlers for all the ServerMixed service endpoints
// using the provided encoder and decoder. The handlers are mounted on the
// given mux using the HTTP verb and path defined in the design. errhandler is
// called whenever a response fails to be encoded. formatter is used to format
// errors returned by the service methods prior to encoding. Both errhandler
// and formatter are optional and can be nil.
func New(
	e *servermixed.Endpoints,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(err error) goahttp.Statuser,
) *Server {
	return &Server{
		Mounts: []*MountPoint{
			{"MethodMixed", "GET", "/{id}"},
			{"/path/to/file1.json", "GET", "/file1.json"},
			{"/path/to/file2.json", "GET", "/file2.json"},
		},
		MethodMixed: NewMethodMixedHandler(e.MethodMixed, mux, decoder, encoder, errhandler, formatter),
	}
}
`

var ServerMultipartConstructorCode = `// New instantiates HTTP handlers for all the ServiceMultipart service
// endpoints using the provided encoder and decoder. The handlers are mounted
// on the given mux using the HTTP verb and path defined in the design.
// errhandler is called whenever a response fails to be encoded. formatter is
// used to format errors returned by the service methods prior to encoding.
// Both errhandler and formatter are optional and can be nil.
func New(
	e *servicemultipart.Endpoints,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(err error) goahttp.Statuser,
	serviceMultipartMethodMultiBasesDecoderFn ServiceMultipartMethodMultiBasesDecoderFunc,
) *Server {
	return &Server{
		Mounts: []*MountPoint{
			{"MethodMultiBases", "GET", "/"},
		},
		MethodMultiBases: NewMethodMultiBasesHandler(e.MethodMultiBases, mux, NewServiceMultipartMethodMultiBasesDecoder(mux, serviceMultipartMethodMultiBasesDecoderFn), encoder, errhandler, formatter),
	}
}
`

var ServerStreamingConstructorCode = `// New instantiates HTTP handlers for all the StreamingResultService service
// endpoints using the provided encoder and decoder. The handlers are mounted
// on the given mux using the HTTP verb and path defined in the design.
// errhandler is called whenever a response fails to be encoded. formatter is
// used to format errors returned by the service methods prior to encoding.
// Both errhandler and formatter are optional and can be nil.
func New(
	e *streamingresultservice.Endpoints,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(err error) goahttp.Statuser,
	upgrader goahttp.Upgrader,
	configurer *ConnConfigurer,
) *Server {
	if configurer == nil {
		configurer = &ConnConfigurer{}
	}
	return &Server{
		Mounts: []*MountPoint{
			{"StreamingResultMethod", "GET", "/"},
		},
		StreamingResultMethod: NewStreamingResultMethodHandler(e.StreamingResultMethod, mux, decoder, encoder, errhandler, formatter, upgrader, configurer.StreamingResultMethodFn),
	}
}
`

var ServerMultipleFilesConstructorCode = `// Mount configures the mux to serve the ServiceFileServer endpoints.
func Mount(mux goahttp.Muxer) {
	MountPathToFileJSON(mux, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "/path/to/file.json")
	}))
	MountPathToFileJSON1(mux, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "/path/to/file.json")
	}))
}
`
