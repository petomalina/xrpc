package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"
)

// Server wraps an implementation of a HTTP server with graceful shutdown
type Server struct {
	ip       string
	host     string
	port     string
	listener net.Listener

	timeout time.Duration
}

// Port returns the port of the server
func (s *Server) Port() string {
	return s.port
}

// Host returns the host of the server
func (s *Server) Host() string {
	return s.host
}

// IP returns the IP address of the server
func (s *Server) IP() string {
	return s.ip
}

const (
	// RandomPort is a constant used to let the system decide on the port
	// for the server. This is commonly used for services that connect to
	// each other via a registry (the server registers itself with host+port
	// in a central list, so multiple can run on a single machine)
	RandomPort = "0"

	// DefaultTimeout represent a default value for the timeout option of the Server
	DefaultTimeout = time.Second * 30
)

// New creates a new server instance on the given port with the given timeout
// If no port is given, RandomPort is used instead
func New(port string, timeout time.Duration, opts ...Option) (*Server, error) {
	s := &Server{}
	for _, opt := range opts {
		opt(s)
	}

	addr := ""
	if port != "" {
		addr = ":" + port
	} else {
		addr = ":" + RandomPort
	}

	if s.host != "" {
		addr = s.host + addr
	}

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to create listener for the server: %w", err)
	}

	s.ip = lis.Addr().(*net.TCPAddr).IP.String()
	s.port = strconv.Itoa(lis.Addr().(*net.TCPAddr).Port)
	s.listener = lis
	s.timeout = timeout

	return s, nil
}

// Start bootstraps a default http server and starts handling requests
func Start(ctx context.Context, port string, timeout time.Duration, handler http.Handler, opts ...Option) error {
	srv, err := New(port, timeout, opts...)
	if err != nil {
		return err
	}

	return srv.ServeHTTPHandler(ctx, handler)
}

// ServeHTTP starts listening while watching the provided context for cancellation
func (s *Server) ServeHTTP(ctx context.Context, srv *http.Server) error {
	errCh := make(chan error, 1)
	go func() {
		<-ctx.Done()

		shutdownCtx, done := context.WithTimeout(context.Background(), s.timeout)
		defer done()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			select {
			case errCh <- err:
			default:
			}
		}
	}()

	if err := srv.Serve(s.listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to serve: %w", err)
	}

	select {
	case err := <-errCh:
		return fmt.Errorf("failed to shutdown: %w", err)
	default:
		return nil
	}
}

// ServeHTTPHandler creates a http.server in case only handlers or mux is used
func (s *Server) ServeHTTPHandler(ctx context.Context, handler http.Handler) error {
	return s.ServeHTTP(ctx, &http.Server{
		Handler: handler,
	})
}
