package xpubsub

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type PushMessage struct {
	Message      *Message `json:"message,omitempty"`
	Subscription string   `json:"subscription,omitempty"`
}

type Message struct {
	Data        []byte            `json:"data,omitempty"`
	Attributes  map[string]string `json:"attributes,omitempty"`
	MessageID   string            `json:"messageId,omitempty"`
	PublishTime string            `json:"publishTime,omitempty"`
	OrderingKey string            `json:"orderingKey,omitempty"`
}

// InterceptHTTP mutates the given http.Request
func InterceptHTTP(r *http.Request) (*http.Request, error) {
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

//func InterceptGRPC(r *http.Request) ([]byte, error) {
//	// handle Google APIs pushed events (PubSub)
//
//	// read the contents of the http request (this will be replaced later)
//	body, err := ioutil.ReadAll(r.Body)
//	if err != nil {
//		return nil, err
//	}
//
//	// unmarshal to the pubsub message
//	psmsg := PushMessage{}
//	err = json.Unmarshal(body, &psmsg)
//	if err != nil {
//		return nil, err
//	}
//
//	// get the body from the pubsub and re-create it
//	psbody, err := base64.StdEncoding.DecodeString(string(psmsg.Message.Data))
//	if err != nil {
//		return nil, err
//	}
//
//	return psbody, nil
//}
