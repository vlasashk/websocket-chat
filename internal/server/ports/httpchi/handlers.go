package httpchi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
	"unicode/utf8"

	"github.com/go-chi/render"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/vlasashk/websocket-chat/config"
	"github.com/vlasashk/websocket-chat/internal/server/resources"
	"github.com/vlasashk/websocket-chat/pkg/listener"
	"github.com/vlasashk/websocket-chat/pkg/response"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(_ *http.Request) bool {
		return true
	},
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, render.M{
		"status": "ok",
	})
}

func EstablishWS(ctx context.Context, container *resources.Resources) http.HandlerFunc {
	broadcast := container.ClientManager.Broadcaster(ctx)
	return func(w http.ResponseWriter, r *http.Request) {
		log := container.Log.With().Caller().Logger()

		con, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Error().Err(err).Send()
			return
		}
		// Doesn't run in goroutine to be able to catch panic by chi router
		reader(ctx, con, container, broadcast)
	}
}

func reader(ctx context.Context, con *websocket.Conn, container *resources.Resources, broadcast chan<- response.Msg) {
	cm := container.ClientManager
	log := container.Log
	cache := container.RedisRepo
	kafkaWriter := container.KafkaWriter

	cm.Store(con)
	defer func() {
		cm.Release(con)
		log.Info().Msg("connection released")
	}()
	// Listens for first message from client that will indicate client's nickname
	userID, err := registerUser(ctx, con, container.Cfg.Storage)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}
	// Sends to client recent messages from chat (up to 10 messages)
	outputRecent(ctx, log, cache, con, cm)
	listen := listener.SocketListen(ctx, log, con)
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("context was canceled")
			return
		case data, ok := <-listen:

			if !ok {
				log.Error().Msg("connection died unexpectedly")
				return
			}

			var msg response.Msg
			if err = json.Unmarshal(data, &msg); err != nil {
				log.Error().Err(err).Msg("failed to unmarshal msg")
				continue
			}
			msg.UserID = userID

			if err = storeMessage(ctx, log, cache, kafkaWriter, msg); err != nil {
				log.Error().Err(err).Send()
				continue
			}
			msg.Print()

			go func() {
				select {
				case <-ctx.Done():
					return
				case broadcast <- msg:
					return
				}
			}()
		}
	}
}

func registerUser(ctx context.Context, con *websocket.Conn, addr config.StorageAddr) (int, error) {
	mt, username, err := con.ReadMessage()
	if err != nil || mt == websocket.CloseMessage {
		return 0, err
	}

	length := utf8.RuneCountInString(string(username))
	if length == 0 || length > 50 {
		return 0, errors.New("username length is not supported")
	}

	return sendRegisterRequest(ctx, addr, string(username))
}

func sendRegisterRequest(ctx context.Context, addr config.StorageAddr, username string) (int, error) {
	requestBody, err := json.Marshal(response.RegisterReq{Username: username})
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("http://%s:%s/register", addr.Host, addr.Port), bytes.NewBuffer(requestBody))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var userID response.RegisterResp
	if err = render.DecodeJSON(resp.Body, &userID); err != nil {
		return 0, err
	}

	return userID.UserID, nil
}

func outputRecent(ctx context.Context, log zerolog.Logger, repo resources.CacheRepo, con *websocket.Conn, cm resources.ClientManager) {
	recentMessages, err := repo.GetLastTen(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get messages from db")
		return
	}

	for _, msg := range recentMessages {
		if err := cm.WriteMsg(con, msg); err != nil {
			log.Error().Err(err).Msg("error on writing")
		}
	}
}

func storeMessage(ctx context.Context, log zerolog.Logger, cache resources.CacheRepo, kafkaWriter resources.MessageBroker, msg response.Msg) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	start := time.Now()
	kafkaWriter.Write(ctx, data)
	log.Info().Dur("kafka wrtie time", time.Since(start)).Send()

	start = time.Now()
	if err = cache.AddMessage(ctx, data); err != nil {
		return err
	}
	log.Info().Dur("redis wrtie time", time.Since(start)).Send()

	return nil
}
