package multiplexer

import (
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/stretchr/testify/suite"
	"net"
	"net/http"
	"strconv"
	"testing"
)

type GRPCWebTextSuite struct {
	suite.Suite

	srv         *http.Server
	port        string
	echoService *EchoService
}

func (s *GRPCWebTextSuite) SetupTest() {
	lis, err := net.Listen("tcp", ":0")
	s.NoError(err)

	s.port = strconv.Itoa(lis.Addr().(*net.TCPAddr).Port)

	s.echoService = &EchoService{createLogger(), nil}
	grpcServer := createGrpcServer(s.echoService)
	grpcWebServer := grpcweb.WrapServer(grpcServer,
		grpcweb.WithOriginFunc(func(origin string) bool {
			return true
		}),
	)

	s.srv = createTestServer(
		GRPCHandler(grpcServer),
		GRPCWebTextHandler(grpcWebServer),
	)

	go func() {
		// an error is returned when the server is closed externally. This is normal
		s.Error(http.ErrServerClosed, s.srv.Serve(lis), "error listening")
	}()
}

func (s *GRPCWebTextSuite) TearDownTest() {
	s.NoError(s.srv.Close(), "error closing the server")
}

func (s *GRPCWebTextSuite) TestOrSelector() {
	// a node.js implementation will probably be needed here :/
	// this will at least run the grpc web text handler in the server
}

func TestGRPCWebTextSuite(t *testing.T) {
	suite.Run(t, &GRPCWebTextSuite{})
}
