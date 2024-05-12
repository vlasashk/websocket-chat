package manager

import (
	"context"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/vlasashk/websocket-chat/pkg/response"
)

const workers = 5

type Manager struct {
	clients map[*websocket.Conn]struct{}
	// mu for sync access to clients
	mu *sync.RWMutex
	// wsMu for sync write operation to WS
	wsMu *sync.Mutex
	log  zerolog.Logger
}

func New(log zerolog.Logger) *Manager {
	return &Manager{
		clients: make(map[*websocket.Conn]struct{}),
		mu:      &sync.RWMutex{},
		wsMu:    &sync.Mutex{},
		log:     log,
	}
}

func (m *Manager) Store(con *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.clients[con] = struct{}{}
}

func (m *Manager) WriteMsg(con *websocket.Conn, msg response.Msg) error {
	m.wsMu.Lock()
	defer m.wsMu.Unlock()
	return con.WriteJSON(msg)
}

func (m *Manager) Release(con *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := con.Close(); err != nil {
		m.log.Error().Err(err).Msg("failed to close ws client")
	}
	delete(m.clients, con)
}

// Broadcaster Single run of broadcast worker pool, exposing channel to share among all clients (supposed to be called only once)
func (m *Manager) Broadcaster(ctx context.Context) chan<- response.Msg {
	data := make(chan response.Msg, workers)
	for w := 0; w < workers; w++ {
		go m.broadcast(ctx, data)
	}
	return data
}

func (m *Manager) broadcast(ctx context.Context, messages <-chan response.Msg) {
	for {
		select {
		case <-ctx.Done():
			m.log.Error().Msg("BroadCast ctx deadline")
			return
		case msg, ok := <-messages:
			if !ok {
				m.log.Error().Msg("BroadCast is dead")
				return
			}
			m.mu.RLock()
			for con := range m.clients {
				if err := m.WriteMsg(con, msg); err != nil {
					m.log.Error().Err(err).Send()
				}
			}
			m.mu.RUnlock()
		}
	}
}
