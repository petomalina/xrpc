package multiplexer

import (
	"context"
	"github.com/petomalina/xrpc/examples/api"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"strconv"
	"testing"
)

type GrpcHandlerSuite struct {
	suite.Suite

	port        string
	srv         *http.Server
	echoService *EchoService
}

func (s *GrpcHandlerSuite) SetupTest() {
	lis, err := net.Listen("tcp", ":0")
	s.NoError(err)
	s.port = strconv.Itoa(lis.Addr().(*net.TCPAddr).Port)

	s.echoService = &EchoService{Logger: createLogger()}
	grpcServer := createGrpcServer(s.echoService)

	s.srv = createTestServer(
		GRPCHandler(grpcServer),
	)

	go func() {
		// an error is returned when the server is closed externally. This is normal
		s.Error(http.ErrServerClosed, s.srv.Serve(lis), "error listening")
	}()
}

func (s *GrpcHandlerSuite) TearDownTest() {
	s.NoError(s.srv.Close(), "error closing the server")
}

func (s *GrpcHandlerSuite) TestGrpcHandler() {
	conn, err := grpc.Dial("localhost:"+s.port, grpc.WithInsecure())
	s.NoError(err)

	client := api.NewEchoServiceClient(conn)
	s.NotNil(client)

	const msg = "Hey There!"
	res, err := client.Call(context.Background(), &api.EchoMessage{
		Message: msg,
	})
	s.NoError(err)
	s.Equal(msg, res.Message)
}

func TestGrpcHandlerSuite(t *testing.T) {
	suite.Run(t, &GrpcHandlerSuite{})
}
