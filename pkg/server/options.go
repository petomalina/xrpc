package server

// Option is an extendable builder for server options
type Option func(s *server)

// WithHost sets the host that the server binds to
func WithHost(host string) Option {
	return func(s *server) {
		s.host = host
	}
}
