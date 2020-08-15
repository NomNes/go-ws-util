package ws

import (
	"net/http"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/kataras/go-events"
)

type HandshakeValidation = func(header http.Header) error
type ExtraHandler = func(header http.Header) Extra

type Server struct {
	addr         string
	connections  ConnMap
	PingDuration time.Duration
	close        chan error
	rooms
	HandshakeValidation
	ExtraHandler
	events.EventEmmiter
}

func NewServer(addr string) *Server {
	return &Server{
		addr:         addr,
		EventEmmiter: events.New(),
		close:        make(chan error),
	}
}

func (s *Server) Listen() error {
	server := &http.Server{
		Addr: s.addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var extra map[string]interface{}
			if s.HandshakeValidation != nil {
				err := s.HandshakeValidation(r.Header)
				if err != nil {
					s.Emit("handshakeValidationError", err)
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
			}
			if s.ExtraHandler != nil {
				extra = s.ExtraHandler(r.Header)
			}
			conn, _, _, err := ws.UpgradeHTTP(r, w)
			if err != nil {
				s.Emit("connectionError", err)
				return
			}
			c := newConnection(conn, ws.StateServerSide)
			c.Extra = extra
			s.connections.Add(c)
			c.On("close", func(i ...interface{}) {
				s.connections.Delete(c)
				s.rooms.delete(c)
				s.Emit("close", i...)
			})
			s.Emit("connection", c, s)
			go func() {
				for {
					if c.closed {
						return
					}
					err := wsutil.WriteServerMessage(conn, ws.OpPing, nil)
					if err != nil {
						c.Close(err)
					}
					if s.PingDuration > 0 {
						time.Sleep(s.PingDuration)
					} else {
						time.Sleep(time.Second)
					}
				}
			}()
			_ = c.Listen()
		}),
	}
	defer func() {
		_ = server.Close()
	}()
	go func() {
		s.close <- server.ListenAndServe()
	}()
	return <-s.close
}

func (s *Server) OnConnection(handler func(c *Connection, s *Server)) {
	s.On("connection", func(i ...interface{}) {
		handler(i[0].(*Connection), i[1].(*Server))
	})
}

func (s *Server) OnError(name string, handler func(err error)) {
	if name != "" {
		s.On(events.EventName(name+"Error"), wrapErrorHandler(handler))
	}
}

func (s *Server) Close(err error) {
	for c := range s.connections.items {
		c.Close(err)
	}
	go func() { s.close <- err }()
}

func (s *Server) Broadcast(name string, msg string) {
	for c := range s.connections.items {
		_ = c.SendMessage(name, msg)
	}
}

func (s *Server) BroadcastJson(name string, v interface{}) {
	for c := range s.connections.items {
		_ = c.SendJson(name, v)
	}
}

func (s *Server) Subscribe(room string, c *Connection) {
	s.subscribe(room, c)
}

func (s *Server) Unsubscribe(room string, c *Connection) {
	s.unsubscribe(room, c)
}

func (s *Server) BroadcastRoom(room, name string, msg string) {
	s.rooms.broadcast(room, name, msg)
}

func (s *Server) BroadcastRoomJson(room, name string, v interface{}) {
	s.rooms.broadcastJson(room, name, v)
}
