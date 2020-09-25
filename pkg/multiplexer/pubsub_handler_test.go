package multiplexer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/petomalina/xrpc/examples/api"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/proto"
	"io/ioutil"
	"net/http"
	"testing"
)

type PubSubHandlerSuite struct {
	suite.Suite

	srv *http.Server
}

func (s *PubSubHandlerSuite) SetupTest() {
	grpcServer := createGrpcServer(createLogger())
	gateway := createGrpcGatewayServer()

	s.srv = createTestServer(
		PubSubHandler(map[string]http.Handler{
			AttributeEncodingHTTP: gateway,
			AttributeEncodingGRPC: grpcServer,
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
	req := makePubSubRequest(reqBody)
	client := &http.Client{}

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
	req := makePubSubRequest(reqBody)
	client := &http.Client{}

	res, err := client.Do(req)
	fmt.Println(res)
	s.NoError(err)
	s.NotNil(res)
	s.Equal(http.StatusOK, res.StatusCode)
}

func makePubSubRequest(body []byte) *http.Request {
	return &http.Request{
		Method: http.MethodPost,
		URL:    testingTargetEndpoint,
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
