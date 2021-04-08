package multiplexer

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
	port     string
	listener net.Listener

	timeout time.Duration
}

// NewServer creates a new Server instance on the given port with the given timeout
func NewServer(port string, timeout time.Duration) (*Server, error) {
	addr := ":" + port
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to create listener for the server: %w", err)
	}

	return &Server{
		ip:       lis.Addr().(*net.TCPAddr).IP.String(),
		port:     strconv.Itoa(lis.Addr().(*net.TCPAddr).Port),
		listener: lis,

		timeout: timeout,
	}, nil
}

// ServeHTTP starts listening while watching the provided context for cancellation
func (s *Server) ServeHTTP(ctx context.Context, srv *http.Server) error {
	errCh := make(chan error, 1)
	go func() {
		<-ctx.Done()

		shutdownCtx, done := context.WithTimeout(context.Background(), 5*time.Second)
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

// ServeHTTPHandler creates a http.Server in case only handlers or mux is used
func (s *Server) ServeHTTPHandler(ctx context.Context, handler http.Handler) error {
	return s.ServeHTTP(ctx, &http.Server{
		Handler: handler,
	})
}
