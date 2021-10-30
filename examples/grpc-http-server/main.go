package main

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/petomalina/xrpc/v2/examples/api"
	"github.com/petomalina/xrpc/v2/pkg/multiplexer"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	// create the zap logger for future use
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	defer func() {
		done()
		if r := recover(); r != nil {
			logger.Fatal("panic recovered, exiting", zap.Any("panic", r))
		}
	}()

	// create and register the grpc server
	grpcServer := grpc.NewServer()
	echoSvc := &EchoService{Logger: logger}
	api.RegisterEchoServiceServer(grpcServer, echoSvc)
	reflection.Register(grpcServer)

	// create the grpc-gateway server and register to grpc server
	gwmux := runtime.NewServeMux()
	err = api.RegisterEchoServiceHandlerFromEndpoint(ctx, gwmux, ":"+os.Getenv("PORT"), []grpc.DialOption{grpc.WithInsecure()})
	if err != nil {
		logger.Fatal("gw: failed to register", zap.Error(err))
	}

	srv, err := multiplexer.NewServer(os.Getenv("PORT"), time.Second*5)
	if err != nil {
		logger.Fatal("could not create a server", zap.Error(err))
	}

	// make multiplexer
	err = srv.ServeHTTPHandler(ctx,
		multiplexer.Make(nil,
			// filters all application/grpc messages into the grpc server
			multiplexer.GRPCHandler(grpcServer),
			// defaults all other messages into the http multiplexer
			multiplexer.HTTPHandler(gwmux),
		))
	done()

	if err != nil {
		logger.Fatal("error serving the server", zap.Error(err))
	}
	logger.Info("Shutting down")
}

// EchoService is the example service
type EchoService struct {
	*zap.Logger

	api.UnimplementedEchoServiceServer
}

// Call logs the message and returns it back
func (e *EchoService) Call(ctx context.Context, m *api.EchoMessage) (*api.EchoMessage, error) {
	headers := metautils.ExtractIncoming(ctx)

	e.Info("A new message was received",
		zap.String("method", "EchoService.SubmitCall"),
		zap.Any("headers", headers),
		zap.Any("message", m),
	)

	return m, nil
}
