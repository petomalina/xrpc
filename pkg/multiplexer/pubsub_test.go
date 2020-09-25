package multiplexer

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"net/http"
	"testing"
)

type PubSubTestSuite struct {
	suite.Suite
}

var (
	goldenPubSubMessageData = []byte("{\"message\":\"Hello World\"}")
	goldenPubSubMessage     = PushMessage{
		Message: &PubSubMessage{
			Data: goldenPubSubMessageData,
			Attributes: map[string]string{
				"my-label": "this-is-value",
			},
			MessageID:   "abc12345",
			PublishTime: "2014-10-02T15:01:23Z",
			OrderingKey: "",
		},
		Subscription: "12345",
	}
)

func (s *PubSubTestSuite) TestIsGRPCRequest() {
	candidates := map[*http.Request]bool{
		&http.Request{
			Method: http.MethodPost,
			Header: map[string][]string{
				"User-Agent": {"APIs-Google; (+https://developers.google.com/webmasters/APIs-Google.html)"},
			},
		}: true,
		// PubSub has a specific User-Agent bound to it
		&http.Request{
			Method: http.MethodPost,
			Header: map[string][]string{
				"User-Agent": {"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.121 Safari/537.36"},
			},
		}: false,
		// PubSub will always send a POST request
		&http.Request{
			Method: http.MethodGet,
			Header: map[string][]string{
				"User-Agent": {"APIs-Google; (+https://developers.google.com/webmasters/APIs-Google.html)"},
			},
		}: false,
	}

	for req, result := range candidates {
		s.Equal(result, IsPubSubRequest(req), "Request is badly considered a grpc/non-grpc request", req.Header)
	}
}

type InterceptedResult struct {
	Body    string
	Headers map[string]string
	Error   error
}

func (s *PubSubTestSuite) TestInterceptPubSubRequest() {
	goldenMsgJson, _ := json.Marshal(goldenPubSubMessage)

	candidates := map[*http.Request]InterceptedResult{
		&http.Request{
			Body:   ioutil.NopCloser(bytes.NewBuffer(goldenMsgJson)),
			Header: map[string][]string{},
		}: {
			Body: "{\"message\":\"Hello World\"}",
			Headers: map[string]string{
				"Grpc-Metadata-x-pubsub-subscription":         goldenPubSubMessage.Subscription,
				"Grpc-Metadata-x-pubsub-message-id":           goldenPubSubMessage.Message.MessageID,
				"Grpc-Metadata-x-pubsub-message-publish-time": goldenPubSubMessage.Message.PublishTime,
				"Grpc-Metadata-x-pubsub-my-label":             goldenPubSubMessage.Message.Attributes["my-label"],
			},
		},
	}

	for req, result := range candidates {
		intercepted, err := InterceptPubSubRequest(req)
		s.Equal(result.Error, err)

		interceptedBody, err := ioutil.ReadAll(intercepted.Body)
		s.NoError(err)
		s.Equal(result.Body, string(interceptedBody))

		for k, v := range result.Headers {
			s.Equal(v, intercepted.Header.Get(k))
		}
	}
}

func TestPubSubTestSuite(t *testing.T) {
	suite.Run(t, &PubSubTestSuite{})
}
