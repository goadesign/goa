package testdata

var ServerMultiEndpointsConstructorCode = `// New instantiates HTTP handlers for all the ServiceMultiEndpoints service
// endpoints.
func New(
	e *servicemultiendpoints.Endpoints,
	mux goahttp.Muxer,
	dec func(*http.Request) goahttp.Decoder,
	enc func(context.Context, http.ResponseWriter) goahttp.Encoder,
	eh func(context.Context, http.ResponseWriter, error),
) *Server {
	return &Server{
		Mounts: []*MountPoint{
			{"MethodMultiEndpoints1", "GET", "/server_multi_endpoints/{id}"},
			{"MethodMultiEndpoints2", "POST", "/server_multi_endpoints"},
		},
		MethodMultiEndpoints1: NewMethodMultiEndpoints1Handler(e.MethodMultiEndpoints1, mux, dec, enc, eh),
		MethodMultiEndpoints2: NewMethodMultiEndpoints2Handler(e.MethodMultiEndpoints2, mux, dec, enc, eh),
	}
}
`

var ServerMultiBasesConstructorCode = `// New instantiates HTTP handlers for all the ServiceMultiBases service
// endpoints.
func New(
	e *servicemultibases.Endpoints,
	mux goahttp.Muxer,
	dec func(*http.Request) goahttp.Decoder,
	enc func(context.Context, http.ResponseWriter) goahttp.Encoder,
	eh func(context.Context, http.ResponseWriter, error),
) *Server {
	return &Server{
		Mounts: []*MountPoint{
			{"MethodMultiBases", "GET", "/base_1/{id}"},
			{"MethodMultiBases", "GET", "/base_2/{id}"},
		},
		MethodMultiBases: NewMethodMultiBasesHandler(e.MethodMultiBases, mux, dec, enc, eh),
	}
}
`

var ServerFileServerConstructorCode = `// New instantiates HTTP handlers for all the ServiceFileServer service
// endpoints.
func New(
	e *servicefileserver.Endpoints,
	mux goahttp.Muxer,
	dec func(*http.Request) goahttp.Decoder,
	enc func(context.Context, http.ResponseWriter) goahttp.Encoder,
	eh func(context.Context, http.ResponseWriter, error),
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

var ServerMixedConstructorCode = `// New instantiates HTTP handlers for all the ServerMixed service endpoints.
func New(
	e *servermixed.Endpoints,
	mux goahttp.Muxer,
	dec func(*http.Request) goahttp.Decoder,
	enc func(context.Context, http.ResponseWriter) goahttp.Encoder,
	eh func(context.Context, http.ResponseWriter, error),
) *Server {
	return &Server{
		Mounts: []*MountPoint{
			{"MethodMixed", "GET", "/{id}"},
			{"/path/to/file1.json", "GET", "/file1.json"},
			{"/path/to/file2.json", "GET", "/file2.json"},
		},
		MethodMixed: NewMethodMixedHandler(e.MethodMixed, mux, dec, enc, eh),
	}
}
`

var ServerMultipartConstructorCode = `// New instantiates HTTP handlers for all the ServiceMultipart service
// endpoints.
func New(
	e *servicemultipart.Endpoints,
	mux goahttp.Muxer,
	dec func(*http.Request) goahttp.Decoder,
	enc func(context.Context, http.ResponseWriter) goahttp.Encoder,
	eh func(context.Context, http.ResponseWriter, error),
	ServiceMultipartMethodMultiBasesDecoderFn ServiceMultipartMethodMultiBasesDecoderFunc,
) *Server {
	return &Server{
		Mounts: []*MountPoint{
			{"MethodMultiBases", "GET", "/"},
		},
		MethodMultiBases: NewMethodMultiBasesHandler(e.MethodMultiBases, mux, NewServiceMultipartMethodMultiBasesDecoder(mux, ServiceMultipartMethodMultiBasesDecoderFn), enc, eh),
	}
}
`
