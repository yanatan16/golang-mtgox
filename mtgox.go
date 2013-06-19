package mtgox

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	"strings"
	"errors"
	"encoding/base64"
	"encoding/json"
)

const (
	api_host   string = "websocket.mtgox.com"
	api_path   string = "/mtgox"
	http_endpoint string = "http://mtgox.com/api/2"
	origin_url string = "http://localhost"
)

var (
	// TODO: use https://mtgox.com/api/2/stream/list_public
	channels map[string]string = make(map[string]string)
	randStr chan string
)

type StreamingApi struct {
	*MsgWatcher
	ws       *JsonWebsocket
	Key string
}

type Sender interface {
	Send(map[string]interface{}) error
}

func NewStreamingApi(currencies ...string) (*StreamingApi, error) {
	url := fmt.Sprintf("ws://%s%s?Currency=%s", api_host, api_path, strings.Join(currencies, ","))
	config, _ := websocket.NewConfig(url, origin_url)
	ws, err := NewJsonWebsocket(config)

	if err != nil {
		return nil, err
	}

	api := &StreamingApi{
		MsgWatcher: NewMsgWatcher(ws.RecvForever(), "op"),
		ws:       ws,
	}

	return api, nil
}

func (api *StreamingApi) Send(msg map[string]interface{}) error {
	return api.ws.Send(msg)
}

func (api *StreamingApi) Close() error {
	return api.ws.Close()
}

func (api *StreamingApi) Unsubscribe(name string) error {
	if key, ok := channels[name]; ok {
		return api.Send(map[string]interface{}{
			"op": "unsubscribe",
			"channel": key,
		})
	}

	return errors.New("No channel with name " + name + ".")
}

func (api *StreamingApi) Subscribe(typ string) error {
	return api.Send(map[string]interface{}{
		"op": "mtgox.subscribe",
		"type": typ,
	})
}

func (api *StreamingApi) Sign(body []byte) (string, error) {
	return "", nil
}

func (api *StreamingApi) AuthenticatedSend(msg map[string]interface{}) error {
	req, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	signedReq, err := api.Sign(req)
	if err != nil {
		return err
	}

	reqid := <- randStr

	fullReq := api.Key + signedReq + string(req)
	var encodedReq []byte
	base64.StdEncoding.Encode(encodedReq, []byte(fullReq))

	return api.Send(map[string]interface{}{
		"op": "call",
		"id": reqid,
		"call": encodedReq,
		"context": "mtgox.com",
	})
}