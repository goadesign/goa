Design:

    var CreateAccountPayload = Type("CreateAccountPayload", func() {
        Attribute("OrgID", String, "ID of organization that owns created account")
        Attribute("Name", String, "Name of account")
        Required("OrgID", "Name")
    })

    var Account = Type("Account, func() {
        Reference(CreateAccountPayload)
        Attribute("Href", String, "Href to account resource")
        Attribute("ID", String, "Unique account ID")
        Attribute("OrgID")
        Attribute("Name")
        Required("Href", "ID", "OrgID", "Name")
    })

    // OR

    var AccountMedia = MediaType("application/vnd.account", func() {
        Reference(CreateAccountPayload)
        Attributes(func() {
        	Attribute("Href", String, "Href to account resource")
            Attribute("ID", String, "Unique account ID")
            Attribute("OrgID")
            Attribute("Name")
        })
        Views("default", func() {
        	Attribute("Href")
            Attribute("ID")
            Attribute("OrgID")
            Attribute("Name")
        })
    })

    var _ = Service("account", func() {
        Description("Accout service") // Optional description

        Endpoint("create", func() {     // Defines a single endpoint
            Description("The create endpoint creates a new bottle")

            Payload(CreateAccountPayload)
            Result(Account) // Can use Type or MediaType. Views generated for GRPC
                            // as well if media type
            Error("NameAlreadyTaken", ErrNameTaken) // ErrNameTaken is optional
               // type that describes error body. If gRPC error attribute is added
               // to type, if return error matches design error then error
               // attribute is set otherwise error is returned to gRPC server.

			// HTTP defines HTTP transport specific properties
            HTTP(func() {
                POST("/Org/{OrgID}")        // Body is CreateAccountPayload minus OrgID attribute and headers
                POST("/")                   // Body is CreateAccountPayload minus headers
                Header("Api-Version")       // May match one of CreateAccountPayload attributes
                Response(func() {
                    Status(OK)              // Default
                    Header("Href:Location") // Href must be an attribute of Account, Location is name of header
                })
                Error("NameAlreadyTaken", func() {
                    Status(BadRequest)      // Default
                    Header("Message")       // Must be an attribute of NameAlreadyTaken
                })
            })

            GRPC(func() {
                Proto("account.create") // rpc definition in proto file
                Error("NameAlreadyTaken", func() { // Defines which field contains error if not "Error"
                    Field("CreationError")
                })
            })
        })
    })

User code:

    // Generated type (generated code)
    type Account struct {
        Href  string
        ID    string
        OrgID string
        Name  string
    }

    // Generated service (scaffold code)
    type AccountService struct {
    }

    // Generated encoders (generated code)
    func HTTPEncodeAccount(rw http.ResponseWriter, a *Account)
    func HTTPEncodeNameAlreadyTaken(rw http.ResponseWriter, e *NameAlreadyTaken)

    // Generated

    func (c * AccountService) Create(ctx context.Context, req *server.AccountCreatePayload) (*server.Account, error) {
        a := server.Account{
            Href: "/..",
        }

        return &a, nil

        // rest.ContextRequest(ctx) -> *http.Request
        // rest.ContextResponse(ctx) -> http.ResponseWriter
    }

    func (c * BottleController) Show(ctx context.Context, req *server.BottleShowRequest) (*server.BottleShowResponse, error) {
        resp := server.BottleShowResponse{
            Status: http.StatusOK,
            MediaType: mt,
        }
        // OR
        resp := server.NewBottleShowShowdResponse("/...", mt)

        // rest.ContextRequest(ctx) -> *http.Request
        // rest.ContextResponse(ctx) -> http.ResponseWriter
        return &resp, nil
    }

Generated code:

`endpoints` package:

    type BottleEndpoints struct {
        Create goa.Endpoint
        Show   goa.Endpoint
        Rate   goa.Endpoint
    }

    func (e *BottleEndpoints) Use(m goa.Middleware) {
        e.Create = m(e.Create)
        e.Show = m(e.Show)
        e.Rate = m(e.Rate)
    }

`server` package:

bottle.go:

    type BottleController interface {
        Create(ctx *context.Context, request *BottleCreateRequest) (*BottleCreateResponse, error)
        Show(ctx context.Context, request *BottleShowRequest) (*Bottle, error)
        Rate(ctx context.Context, request *BottleRateRequest) error

	// If multiple 2xx responses with different media types:
	Index(ctx context.Context, request *BottleIndexRequest) (interface{}, error)
	// Use appropriate server.NewXXXResponse
    }

    type BottleHTTPServer struct {
        Logger  goa.LogAdapter
        Encoder rest.Encoder
        Decoder rest.Decoder

        Create http.HandlerFunc
        Show   http.HandlerFunc
        Rate   http.HandlerFunc
    }

    func NewBottleEndpoints(c *BottleController) *BottleEndpoints {
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
            Show:   show,
            Rate:   rate,
        }
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

    func NewBottleHTTP(ep *BottleEndpoints) *BottleHTTPServer {
        // ...
    }

    func NewBottleGRP(ep *BottleEndpoints) *BottleGRPCServer {
        // ...
    }

    func MountBottleHTTPHandler(server *goa.HTTPServer, handlers *BottleHTTPServer) {
        // ... config server mux
    }

    func MountBottleGRPCServer(server *goa.GRPCServer, handlers *BottleGRPCServer) {
        // ... build proto compat shim and call proto generated Register
    }

`client` package:

    type BottleClient struct {
        Create(ctx *context.Context, request *BottleCreateRequest) (*BottleCreateResponse, error)
        Show(ctx context.Context, request *BottleShowRequest) (*Bottle, error)
        Rate(ctx context.Context, request *BottleRateRequest) error

	// If multiple 2xx responses with different media types:
	Index(ctx context.Context, request *BottleIndexRequest) (interface{}, error)
	// Use appropriate server.NewXXXResponse
    }

    func NewBottleEndpoints(c *BottleClient) *BottleEndpoints {
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
            Show:   show,
            Rate:   rate,
        }
    }


`main` package:

bottle.go:

    func NewBottleController() app.BottleController {
        // ...
    }

main.go:

    bottleController := NewBottleController()
    bottleEndpoints := server.NewBottleEndpoints(bottleController)
    bottleHTTPServer := server.NewBottleHTTP(bottleEndpoints)
    bottleGRPCServer := server.NewBottleGRPC(bottleEndpoints)

    // bottleHTTPServer.Use(tracer)
    // bottleHTTPServer.Use(logger)

    service := goa.New()
    server.MountBottleHTTP(service, bottleHTTPServer)
    server.MountBottleGRPC(service, bottleGRPCServer)

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
