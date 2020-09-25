package multiplexer

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"google.golang.org/grpc"
	"io/ioutil"
	"net/http"
	"strings"
)

// IsPubSubRequest returns true if the given request is considered
// to be made by Google servers and thus pushed by Pub/Sub service
func IsPubSubRequest(r *http.Request) bool {
	return strings.Contains(r.Header.Get("user-agent"), "APIs-Google") && r.Method == http.MethodPost
}

// IsPubSubGRPCRequest returns true if the given request is considered
// to be a GRPC request and PubSub request at the same time
func IsPubSubGRPCRequest(r *http.Request) bool {
	return IsGRPCRequest(r) && IsPubSubRequest(r)
}

// PubSubHTTPHandler fulfills requests that are considered to be PubSub requests,
// automatically unwrapping their bodies and appending metadata as headers
func PubSubHTTPHandler(handler http.Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request) bool {
		if !IsPubSubRequest(r) {
			return false
		}

		req, err := InterceptPubSubHTTP(r)
		if err != nil {
			_, _ = w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			return true
		}

		handler.ServeHTTP(w, req)
		return true
	}
}

// not working yet
func pubSubGRPCHandler(conn *grpc.Server, opts ...grpc.CallOption) Handler {
	return func(w http.ResponseWriter, r *http.Request) bool {
		if !IsPubSubGRPCRequest(r) {
			return false
		}

		_, err := InterceptPubSubGRPC(r)
		if err != nil {
			_, _ = w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusBadRequest)
		}

		//err = conn.Invoke(r.Context(), r.URL.Path, &body, nil, opts...)
		//if err != nil {
		//	_, _ = w.Write([]byte(err.Error()))
		//	w.WriteHeader(http.StatusBadRequest)
		//}
		return true
	}
}

// PushMessage is a definition of the Google PubSub message received as a push message
type PushMessage struct {
	Message      *PubSubMessage `json:"message,omitempty"`
	Subscription string         `json:"subscription,omitempty"`
}

// PubSubMessage is a definition of the internal PubSub message received on PubSub push
type PubSubMessage struct {
	Data        []byte            `json:"data,omitempty"`
	Attributes  map[string]string `json:"attributes,omitempty"`
	MessageID   string            `json:"messageId,omitempty"`
	PublishTime string            `json:"publishTime,omitempty"`
	OrderingKey string            `json:"orderingKey,omitempty"`
}

// InterceptPubSubHTTP mutates the given http.Request, reading its body and converting it
// into the PubSub body. it also adds all PubSub metadata into headers, refills the body
// with the PubSub data, and corrects the Content-Length header
func InterceptPubSubHTTP(r *http.Request) (*http.Request, error) {
	// handle Google APIs pushed events (PubSub)

	// read the contents of the http request (this will be replaced later)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	// unmarshal to the pubsub message
	psmsg := PushMessage{}
	err = json.Unmarshal(body, &psmsg)
	if err != nil {
		return nil, err
	}

	// get the body from the pubsub and re-create it
	psbody := psmsg.Message.Data

	r.Body = ioutil.NopCloser(bytes.NewBuffer(psbody))
	r.ContentLength = int64(len(psbody))

	// the Grpc-Metadata- prefix is stripped by the grpc-gateway, so these headers
	// are accessible by their original names
	r.Header.Add("Grpc-Metadata-x-pubsub-subscription", psmsg.Subscription)
	r.Header.Add("Grpc-Metadata-x-pubsub-message-id", psmsg.Message.MessageID)
	r.Header.Add("Grpc-Metadata-x-pubsub-message-pubslish-time", psmsg.Message.PublishTime)
	for k, v := range psmsg.Message.Attributes {
		r.Header.Add("Grpc-Metadata-x-pubsub-"+k, v)
	}

	return r, nil
}

// InterceptPubSubGRPC mutates an HTTP PubSub request and transforms it into a GRPC request
// which can be fulfilled by a GRPC server.
func InterceptPubSubGRPC(r *http.Request) ([]byte, error) {
	// handle Google APIs pushed events (PubSub)

	// read the contents of the http request (this will be replaced later)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	// unmarshal to the pubsub message
	psmsg := PushMessage{}
	err = json.Unmarshal(body, &psmsg)
	if err != nil {
		return nil, err
	}

	// get the body from the pubsub and re-create it
	psbody, err := base64.StdEncoding.DecodeString(string(psmsg.Message.Data))
	if err != nil {
		return nil, err
	}

	return psbody, nil
}
