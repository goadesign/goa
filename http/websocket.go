package http

import (
	"net/http"

	"github.com/gorilla/websocket"
)

type (
	// Upgrader is an HTTP connection that is able to upgrade to websocket.
	Upgrader interface {
		// Upgrade upgrades the HTTP connection to the websocket protocol.
		Upgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*websocket.Conn, error)
	}

	// Dialer creates a websocket connection to a given URL.
	Dialer interface {
		// Dial creates a client connection to the websocket server.
		Dial(url string, h http.Header) (*websocket.Conn, *http.Response, error)
	}

	// ConnConfigureFunc is used to configure a websocket connection with
	// custom handlers.
	ConnConfigureFunc func(*websocket.Conn) *websocket.Conn
)
