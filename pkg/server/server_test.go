package server

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type ServerSuite struct {
	suite.Suite

	// register this so the connection can be torn down in case
	// the test panicked
	srv *Server
}

func (s *ServerSuite) SetupTest() {}

func (s *ServerSuite) TearDownTest() {
	s.NoError(s.srv.listener.Close(), "error closing the server")
}

func (s *ServerSuite) TestNewRandom() {
	srv, err := New(RandomPort, DefaultTimeout, WithHost("localhost"))
	s.NoError(err)
	s.srv = srv

	s.NotEqual("", srv.IP())
	s.NotEqual("", srv.Host())
	s.NotEqual("", srv.Port())
}

func (s *ServerSuite) TestNewDefined() {
	srv, err := New("50099", DefaultTimeout, WithHost("localhost"))
	s.NoError(err)
	s.srv = srv

	s.Equal("127.0.0.1", srv.IP())
	s.Equal("localhost", srv.Host())
	s.Equal("50099", srv.Port())
}

// NoPort means the port will be overwritten to the RandomPort
func (s *ServerSuite) TestNewWithHostNoPort() {
	srv, err := New("", DefaultTimeout, WithHost("localhost"))
	s.NoError(err)
	s.srv = srv

	s.Equal("127.0.0.1", srv.IP())
	s.Equal("localhost", srv.Host())
	s.NotEqual("", srv.Port())
}

func (s *ServerSuite) TestDefinedPortNoHost() {
	srv, err := New("50099", DefaultTimeout)
	s.NoError(err)
	s.srv = srv

	s.Equal("::", srv.IP())
	s.Equal("", srv.Host())
	s.Equal("50099", srv.Port())
}

func TestServerSuite(t *testing.T) {
	suite.Run(t, &ServerSuite{})
}
