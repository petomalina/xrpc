module github.com/petomalina/xrpc/examples

go 1.15

require (
	github.com/blendle/zapdriver v1.3.1
	github.com/golang/protobuf v1.4.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/grpc-ecosystem/grpc-gateway v1.15.0
	github.com/petomalina/xrpc v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.16.0
	google.golang.org/genproto v0.0.0-20200923140941-5646d36feee1
	google.golang.org/grpc v1.32.0
	google.golang.org/protobuf v1.24.0
)

replace github.com/petomalina/xrpc => ../
