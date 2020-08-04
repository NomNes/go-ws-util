package ws

import (
	"encoding/json"
	"io"
	"net"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/kataras/go-events"
)

type Extra = map[string]interface{}

type Connection struct {
	events.EventEmmiter
	conn   net.Conn
	state  ws.State
	close  chan error
	closed bool
	Extra
}

func newConnection(conn net.Conn, s ws.State) *Connection {
	return &Connection{
		EventEmmiter: events.New(),
		conn:         conn,
		state:        s,
		close:        make(chan error),
	}
}

func (c *Connection) Close(err error) {
	c.closed = true
	c.Emit("close", err)
	go func() { c.close <- err }()
	var reason []byte
	if err != nil {
		reason = []byte(err.Error())
	}
	_ = wsutil.WriteServerMessage(c.conn, ws.OpClose, reason)
	_ = c.conn.Close()
}

func (c *Connection) Listen() error {
	go func() {
		var err error
		defer c.Close(err)
		for {
			msg, err := wsutil.ReadMessage(c.conn, c.state, nil)
			if err != nil {
				if err == io.EOF {
					return
				}
				c.Emit("messageError", err)
				continue
			}
			for _, m := range msg {
				if m.OpCode.IsControl() {
					err = wsutil.HandleControlMessage(c.conn, c.state, m)
					if err != nil {
						if ce, ok := err.(wsutil.ClosedError); ok {
							err = ce
							return
						}
						c.Emit("messageError", err)
					}
					continue
				}
				name, message := parsePayload(m.Payload)
				c.Emit("message", name, message)
				c.Emit(events.EventName("message "+name), message)
			}
		}
	}()
	return <-c.close
}

func (c *Connection) OnError(name string, handler func(err error)) {
	if name != "" {
		c.On(events.EventName(name+"Error"), wrapErrorHandler(handler))
	}
}

func (c *Connection) OnMessage(handler func(name string, msg Message)) {
	c.On("message", func(i ...interface{}) {
		handler(i[0].(string), i[1].(Message))
	})
}

func (c *Connection) OnMessageOnce(handler func(name string, msg Message)) {
	c.Once("message", func(i ...interface{}) {
		handler(i[0].(string), i[1].(Message))
	})
}

func (c *Connection) OnNamedMessage(name string, handler func(msg Message)) {
	c.On(events.EventName("message "+name), func(i ...interface{}) {
		handler(i[0].(Message))
	})
}

func (c *Connection) OnNamedMessageOnce(name string, handler func(msg Message)) {
	c.Once(events.EventName("message "+name), func(i ...interface{}) {
		handler(i[0].(Message))
	})
}

func (c *Connection) SendMessage(name string, msg string) error {
	if name != "" {
		msg = "@@" + name + "\n" + msg
	}
	return wsutil.WriteMessage(c.conn, c.state, ws.OpText, []byte(msg))
}

func (c *Connection) SendJson(name string, v interface{}) error {
	msg, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.SendMessage(name, string(msg))
}
