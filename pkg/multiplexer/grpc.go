package multiplexer

import (
	"net/http"
	"strings"
)

// IsGRPCRequest returns true if the message is considered to be
// a GRPC message
func IsGRPCRequest(r *http.Request) bool {
	return r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc")
}

// GRPCHandler fulfills requests that are considered to be grpc requests
func GRPCHandler(server http.Handler, selectors ...Selector) Handler {
	filter := append([]Selector{IsGRPCRequest}, selectors...)

	return func(w http.ResponseWriter, r *http.Request) bool {
		for _, f := range filter {
			if !f(r) {
				return false
			}
		}

		server.ServeHTTP(w, r)
		return true
	}
}
