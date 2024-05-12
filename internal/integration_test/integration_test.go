package integration_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vlasashk/websocket-chat/config"
	"github.com/vlasashk/websocket-chat/internal/server"
	"github.com/vlasashk/websocket-chat/internal/storage"
	"github.com/vlasashk/websocket-chat/internal/storage/adapters/pgrepo"
	"github.com/vlasashk/websocket-chat/migrations"
	"github.com/vlasashk/websocket-chat/pkg/response"
)

const (
	httpServ      = "localhost:8080"
	httpStorage   = "localhost:8000"
	chatPath      = "/chat"
	healthzPath   = "/healthz"
	dbPass        = "postgres"
	migrationPath = "../../migrations"
)

var testPool *pgxpool.Pool

func TestMain(t *testing.M) {
	os.Setenv("DB_PASSWORD", dbPass)
	os.Setenv("DB_MIGRATION_PATH", migrationPath)

	go func() {
		cfg, err := config.NewStorageCfg()
		if err != nil {
			log.Fatal().Err(err).Msg("could not load server config")
		}
		log.Error().Err(storage.Run(context.Background(), cfg)).Send()
	}()

	go func() {
		cfg, err := config.NewServerCfg()
		if err != nil {
			log.Fatal().Err(err).Msg("could not load server config")
		}
		log.Error().Err(server.Run(context.Background(), cfg)).Send()
	}()

	if err := healthcheck("http://" + httpStorage + healthzPath); err != nil {
		log.Fatal().Err(err).Send()
	}

	if err := healthcheck("http://" + httpServ + healthzPath); err != nil {
		log.Fatal().Err(err).Send()
	}

	tearDown, err := setupDB()
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	code := t.Run()

	if err = tearDown(); err != nil {
		log.Error().Err(err).Send()
	}

	os.Exit(code)
}

func TestServer(t *testing.T) {
	t.Run("NicknameRegister", func(t *testing.T) {
		urlDial := url.URL{Scheme: "ws", Host: httpServ, Path: chatPath}
		con, _, err := websocket.DefaultDialer.Dial(urlDial.String(), nil)
		require.NoError(t, err)
		defer func() {
			assert.NoError(t, con.Close())
		}()

		registerTest(t, con, "first_test", 1)

		closeMsg := websocket.FormatCloseMessage(websocket.CloseMessage, "close connection")

		err = con.WriteControl(websocket.CloseMessage, closeMsg, time.Now().Add(2*time.Second))
		assert.NoError(t, err)
		// wait server to close con
		time.Sleep(100 * time.Millisecond)
	})
	t.Run("SendFiveMessages", func(t *testing.T) {
		urlDial := url.URL{Scheme: "ws", Host: httpServ, Path: chatPath}
		con, _, err := websocket.DefaultDialer.Dial(urlDial.String(), nil)
		require.NoError(t, err)
		defer func() {
			assert.NoError(t, con.Close())
		}()

		registerTest(t, con, "second_test", 2)
		go testReceive(t, con, 5)

		err = sendMsg(con, 5)
		assert.NoError(t, err)

		closeMsg := websocket.FormatCloseMessage(websocket.CloseMessage, "close connection")
		// wait before sending close
		time.Sleep(200 * time.Millisecond)

		err = con.WriteControl(websocket.CloseMessage, closeMsg, time.Now().Add(2*time.Second))
		assert.NoError(t, err)
		// wait server to close con
		time.Sleep(100 * time.Millisecond)
	})
	t.Run("GetLastFive", func(t *testing.T) {
		urlDial := url.URL{Scheme: "ws", Host: httpServ, Path: chatPath}
		con, _, err := websocket.DefaultDialer.Dial(urlDial.String(), nil)
		require.NoError(t, err)
		defer func() {
			assert.NoError(t, con.Close())
		}()

		registerTest(t, con, "third_test", 3)

		testReceive(t, con, 5)

		closeMsg := websocket.FormatCloseMessage(websocket.CloseMessage, "close connection")
		// wait before sending close
		time.Sleep(200 * time.Millisecond)

		err = con.WriteControl(websocket.CloseMessage, closeMsg, time.Now().Add(2*time.Second))
		assert.NoError(t, err)
		// wait server to close con
		time.Sleep(100 * time.Millisecond)
	})
}

func sendMsg(con *websocket.Conn, amount int) error {
	var sent int
	for sent < amount {
		msg := response.Msg{
			Username: "second_test",
			Text:     fmt.Sprintf("test_%d", sent),
		}
		if err := con.WriteJSON(msg); err != nil {
			return err
		}
		sent++
		time.Sleep(1 * time.Second)
	}
	return nil
}

func testReceive(t *testing.T, con *websocket.Conn, amount int) {
	t.Helper()
	var sent int
	for sent < amount {
		_, message, err := con.ReadMessage()
		assert.NoError(t, err)

		msg := strings.TrimSuffix(string(message), "\n")
		assert.Equal(t, fmt.Sprintf("{\"user_id\":2,\"username\":\"second_test\",\"text\":\"test_%d\"}", sent), msg)
		sent++
	}
}

func registerTest(t *testing.T, con *websocket.Conn, username string, expect int) {
	t.Helper()
	err := con.WriteMessage(websocket.TextMessage, []byte(username))
	require.NoError(t, err)
	// wait server to process msg
	time.Sleep(1000 * time.Millisecond)
	var count int

	err = testPool.QueryRow(context.Background(), `SELECT count(*) FROM users`).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, expect, count)
}

func healthcheck(serviceURL string) error {
	retries := 10
	for retries > 0 {
		time.Sleep(1 * time.Second)
		retries--

		resp, err := http.Get(serviceURL)
		if err != nil {
			log.Error().Err(err).Int("attempts left", retries).Msg("server is not healthy")
			continue
		}

		if resp.StatusCode != http.StatusOK {
			log.Error().Err(err).Int("attempts left", retries).Msg("server is not healthy")
			continue
		}
		return nil
	}
	return errors.New("server is not healthy")
}

func setupDB() (func() error, error) {
	cfg, err := config.NewStorageCfg()
	if err != nil {
		return nil, err
	}

	repo, err := pgrepo.New(context.Background(), cfg.Repo, zerolog.New(os.Stdout))
	if err != nil {
		return nil, err
	}

	testPool = repo.Pool

	if err = migrations.Up(testPool, migrationPath); err != nil {
		return nil, err
	}

	return func() error {
		return migrations.Down(testPool, migrationPath)
	}, nil
}
