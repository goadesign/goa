Notes of things I know still have to be done. Lots more I don't know about.

- Generalize middleware, split with http specific middleware

- Fix up client so New works with stdlib client (or add new NewStd)
- Add back security
- Add back CORS


Generated code to be:

app package:

    type BottleController interface {
        Create(ctx context.Context, payload *BottleCreatePayload) (*Bottle, error)
        Show(ctx context.Context, bottleID int) (*Bottle, error)
        Rate(ctx context.Context, bottleID int) error
    }

    type BottleEndpoints struct {
        Logger goa.LogAdapter
        Create goa.Endpoint
        Show   goa.Endpoint
        Rate   goa.Endpoint
    }

    func (e *BottleEndpoints) Use(m goa.Middleware) {
        e.Create = m(e.Create)
        e.Show = m(e.Show)
        e.Rate = m(e.Rate)
    }

    type BottleHTTPServer struct {
    Logger goa.LogAdapter
    Encoder goahttp.Encoder
    Decoder goahttp.Decoder

        Create http.HandlerFunc
        Show   http.HandlerFunc
        Rate   http.HandlerFunc
    }

    func (s *BottleHTTPServer) Use(m http.Middleware) {
        s.Create = m(s.Create)
        s.Show = m(s.Show)
        s.Rate = m(s.Rate)
    }

    type BottleGRPCServer struct {
    Logger goa.LogAdapter
        Create func(ctx context.Context, req *pb.CreateRequest) (*pb.CreateReply, error)
        Show   func(ctx context.Context, req *pb.ShowRequest) (*pb.ShowReply, error)
        Rate   func(ctx context.Context, req *pb.RateRequest) (*pb.RateReply, error)
    }

    func NewBottleEndpoints(s BottleController) *BottleEndpoints {
        create := func(ctx context.Context, req interface{}) (resp interface{}, error) {
            r := req.(*BottleCreatePayload)
            resp, err := s.Create(r)
            if f, ok := err.(goa.Failure) {
                return nil, err
            }
            return resp, nil
        }
        // show := ...
        // rate := ...
        return &BottleEndpoints{
            Create: create,
            Show: show,
            Rate: rate,
        }
    }

    func NewBottleHTTPServer(ep *BottleEndpoints) *BottleHTTPServer {
        // ...
    }

    func NewBottleGRPCServer(ep *BottleEndpoints) *BottleGRPCServer {
        // ...
    }

    func MountBottleHTTPHandler(server *goa.HTTPServer, hanlders *BottleHTTPServer) {
        // ... config server mux
    }

    func MountBottleGRPCServer(server *goa.GRPCServer, handlers *BottleGRPCServer) {
        // ... build proto compat shim and call proto generated Register
    }

main package

    ## bottle.go
    func NewBottleController() app.BottleController {
        // ...
    }

    ## main.go
    bottleController := NewBottleController()
    bottleEndpoints := app.NewBottleEndpoints(bottleController)
    bottleHTTPServer := app.NewBottleHTTPServer(bottleEndpoints)
    bottleGRPCServer := app.NewBottleGRPCServer(bottleEndpoints)

    // bottleHTTPServer.Use(tracer)
    // bottleHTTPServer.Use(logger)

    host := goa.New()
    app.MountBottleHTTPServer(host, bottleHTTPServer)
    app.MountBottleGRPCServer(host, bottleGRPCServer)

    errc := make(chan error)
    ctx := context.Background()

    // Interrupt handler.
    go func() {
        c := make(chan os.Signal, 1)
        signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
        errc <- fmt.Errorf("%s", <-c)
    }()

    // HTTP transport.
    go func() {
        logger := log.NewContext(logger).With("transport", "HTTP")
        logger.Log("addr", *httpAddr)
        errc <- http.ListenAndServe(*httpAddr, host.Mux)
    }()

    // gRPC transport.
    go func() {
        logger := log.NewContext(logger).With("transport", "gRPC")

        ln, err := net.Listen("tcp", *grpcAddr)
        if err != nil {
            errc <- err
            return
        }

        s := grpc.NewServer()
        pb.RegisterBottleServer(s, bottleGRPCServer)

        logger.Log("addr", *grpcAddr)
        errc <- s.Serve(ln)
    }()

    // Run!
    logger.Log("exit", <-errc)

future

    // Debug listener.
    go func() {
        logger := log.NewContext(logger).With("transport", "debug")

        m := http.NewServeMux()
        m.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
        m.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
        m.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
        m.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
        m.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
        m.Handle("/metrics", stdprometheus.Handler())

        logger.Log("addr", *debugAddr)
        errc <- http.ListenAndServe(*debugAddr, m)
    }()

library

    ## package goa

    type Endpoint func(context.Context, interface{}) (interface{}, error)
    type Middleware func(Endpoint) Endpoint

    ## package http
    type Middleware func(http.HandlerFunc) http.HandlerFunc
