package jsonws

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"testing"
	"time"
)

var started bool

func ensure(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}

func startTimer(t *testing.T) chan bool {
	c := make(chan bool)
	go func() {
		select {
		case <-c:
		case <-time.After(1 * time.Second):
			t.Error("Timed out!")
		}
	}()
	return c
}

func JsonWsHandler(ws *websocket.Conn) {
	for {
		obj := make(map[string]interface{})
		err := websocket.JSON.Receive(ws, &obj)
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}

		fmt.Println("JsonWsHandler received", obj)

		obj["handler"] = "received"

		err = websocket.JSON.Send(ws, obj)
		if err != nil {
			panic(err)
		}
	}
}

func ensureJsonWebsocketServer() {
	if !started {
		http.Handle("/test", websocket.Handler(JsonWsHandler))
		go func() {
			err := http.ListenAndServe(":12345", nil)
			if err != nil {
				panic("ListenAndServe: " + err.Error())
			}
		}()
		<-time.After(50 * time.Millisecond)
		started = true
	}
}

func newJWS() *JsonWebsocket {
	cfg, err := websocket.NewConfig("ws://localhost:12345/test", "http://localhost")
	if err != nil {
		panic(err)
	}
	jws, err := New(cfg)
	if err != nil {
		panic(err)
	}
	return jws
}

func TestJsonWebsocketReceive(t *testing.T) {
	ensureJsonWebsocketServer()
	startTimer(t)
	jws := newJWS()
	ensure(t, jws.Send(map[string]interface{}{"client": "sent"}))

	obj, err := jws.Recv()
	ensure(t, err)

	if !reflect.DeepEqual(obj, map[string]interface{}{"client": "sent", "handler": "received"}) {
		t.Fail()
	}

	ensure(t, jws.Send(map[string]interface{}{"client": "sent2"}))
	<-time.After(20 * time.Millisecond)

	obj, err = jws.Recv()
	ensure(t, err)

	if !reflect.DeepEqual(obj, map[string]interface{}{"client": "sent2", "handler": "received"}) {
		t.Fail()
	}

	jws.Close()
}

func TestJsonWebsocketRecvForever(t *testing.T) {
	ensureJsonWebsocketServer()
	startTimer(t)
	jws := newJWS()
	c := jws.RecvForever()

	ensure(t, jws.Send(map[string]interface{}{"count": 1}))
	ensure(t, jws.Send(map[string]interface{}{"count": 2}))

	<-time.After(20 * time.Millisecond)
	select {
	case msg := <-c:
		if !reflect.DeepEqual(msg, map[string]interface{}{"count": float64(1), "handler": "received"}) {
			t.Error("First message didn't match")
		}
	default:
		t.Error("Message not received.")
	}

	<-time.After(20 * time.Millisecond)
	select {
	case msg := <-c:
		if !reflect.DeepEqual(msg, map[string]interface{}{"count": float64(2), "handler": "received"}) {
			t.Error("Second message didn't match")
		}
	default:
		t.Error("Second message not received.")
	}

	jws.Close()
}
