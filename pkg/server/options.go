package server

// Option is an extendable builder for server options
type Option func(s *Server)

// WithHost sets the host that the server binds to
func WithHost(host string) Option {
	return func(s *Server) {
		s.host = host
	}
}
