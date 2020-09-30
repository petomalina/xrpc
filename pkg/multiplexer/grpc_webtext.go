package multiplexer

import (
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"net/http"
)

// GRPCWebTextHandler fulfills requests that are considered to be grpc web text requests
func GRPCWebTextHandler(server *grpcweb.WrappedGrpcServer, selectors ...Selector) Handler {
	filter := append([]Selector{
		OrSelector(
			server.IsAcceptableGrpcCorsRequest,
			server.IsGrpcWebRequest,
		),
	}, selectors...)

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
