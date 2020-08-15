package ws

import "sync"

type rooms struct {
	rooms map[string]*ConnMap
	sync.Mutex
}

func (r *rooms) init() {
	if r.rooms == nil {
		r.rooms = map[string]*ConnMap{}
	}
}

func (r *rooms) initRoom(name string) {
	r.init()
	if _, ok := r.rooms[name]; !ok {
		r.rooms[name] = &ConnMap{}
	}
}

func (r *rooms) delete(c *Connection) {
	r.Lock()
	defer r.Unlock()
	r.init()
	for _, m := range r.rooms {
		m.Delete(c)
	}
}

func (r *rooms) subscribe(name string, c *Connection) {
	r.Lock()
	defer r.Unlock()
	r.initRoom(name)
	r.rooms[name].Add(c)
}

func (r *rooms) unsubscribe(name string, c *Connection) {
	r.Lock()
	defer r.Unlock()
	r.initRoom(name)
	r.rooms[name].Delete(c)
}

func (r *rooms) broadcast(room, name string, msg string) {
	r.initRoom(room)
	for c := range r.rooms[room].items {
		_ = c.SendMessage(name, msg)
	}
}

func (r *rooms) broadcastJson(room, name string, v interface{}) {
	for c := range r.rooms[room].items {
		_ = c.SendJson(name, v)
	}
}
