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
	goldenPubSubMessageJSON = makeGoldenPubSubMessage(goldenPubSubMessageData, "")
)

func makeGoldenPubSubMessage(data []byte, encoding string) PushMessage {
	return PushMessage{
		Message: &PubSubMessage{
			Data: data,
			Attributes: map[string]string{
				"my-label": "this-is-value",
				"encoding": encoding,
			},
			MessageID:   "abc12345",
			PublishTime: "2014-10-02T15:01:23Z",
			OrderingKey: "",
		},
		Subscription: "12345",
	}
}

func (s *PubSubTestSuite) TestIsPubSubRequest() {
	candidates := map[*http.Request]bool{
		{
			Method: http.MethodPost,
			Header: map[string][]string{
				"User-Agent": {"APIs-Google; (+https://developers.google.com/webmasters/APIs-Google.html)"},
			},
		}: true,
		// PubSub has a specific User-Agent bound to it
		{
			Method: http.MethodPost,
			Header: map[string][]string{
				"User-Agent": {"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.121 Safari/537.36"},
			},
		}: false,
		// PubSub will always send a POST request
		{
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
	goldenMsgJSON, _ := json.Marshal(goldenPubSubMessageJSON)

	candidates := map[*http.Request]InterceptedResult{
		{
			URL:    mustURL("http://localhost"),
			Body:   ioutil.NopCloser(bytes.NewBuffer(goldenMsgJSON)),
			Header: map[string][]string{},
		}: {
			Body: "{\"message\":\"Hello World\"}",
			Headers: map[string]string{
				"Grpc-Metadata-x-pubsub-subscription":  goldenPubSubMessageJSON.Subscription,
				"Grpc-Metadata-x-pubsub-message-id":    goldenPubSubMessageJSON.Message.MessageID,
				"Grpc-Metadata-x-pubsub-publish-time":  goldenPubSubMessageJSON.Message.PublishTime,
				"Grpc-Metadata-x-pubsub-attr-my-label": goldenPubSubMessageJSON.Message.Attributes["my-label"],
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
