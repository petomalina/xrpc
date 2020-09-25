package multiplexer

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/suite"
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
		PubSubHTTPHandler(gateway),
		//pubSubGRPCHandler(grpcServer),
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
	reqBody, _ := json.Marshal(goldenPubSubMessage)
	req := &http.Request{
		Method: http.MethodPost,
		URL:    testingTargetEndpoint,
		Body:   ioutil.NopCloser(bytes.NewReader(reqBody)),
		Header: map[string][]string{
			"User-Agent":   {"APIs-Google; (+https://developers.google.com/webmasters/APIs-Google.html)"},
			"Content-Type": {"application/json"},
		},
	}
	client := &http.Client{}

	res, err := client.Do(req)
	s.NoError(err)
	s.NotNil(res)
	s.Equal(http.StatusOK, res.StatusCode)

	bb, err := ioutil.ReadAll(res.Body)
	s.NoError(err)
	s.Equal(string(goldenPubSubMessageData), string(bb))
}

func TestPubSubHandlerSuite(t *testing.T) {
	suite.Run(t, &PubSubHandlerSuite{})
}
