package multiplexer

import (
	"google.golang.org/grpc"
	"net/http"
	"strings"
)

// IsGRPCRequest returns true if the message is considered to be
// a GRPC message
func IsGRPCRequest(r *http.Request) bool {
	return r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc")
}

// GRPCHandler fulfills requests that are considered to be grpc requests
func GRPCHandler(server *grpc.Server) Handler {
	return func(w http.ResponseWriter, r *http.Request) bool {
		if !IsGRPCRequest(r) {
			return false
		}

		server.ServeHTTP(w, r)
		return true
	}
}
