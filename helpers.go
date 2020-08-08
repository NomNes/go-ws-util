package ws

import (
	"sync"

	"github.com/kataras/go-events"
)

func wrapErrorHandler(handler func(error)) events.Listener {
	return func(i ...interface{}) {
		if i[0] == nil {
			handler(nil)
		} else {
			handler(i[0].(error))
		}
	}
}

type ConnMap struct {
	sync.Mutex
	items map[*Connection]bool
}

func (cm *ConnMap) init() {
	if cm.items == nil {
		cm.items = map[*Connection]bool{}
	}
}

func (cm *ConnMap) Add(c *Connection) {
	cm.Lock()
	defer cm.Unlock()
	cm.init()
	cm.items[c] = true
}

func (cm *ConnMap) Delete(c *Connection) {
	cm.Lock()
	defer cm.Unlock()
	cm.init()
	delete(cm.items, c)
}

func (cm *ConnMap) Len() int {
	cm.Lock()
	defer cm.Unlock()
	cm.init()
	return len(cm.items)
}

func (cm *ConnMap) Has(c *Connection) bool {
	cm.Lock()
	defer cm.Unlock()
	cm.init()
	_, ok := cm.items[c]
	return ok
}

func (cm *ConnMap) Items() []*Connection {
	cm.Lock()
	defer cm.Unlock()
	cm.init()
	var items []*Connection
	for i := range cm.items {
		items = append(items, i)
	}
	return items
}
