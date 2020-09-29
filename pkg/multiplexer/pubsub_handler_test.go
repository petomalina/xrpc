package multiplexer

import (
	"bytes"
	"encoding/json"
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

	s.echoService = &EchoService{createLogger(), nil}
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

func (s *PubSubHandlerSuite) TestPubSubHandler() {
	reqBody, _ := json.Marshal(goldenPubSubMessageJSON)
	req := makePubSubRequest(reqBody, s.port)
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

func makePubSubRequest(body []byte, port string) *http.Request {
	uri := "http://localhost:" + port
	uri += "/echo"

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
