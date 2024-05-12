package listener

import (
	"context"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

// SocketListen Separate goroutine to listen WS connection, exposing channel for received data
// (was required to be able to release connection when gracefully shut down, since ReadMessage() is blocking)
func SocketListen(ctx context.Context, log zerolog.Logger, con *websocket.Conn) <-chan []byte {
	listen := make(chan []byte)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				mt, message, err := con.ReadMessage()
				if err != nil || mt == websocket.CloseMessage {
					close(listen)
					if err != nil {
						log.Error().Err(err).Send()
					}
					return
				}
				listen <- message
			}
		}
	}()
	return listen
}
