package multiplexer

import (
	"errors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"net/http"
)

var (
	// ErrNoHandlerFulfilled is returned if no provided handler can fulfill a request
	// accepted by the server. These errors can be avoided using default handlers such
	// as the HTTP HTTPHandler, which automatically fulfills all requests
	ErrNoHandlerFulfilled = errors.New("no handler was fulfilled for your request")
)

// Handler is a http.Handler that returns true/false based on if the
// request is being fulfilled by the handler or not
type Handler func(http.ResponseWriter, *http.Request) bool

// Selector is a function that returns true if the request should be
// handled by the Handler is corresponds to
type Selector func(*http.Request) bool

// Make creates a new multiplexer with given handlers. It combines all handlers
// to create a new h2c handler. If the server is not provided, a default http2 server
// will be created instead.
func Make(server *http2.Server, handlers ...Handler) http.Handler {
	if server == nil {
		server = &http2.Server{}
	}

	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, h := range handlers {
			if h(w, r) {
				return
			}
		}

		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(ErrNoHandlerFulfilled.Error()))
	}), server)
}
