package multiplexer

import (
	"context"
	"github.com/blendle/zapdriver"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/petomalina/xrpc/v2/examples/api"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net/http"
	"net/url"
)

const (
	testingPort = "8787"
)

var (
	testingTarget, _ = url.Parse("localhost:" + testingPort)
)

func createLogger() *zap.Logger {
	// create the zap logger for future use
	config := zapdriver.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	logger, err := config.Build(zapdriver.WrapCore(
		zapdriver.ReportAllErrors(true),
		zapdriver.ServiceName("echo"),
	))
	if err != nil {
		panic(err)
	}

	return logger
}

func createGrpcServer(service *EchoService) *grpc.Server {
	// create and register the grpc server
	grpcServer := grpc.NewServer()
	api.RegisterEchoServiceServer(grpcServer, service)
	reflection.Register(grpcServer)

	return grpcServer
}

func createGrpcGatewayServer(port string) http.Handler {
	ctx := context.Background()

	// create the grpc-gateway server and register to grpc server
	gwmux := runtime.NewServeMux()
	err := api.RegisterEchoServiceHandlerFromEndpoint(ctx, gwmux, "localhost:"+port, []grpc.DialOption{grpc.WithInsecure()})
	if err != nil {
		panic(err)
	}

	return gwmux
}

func createTestServer(hh ...Handler) *http.Server {
	handler := Make(nil,
		hh...,
	)

	return &http.Server{Handler: handler}
}

type EchoService struct {
	*zap.Logger

	onCall func(ctx context.Context, m *api.EchoMessage)

	api.UnimplementedEchoServiceServer
}

func (e *EchoService) Call(ctx context.Context, m *api.EchoMessage) (*api.EchoMessage, error) {
	if e.onCall != nil {
		e.onCall(ctx, m)
	}

	headers := metautils.ExtractIncoming(ctx)

	e.Debug("A new message was received",
		zap.String("method", "EchoService.SubmitCall"),
		zap.Any("headers", headers),
		zap.Any("message", m),
	)

	return m, nil
}
