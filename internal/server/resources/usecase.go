package resources

import (
	"context"

	"github.com/gorilla/websocket"
	"github.com/vlasashk/websocket-chat/pkg/response"
)

type ClientManager interface {
	Store(con *websocket.Conn)
	Release(con *websocket.Conn)
	Broadcaster(ctx context.Context) chan<- response.Msg
	WriteMsg(con *websocket.Conn, msg response.Msg) error
}

type CacheRepo interface {
	AddMessage(ctx context.Context, data []byte) error
	GetLastTen(ctx context.Context) ([]response.Msg, error)
}

type MessageBroker interface {
	Write(ctx context.Context, data []byte)
}
