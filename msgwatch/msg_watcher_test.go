package msgwatch

import (
	"reflect"
	"testing"
	"time"
)

func testReceive(t *testing.T, c chan map[string]interface{}, m map[string]interface{}) {
	select {
	case x := <-c:
		if !reflect.DeepEqual(x, m) {
			t.Error("Maps arent equal", x, m)
		}
	case <-time.After(5 * time.Millisecond):
		t.Error("No message received")
	}
}

func testNotReceived(t *testing.T, c chan map[string]interface{}) {
	select {
	case <-c:
		t.Error("Message received erroneously")
	case <-time.After(5 * time.Millisecond):
	}
}

func TestMsgWatcher(t *testing.T) {
	in := make(chan map[string]interface{})
	m := New(in)

	subscribe := m.Listen("subscribe")
	subscribe2 := m.Listen("subscribe")
	ticker := m.Listen("ticker")

	s1 := map[string]interface{}{"op": "subscribe", "yoyo": "mama"}
	s2 := map[string]interface{}{"op": "subscribe", "foo": "bar"}
	n1 := map[string]interface{}{"op": "haxing"}
	t1 := map[string]interface{}{"op": "private", "private": "ticker", "stuff": "good"}
	t2 := map[string]interface{}{"op": "private", "private": "ticker", "happy": "baby"}

	in <- s1
	testReceive(t, subscribe, s1)
	testReceive(t, subscribe2, s1)
	in <- t1
	testReceive(t, ticker, t1)
	in <- n1
	in <- s2
	testReceive(t, subscribe, s2)
	testReceive(t, subscribe2, s2)
	in <- t2
	testReceive(t, ticker, t2)
	testNotReceived(t, subscribe)
	testNotReceived(t, subscribe2)
	testNotReceived(t, ticker)
}

func TestChannelMonitoring(t *testing.T) {
	in := make(chan map[string]interface{})
	m := New(in)

	ticker := m.Listen("ticker")
	ch := "ABC-123-DEF-456"

	t1 := map[string]interface{}{"op": "private", "private": "ticker", "stuff": "good", "channel": ch}
	in <- t1

	testReceive(t, ticker, t1)
	if m.Channels["ticker"] != ch {
		t.Log(ch, m.Channels)
		t.Fail()
	}
}
