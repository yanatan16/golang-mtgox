package msgwatch

import (
	"log"
)

type MsgWatcher struct {
	Channels        map[string]string
	listeners       map[string][]chan map[string]interface{}
	in              chan map[string]interface{}
	resultListeners map[string]chan map[string]interface{}
	results         map[string]map[string]interface{}
}

func New(in chan map[string]interface{}) *MsgWatcher {
	m := &MsgWatcher{
		Channels:        make(map[string]string),
		listeners:       make(map[string][]chan map[string]interface{}),
		in:              in,
		resultListeners: make(map[string]chan map[string]interface{}),
		results:         make(map[string]map[string]interface{}),
	}
	go m.listen()
	return m
}

func (m *MsgWatcher) Listen(typ string) chan map[string]interface{} {
	c := make(chan map[string]interface{}, 2)
	m.listeners[typ] = append(m.listeners[typ], c)
	return c
}

func (m *MsgWatcher) ListenResult(id string) chan map[string]interface{} {
	c := make(chan map[string]interface{}, 1)
	if msg, ok := m.results[id]; ok {
		c <- msg
	} else {
		m.resultListeners[id] = c
	}
	return c
}

func (m *MsgWatcher) close() {
	for _, ls := range m.listeners {
		for _, c := range ls {
			close(c)
		}
	}
	for _, rl := range m.resultListeners {
		close(rl)
	}
}

func (m *MsgWatcher) receiveResult(msg map[string]interface{}) {
	if id, ok := msg["id"]; ok {
		if rl, ok := m.resultListeners[id.(string)]; ok {
			rl <- msg
			close(rl)
		} else {
			m.results[id.(string)] = msg
		}
	} else {
		log.Println("Error. Result message with no id field!?", msg)
	}
}

func (m *MsgWatcher) listen() {
	for msg := range m.in {
		if key, ok := msg["op"]; ok {
			if key == "private" {
				if key2, ok := msg["private"]; ok {
					key = key2
				}

				if channel, ok := msg["channel"]; ok {
					if schan, ok := channel.(string); ok {
						m.Channels[key.(string)] = schan
					}
				}
			} else if key == "result" {
				m.receiveResult(msg)
				continue
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
			log.Println("No op field in message.", msg)
		}
	}
	m.close()
}
