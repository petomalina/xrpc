package multiplexer

import (
	"context"
	"github.com/petomalina/xrpc/examples/api"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"net/http"
	"testing"
)

type GrpcHandlerSuite struct {
	suite.Suite

	srv         *http.Server
	echoService *EchoService
}

func (s *GrpcHandlerSuite) SetupTest() {
	s.echoService = &EchoService{createLogger(), nil}
	grpcServer := createGrpcServer(s.echoService)

	s.srv = createTestServer(
		GRPCHandler(grpcServer),
	)

	go func() {
		// an error is returned when the server is closed externally. This is normal
		s.Error(http.ErrServerClosed, s.srv.ListenAndServe(), "error listening")
	}()
}

func (s *GrpcHandlerSuite) TearDownTest() {
	s.NoError(s.srv.Close(), "error closing the server")
}

func (s *GrpcHandlerSuite) TestGrpcHandler() {
	conn, err := grpc.Dial(testingTarget.String(), grpc.WithInsecure())
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
