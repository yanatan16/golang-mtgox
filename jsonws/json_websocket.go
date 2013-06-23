package jsonws

import (
	"code.google.com/p/go.net/websocket"
)

type JsonWebsocket struct {
	ws   *websocket.Conn
	msgs chan map[string]interface{}
}

func New(cfg *websocket.Config) (*JsonWebsocket, error) {
	ws, err := websocket.DialConfig(cfg)
	if err != nil {
		return nil, err
	}

	return &JsonWebsocket{ws: ws}, nil
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
	if jws.msgs != nil {
		close(jws.msgs)
	}
	return jws.ws.Close()
}

func (ws *JsonWebsocket) RecvForever() chan map[string]interface{} {
	ws.msgs = make(chan map[string]interface{})
	go func() {
		for {
			obj, err := ws.Recv()
			if err != nil {
				if err.Error() == "use of closed network connection" {
					break
				}
				panic(err)
			}

			ws.msgs <- obj
		}
	}()
	return ws.msgs
}
