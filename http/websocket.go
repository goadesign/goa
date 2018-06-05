package http

import (
	"net/http"

	"github.com/gorilla/websocket"
)

type (
	// Upgrader defines the interface a websocket connection upgrader must satisfy.
	Upgrader interface {
		// Upgrade upgrades the HTTP connection to the websocket protocol.
		Upgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*websocket.Conn, error)
	}

	// Dialer defines the interface a client websocket dialer must satisfy.
	Dialer interface {
		// Dial creates a client connection to the websocket server.
		Dial(url string, h http.Header) (*websocket.Conn, *http.Response, error)
	}

	// ConnConfigureFunc is used to configure a websocket connection with
	// custom handlers.
	ConnConfigureFunc func(*websocket.Conn) *websocket.Conn
)

var (
	// NormalSocketCloseErrors lists the normal close errors sent by the websocket.
	NormalSocketCloseErrors = []int{websocket.CloseNormalClosure}
)
