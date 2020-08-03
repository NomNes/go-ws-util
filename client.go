package ws

import (
	"context"
	"net/http"

	"github.com/gobwas/ws"
)

type Client struct {
	*Connection
	cancel context.CancelFunc
}

func Dial(url string, header http.Header) (*Client, error) {
	ctx, cancel := context.WithCancel(context.Background())
	c := &Client{
		cancel: cancel,
	}
	ws.DefaultDialer.Header = ws.HandshakeHeaderHTTP(header)
	conn, _, _, err := ws.DefaultDialer.Dial(ctx, url)
	if err != nil {
		return nil, err
	}
	c.Connection = newConnection(conn, ws.StateClientSide)
	c.On("close", func(...interface{}) {
		c.cancel()
	})
	return c, nil
}
