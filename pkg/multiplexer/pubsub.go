package multiplexer

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

// IsPubSubRequest returns true if the given request is considered
// to be made by Google servers and thus pushed by Pub/Sub service
func IsPubSubRequest(r *http.Request) bool {
	return strings.Contains(r.Header.Get("user-agent"), "APIs-Google") && r.Method == http.MethodPost
}

// PubSubHandler fulfills requests that are considered to be PubSub requests,
// automatically unwrapping their bodies and appending metadata as headers
func PubSubHandler(handler http.Handler, selectors ...Selector) Handler {
	filter := append([]Selector{IsPubSubRequest}, selectors...)

	return func(w http.ResponseWriter, r *http.Request) bool {
		for _, f := range filter {
			if !f(r) {
				return false
			}
		}

		req, err := InterceptPubSubRequest(r)
		if err != nil {
			_, _ = w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			return true
		}

		handler.ServeHTTP(w, req)
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

type PubSubMetaAttribute string

const (
	PubSubMetaSubscription PubSubMetaAttribute = "subscription"
	PubSubMetaMessageID                        = "message-id"
	PubSubMetaPublishTime                      = "publish-time"
)

type PubSubQueryParam string

const (
	PubSubQueryToken         = "token"
	PubSubQueryAuthorization = "authorization"
)

// PubSubMetaAttributeHeader creates a key for header access to the PubSUb
// meta information (see PubSubMetaAttribute)
func PubSubMetaAttributeHeader(attr PubSubMetaAttribute) string {
	return "x-pubsub-" + string(attr)
}

// PubSubAttributeHeader returns a key for header access to the PubSub message
// attributes (PubSubMessage.Attributes)
func PubSubAttributeHeader(attr string) string {
	return "x-pubsub-attr-" + attr
}

// PubSubQueryHeader returns a key for header access to query parameters passed by
// the PubSub Request, such as token or Authorization
func PubSubQueryHeader(param PubSubQueryParam) string {
	return "x-query-" + string(param)
}

// InterceptPubSubRequest mutates the given http.Request, reading its body and converting it
// into the PubSub body. it also adds all PubSub metadata into headers, refills the body
// with the PubSub data, and corrects the Content-Length header
func InterceptPubSubRequest(r *http.Request) (*http.Request, error) {
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

	// headerPrefix is set to this prefix for GRPC gateway. The prefix is special and is stripped away
	// by the grpc gateway
	var headerPrefix = "Grpc-Metadata-"

	// the Grpc-Metadata- prefix is stripped by the grpc-gateway, so these headers
	// are accessible by their original names
	r.Header.Add(headerPrefix+PubSubMetaAttributeHeader(PubSubMetaSubscription), psmsg.Subscription)
	r.Header.Add(headerPrefix+PubSubMetaAttributeHeader(PubSubMetaMessageID), psmsg.Message.MessageID)
	r.Header.Add(headerPrefix+PubSubMetaAttributeHeader(PubSubMetaPublishTime), psmsg.Message.PublishTime)
	for k, v := range psmsg.Message.Attributes {
		r.Header.Add(headerPrefix+PubSubAttributeHeader(k), v)
	}

	// copy query parameters into headers as well
	for k, vv := range r.URL.Query() {
		for _, v := range vv {
			r.Header.Add(headerPrefix+PubSubQueryHeader(PubSubQueryParam(k)), v)
		}
	}

	return r, nil
}
