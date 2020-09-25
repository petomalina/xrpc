package multiplexer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/petomalina/xrpc/examples/api"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/proto"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
)

type PubSubHandlerSuite struct {
	suite.Suite

	srv         *http.Server
	echoService *EchoService
}

func (s *PubSubHandlerSuite) SetupTest() {
	s.echoService = &EchoService{createLogger(), nil}
	grpcServer := createGrpcServer(s.echoService)
	gateway := createGrpcGatewayServer()

	s.srv = createTestServer(
		PubSubHandler(map[string]Handler{
			AttributeEncodingHTTP: HTTPHandler(gateway),
			AttributeEncodingGRPC: GRPCHandler(grpcServer),
		}),
		GRPCHandler(grpcServer),
	)

	go func() {
		// an error is returned when the server is closed externally. This is normal
		s.Error(http.ErrServerClosed, s.srv.ListenAndServe(), "error listening")
	}()
}

func (s *PubSubHandlerSuite) TearDownTest() {
	s.NoError(s.srv.Close(), "error closing the server")
}

func (s *PubSubHandlerSuite) TestPubSubHTTPHandler() {
	reqBody, _ := json.Marshal(goldenPubSubMessageJSON)
	req := makePubSubRequest(reqBody, false)
	client := &http.Client{}

	//s.echoService.onCall = func(ctx context.Context, m *api.EchoMessage) {
	//	headers := metautils.ExtractIncoming(ctx)
	//	fmt.Println(headers)
	//}

	res, err := client.Do(req)
	s.NoError(err)
	s.NotNil(res)
	s.Equal(http.StatusOK, res.StatusCode)

	bb, err := ioutil.ReadAll(res.Body)
	s.NoError(err)
	s.Equal(string(goldenPubSubMessageData), string(bb))
}

func (s *PubSubHandlerSuite) TestPubSubGRPCHandler() {
	// create the protobuf echo message model
	msg := &api.EchoMessage{
		Message: "Hello World",
	}
	// marshal contents of the message
	bb, err := proto.Marshal(msg)
	s.NoError(err)

	// marshal around the PubSub push message
	reqBody, err := json.Marshal(makeGoldenPubSubMessage(bb, AttributeEncodingGRPC))
	s.NoError(err)

	// create the http request with PubSUb message and the data within
	req := makePubSubRequest(reqBody, true)
	client := &http.Client{}

	s.echoService.onCall = func(ctx context.Context, m *api.EchoMessage) {
		headers := metautils.ExtractIncoming(ctx)
		fmt.Println("HEADERS:", headers)
	}

	res, err := client.Do(req)
	fmt.Printf("%+v\n", res)
	s.NoError(err)
	s.NotNil(res)
	s.Equal(http.StatusOK, res.StatusCode)
}

func makePubSubRequest(body []byte, grpc bool) *http.Request {
	uri := "http://localhost:" + testingPort
	if grpc {
		uri += "/api.EchoService/Call"
	} else {
		uri += "/echo"
	}

	u, _ := url.Parse(uri)

	return &http.Request{
		Method: http.MethodPost,
		URL:    u,
		Body:   ioutil.NopCloser(bytes.NewReader(body)),
		Header: map[string][]string{
			"User-Agent":   {"APIs-Google; (+https://developers.google.com/webmasters/APIs-Google.html)"},
			"Content-Type": {"application/json"},
		},
	}
}

func TestPubSubHandlerSuite(t *testing.T) {
	suite.Run(t, &PubSubHandlerSuite{})
}
