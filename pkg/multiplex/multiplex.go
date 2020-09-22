package multiplex

import (
	"errors"
	"github.com/petomalina/xrpc/pkg/xpubsub"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"net/http"
	"strings"
)

var (
	ErrNoHandlerFulfilled = errors.New("no handler was fulfilled for your request")
)

type Multiplexer func(http.ResponseWriter, *http.Request) bool

func GrpcMultiplexer(handler http.Handler) Multiplexer {
	return func(w http.ResponseWriter, r *http.Request) bool {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			handler.ServeHTTP(w, r)
			return true
		}

		return false
	}
}

func PubSubMultiplexer(handler http.Handler) Multiplexer {
	return func(w http.ResponseWriter, r *http.Request) bool {
		if strings.Contains(r.Header.Get("user-agent"), "APIs-Google") {
			req, err := xpubsub.InterceptHTTP(r)
			if err != nil {
				_, _ = w.Write([]byte(err.Error()))
				w.WriteHeader(http.StatusBadRequest)
			}

			handler.ServeHTTP(w, req)
			return true
		}

		return false
	}
}

//func PubSubGrpcMultiplexer(conn *grpc.ClientConn) Multiplexer {
//	return func(w http.ResponseWriter, r *http.Request) bool {
//		if strings.Contains(r.Header.Get("user-agent"), "APIs-Google") {
//			body, err := xpubsub.InterceptGRPC(r)
//			if err != nil {
//				_, _ = w.Write([]byte(err.Error()))
//				w.WriteHeader(http.StatusBadRequest)
//			}
//
//			err := conn.Invoke(r.Context(), r.URL.Path, &body)
//			if err != nil {
//				_, _ = w.Write([]byte(err.Error()))
//				w.WriteHeader(http.StatusBadRequest)
//			}
//			return true
//		}
//
//		return false
//	}
//}

func DefaultHTTPMultiplexer(handler http.Handler) Multiplexer {
	return func(w http.ResponseWriter, r *http.Request) bool {
		handler.ServeHTTP(w, r)
		return true
	}
}

func MakeHandler(server *http2.Server, handlers ...Multiplexer) http.Handler {
	if server == nil {
		server = &http2.Server{}
	}

	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, h := range handlers {
			if h(w, r) {
				return
			}
		}

		_, _ = w.Write([]byte(ErrNoHandlerFulfilled.Error()))
		w.WriteHeader(http.StatusBadRequest)
	}), server)
}
