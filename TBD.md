Notes of things I know still have to be done. Lots more I don't know about.

- Generalize middleware, split with http specific middleware
- Move mux out of service
- Generalize errors

- Fix up client so New works with stdlib client (or add new NewStd)
- Add back security


Generated code to be:

app package:

    type BottleService interface {
        Create(ctx context.Context, payload *BottleCreatePayload) (*Bottle, error)
        Show(ctx context.Context, bottleID int) (*Bottle, error)
        Rate(ctx context.Context, bottleID int) error
    }

    type BottleEndpoints struct {
        Create goa.Endpoint
        Show   goa.Endpoint
        Rate   goa.Endpoint
    }

    type BottleHTTPHandlers struct {
        Create func(http.ResponseWriter, *http.Request)
        Show   func(http.ResponseWriter, *http.Request)
        Rate   func(http.ResponseWriter, *http.Request)
    }

    type BottleGRPCServer struct {
        Create func(ctx context.Context, req *pb.CreateRequest) (*pb.CreateReply, error)
        Show   func(ctx context.Context, req *pb.ShowRequest) (*pb.ShowReply, error)
        Rate   func(ctx context.Context, req *pb.RateRequest) (*pb.RateReply, error)
    }

    func NewBottleEndpoints(s BottleService) *BottleEndpoints {
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

    func NewBottleHTTPHandlers(ep *BottleEndpoints) *BottleHTTPHandlers {
        // ...
    }

    func NewBottleGRPCServer(ep *BottleEndpoints) *BottleGRPCServer {
        // ...
    }

    func MountBottleHTTPHandler(server *goa.HTTPServer, hanlders *BottleHTTPHandlers) {
        // ... config server mux
    }

    func MountBottleGRPCServer(server *goa.GRPCServer, handlers *BottleGRPCServer) {
        // ... build proto compat shim and call proto generated Register
    }

main package

    func NewBottleService() app.BottleService {
        // ...
    }

    bottleService := NewBottleService()
    bottleEndpoints := NewBottleEndpoints(bottleService)
    bottleHTTPHandlers := NewBottleHTTPHandlers(bottleEndpoints)
    bottleGRPCServer := NewBottleGRPCServer(bottleEndpoints)

    app.MountBottleHTTPHandlers(server, bottleHTTPHandlers)
    app.MountBottleGRPCServer(grpcServer, bottleGRPCServer)

