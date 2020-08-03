package ws

import (
	"encoding/json"
	"strings"
)

type Message string

func (m *Message) Unmarshal(v interface{}) error {
	return json.Unmarshal([]byte(*m), &v)
}

func parsePayload(p []byte) (string, Message) {
	msg := string(p)
	var name string
	if strings.Contains(msg, "\n") {
		mc := strings.Split(msg, "\n")
		if strings.HasPrefix(mc[0], "@@") {
			name = mc[0][2:]
			msg = strings.Join(mc[1:], "\n")
		}
	}
	return name, Message(msg)
}
