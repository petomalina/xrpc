package multiplexer

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/petomalina/xrpc/examples/api"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"testing"
)

type PubSubHandlerSuite struct {
	suite.Suite

	srv         *http.Server
	port        string
	echoService *EchoService
}

func (s *PubSubHandlerSuite) SetupTest() {
	lis, err := net.Listen("tcp", ":0")
	s.NoError(err)

	s.port = strconv.Itoa(lis.Addr().(*net.TCPAddr).Port)

	s.echoService = &EchoService{Logger: createLogger()}
	grpcServer := createGrpcServer(s.echoService)
	gateway := createGrpcGatewayServer(s.port)

	s.srv = createTestServer(
		PubSubHandler(gateway),
		GRPCHandler(grpcServer),
	)

	go func() {
		// an error is returned when the server is closed externally. This is normal
		s.Error(http.ErrServerClosed, s.srv.Serve(lis), "error listening")
	}()
}

func (s *PubSubHandlerSuite) TearDownTest() {
	s.NoError(s.srv.Close(), "error closing the server")
}

type responseError struct {
	res *http.Response
	err error

	callQueryHeader map[PubSubQueryParam]string
}

func mustURL(u string) *url.URL {
	parsed, err := url.Parse(u)
	if err != nil {
		panic(err)
	}

	return parsed
}

func (s *PubSubHandlerSuite) TestPubSubHandler() {
	client := &http.Client{}

	reqBody, err := json.Marshal(goldenPubSubMessageJSON)
	s.NoError(err)

	candidates := map[*http.Request]responseError{
		&http.Request{
			Method: http.MethodPost,
			URL:    mustURL("http://localhost:" + s.port + "/echo"),
			Body:   ioutil.NopCloser(bytes.NewReader(reqBody)),
			Header: map[string][]string{
				"User-Agent":   {"APIs-Google; (+https://developers.google.com/webmasters/APIs-Google.html)"},
				"Content-Type": {"application/json"},
			},
		}: {
			res: &http.Response{
				StatusCode: http.StatusOK,
			},
			err: nil,
		},
		&http.Request{
			Method: http.MethodPost,
			URL:    mustURL("http://localhost:" + s.port + "/echo?token=a12345&authorization=Bearer%20abcdefgh"),
			Body:   ioutil.NopCloser(bytes.NewReader(reqBody)),
			Header: map[string][]string{
				"User-Agent":   {"APIs-Google; (+https://developers.google.com/webmasters/APIs-Google.html)"},
				"Content-Type": {"application/json"},
			},
		}: {
			res: &http.Response{
				StatusCode: http.StatusOK,
			},
			callQueryHeader: map[PubSubQueryParam]string{
				PubSubQueryToken:         "a12345",
				PubSubQueryAuthorization: "Bearer abcdefgh",
			},
			err: nil,
		},
	}

	for req, res := range candidates {
		clientRes, err := client.Do(req)

		s.Equal(res.err, err)
		s.Equal(res.res.StatusCode, clientRes.StatusCode, "Request:", req.URL.String())

		s.echoService.onCall = func(ctx context.Context, m *api.EchoMessage) {
			headers := metautils.ExtractIncoming(ctx)

			for hk, hv := range res.callQueryHeader {
				s.Equal(hv, headers.Get(PubSubQueryHeader(hk)))
			}
		}
	}
}

func TestPubSubHandlerSuite(t *testing.T) {
	suite.Run(t, &PubSubHandlerSuite{})
}
