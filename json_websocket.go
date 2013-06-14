package mtgox

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
)

type JsonWebsocket struct {
	ws *websocket.Conn
}

func NewJsonWebsocket(cfg *websocket.Config) (*JsonWebsocket, error) {
	ws, err := websocket.DialConfig(cfg)
	if err != nil {
		return nil, err
	}

	return &JsonWebsocket{ws}, nil
}

func (jws *JsonWebsocket) Recv() (map[string]interface{}, error) {
	obj := make(map[string]interface{})
	err := websocket.JSON.Receive(jws.ws, &obj)
	return obj, err
}

func (jws *JsonWebsocket) Send(obj map[string]interface{}) error {
	return websocket.JSON.Send(jws.ws, obj)
}

func (jws *JsonWebsocket) Close() error {
	return jws.ws.Close()
}

func (ws *JsonWebsocket) RecvForever() chan map[string]interface{} {
	msgs := make(chan map[string]interface{})
	go func() {
		for {
			obj, err := ws.Recv()
			if err != nil {
				if err.Error() == "use of closed network connection" {
					break
				}
				panic(err)
			}

			fmt.Println("Recieved message", obj)

			msgs <- obj
		}
	}()
	return msgs
}
