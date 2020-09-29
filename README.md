# :rocket: xrpc

xrpc is a small library used to **multiplex GRPC, HTTP & PubSub traffic on a single port**.
The implementation allows multiple servers to act as one. It makes it easy to run e.g. GRPC and grpc-gateway (JSON transcoder)
on a single port in a serverless environment.

## :wrench: Installation

```
go get github.com/petomalina/xrpc
```

## :book: Guide

Common use-cases are documented below, but creating new handlers and selectors for the multiplexer
is as easy as implementing a function.

The xrpc library works on a concept of *Multiplexer* with multiple *Handlers*. Every Handler has a
corresponding set of *Selectors*.

**Multiplexer** is the entrypoint for the traffic that is being served. You will want the *Multiplexer* to
server all incoming traffic to your server.

**Handler** is looks close to the official `type HandlerFunc func(ResponseWriter, *Request)`, however, it returns bool `type Handler func(ResponseWriter, *Request) bool`.
This behavior lets the *Multiplexer* know if the *Handler* is serving the request.

**Selector** is a filter. A Request must pass all Selectors to get handled by a corresponding *Handler*.

#### GRPC & Default HTTP

Let's start with a pure GRPC. This creates a new `EchoService` and registers it with the multiplexer. You can further
filter requests coming to the `grpcServer` by adding custom Selectors.

```go
// create and register the grpc server
grpcServer := grpc.NewServer()
echoSvc := &EchoService{}
api.RegisterEchoServiceServer(grpcServer, echoSvc)
reflection.Register(grpcServer)

// make the multiplexer
multiplexer.Make(nil,
    // filters all application/grpc messages into the grpc server
    multiplexer.GRPCHandler(grpcServer/*, customSelector*/),
)
```

> :bulb: Note that all unhandled messages will get "404 - no handler was fulfilled for your request"

#### GRPC & HTTP (grpc-gateway)

This example leverates the [grpc-ecosystem/grpc-gateway ](https://github.com/grpc-ecosystem/grpc-gateway) which transcodes GRPC
messages to their JSON equivalents and the other way around. The example is very similiar to the GRPC one, however,
we'll register the gateway along the GRPC handler.

```go
// create and register the grpc server
grpcServer := grpc.NewServer()
echoSvc := &EchoService{}
api.RegisterEchoServiceServer(grpcServer, echoSvc)
reflection.Register(grpcServer)

// create the grpc-gateway server and register to grpc server
gwmux := runtime.NewServeMux()
err = api.RegisterEchoServiceHandlerFromEndpoint(ctx, gwmux, ":"+os.Getenv("PORT"), []grpc.DialOption{grpc.WithInsecure()})
if err != nil {
    logger.Fatal("gw: failed to register: %v", zap.Error(err))
}

// make the multiplexer
multiplexer.Make(nil,
    // filters all application/grpc messages into the grpc server
    multiplexer.GRPCHandler(grpcServer),
    // defaults all other messages into the http multiplexer
    multiplexer.HTTPHandler(gwmux),
)
```

The full example can be found in the **examples/grpc-http** folder.

#### Pub/Sub

Lastly, we'll register a Pub/Sub push endpoint that will handle any Pub/Sub messages sent by subscriptions. No need to register
additional services, as Pub/Sub messages can be handled by the grpc-gateway.

```go
// make the multiplexer
multiplexer.Make(nil,
    // filters all application/grpc messages into the grpc server
    multiplexer.GRPCHandler(grpcServer),
    // filters all messages with Google Agent into the gwmux and
    // unpacks the PubSub message
    multiplexer.PubSubHandler(gwmux),
+    // defaults all other messages into the http multiplexer
+    multiplexer.HTTPHandler(gwmux),
)
```

The full example can be found in the **examples/grpc-http-pubsub** folder.

### Creating Custom Selectors

There are times, when we need to handle specific cases (e.g. all requests to a certain server must contain some header).
The xrpc library has *Selectors* exactly for this use-case. Every handler accepts a list of selectors as the second parameter.

```go
func IsGoogleRobot(r *http.Request) bool {
	return strings.Contains(r.Header.Get("User-Agent"), "GoogleBot")
}

func main() {
    // ...

    multiplexer.Make(nil,
        multiplexer.HTTPHandler(myRobotHandler, IsGoogleRobot)
        // register other handlers
    )

    // ...
}
```

This example creates a new *Selector* that will be executed as a part of the filtration mechanism. Any requests
containing the User-Agent header with "GoogleBot" in them will be redirected to the `myBotHandler`.