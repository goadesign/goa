package goa

import "net"

type (
	// Service represents a network service that runs one ore more servers.
	//
	// A service implements an API defined in a design.
	Service struct {
		// Name of service used for logging, tracing etc.
		Name string
		// Servers that serve network requests made to the service.
		Servers []Server
	}

	// A Server that serves requests made by remote clients.
	Server interface {
		// Mount registers the server with the service.
		Mount(service *Service)
		// Serve accepts incoming connections on the Listener l,
		// creating a new service goroutine for each.
		Serve(l net.Listener) error
	}
)

// New instantiates a service with the given name.
func New(name string) *Service {
	return &Service{
		Name: name,
	}
}

// Serve calls Serve on all the service servers.
func (service *Service) Serve(l net.Listener) error {
	for _, server := range service.Servers {
		if err := server.Serve(l); err != nil {
			return err
		}
	}
	return nil
}
