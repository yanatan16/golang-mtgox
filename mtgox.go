package mtgox

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
)

const (
	api_host   string = "websocket.mtgox.com"
	api_path   string = "/mtgox"
	origin_url string = "http://github.com/yanatan16"
)

var (
	// TODO: use https://mtgox.com/api/2/stream/list_public
	channels map[string]string = map[string]string{
		"ticker": "d5f06780-30a8-4a48-a2f8-7ed181b4a13f",
		"depth":  "24e67e0d-1cad-4cc0-9e7a-f8523ef460fe",
		"trades": "dbf1dee9-4f2e-4a08-8cb7-748919a71b21",
	}
)

type StreamingApi struct {
	ws       *JsonWebsocket
	Messages chan map[string]interface{}
}

func NewStreamingApi(currencies string) (*StreamingApi, error) {
	config, _ := websocket.NewConfig(fmt.Sprintf("ws://%s%s?Currency=%s", api_host, api_path, currencies), "http://localhost")
	ws, err := NewJsonWebsocket(config)

	if err != nil {
		return nil, err
	}

	api := &StreamingApi{
		ws:       ws,
		Messages: ws.RecvForever(),
	}

	return api, nil
}

func (api *StreamingApi) Send(msg map[string]interface{}) error {
	return api.ws.Send(msg)
}

func (api *StreamingApi) Close() error {
	close(api.Messages)
	return api.ws.Close()
}
