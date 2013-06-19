package mtgox

import (
	"log"
)

type MsgWatcher struct {
	listeners map[string] []chan map[string]interface{}
	in chan map[string]interface{}
	keyfield string
}

func NewMsgWatcher(in chan map[string]interface{}, keyfield string) *MsgWatcher {
	m := &MsgWatcher{
		listeners: make(map[string] []chan map[string]interface{}),
		in: in,
		keyfield: keyfield,
	}
	go m.listen()
	return m
}

func (m *MsgWatcher) Listen(typ string) chan map[string]interface{} {
	c := make(chan map[string]interface{}, 2)
	m.listeners[typ] = append(m.listeners[typ], c)
	return c
}

func (m *MsgWatcher) close() {
	for _, ls := range m.listeners {
		for _, c := range ls {
			close(c)
		}
	}
}

func (m *MsgWatcher) listen() {
	for msg := range m.in {
		if key, ok := msg[m.keyfield]; ok {
			if key2, ok := msg[key.(string)]; ok {
				key = key2
			}

			if ls, ok := m.listeners[key.(string)]; ok {
				for _, l := range ls {
					select {
					case l <- msg:
					default:
					}
				}
			} else {
				log.Println("No listeners found for", key)
			}
		} else {
			log.Println("No field " + m.keyfield + " in message.", msg)
		}
	}
	m.close()
}