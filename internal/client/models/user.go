package models

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/vlasashk/websocket-chat/config"
	"github.com/vlasashk/websocket-chat/pkg/listener"
	"github.com/vlasashk/websocket-chat/pkg/response"
)

type User struct {
	Username string
	Reader   *bufio.Reader
	Con      *websocket.Conn
}

func NewUser(cfg config.ClientCfg) (*User, error) {
	reader := bufio.NewReader(os.Stdin)
	urlDial := url.URL{Scheme: cfg.Scheme, Host: net.JoinHostPort(cfg.Host, cfg.Port), Path: cfg.Path}

	username, err := setUsername(reader)
	if err != nil {
		return nil, err
	}

	con, _, err := websocket.DefaultDialer.Dial(urlDial.String(), nil)
	if err != nil {
		return nil, err
	}

	if err = con.WriteMessage(websocket.TextMessage, []byte(username)); err != nil {
		if err := con.Close(); err != nil {
			log.Error().Err(err).Send()
		}
		return nil, err
	}

	return &User{
		Reader:   reader,
		Username: username,
		Con:      con,
	}, nil
}

func (u *User) Receiver(ctx context.Context, log zerolog.Logger) error {
	listen := listener.SocketListen(ctx, log, u.Con)
	for {
		select {
		case <-ctx.Done():
			return nil
		case data, ok := <-listen:
			if !ok {
				return errors.New("connection died unexpectedly")
			}
			var msg response.Msg
			if err := json.Unmarshal(data, &msg); err != nil {
				log.Error().Err(err).Msg("failed to unmarshal msg")
				continue
			}
			msg.Print()
		}
	}
}

func (u *User) Sender(ctx context.Context, log zerolog.Logger) error {
	input := u.typer(log)
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-input:
			if !ok {
				return errors.New("console reader is dead")
			}
			if err := u.Con.WriteJSON(msg); err != nil {
				log.Error().Err(err).Send()
				return err
			}
		}
	}
}

func (u *User) Close(log zerolog.Logger) {
	if err := u.Con.Close(); err != nil {
		log.Error().Err(err).Send()
	}
}

// typer runs separately exposing chanel to be able to gracefully shut down, since reading from console is blocking
func (u *User) typer(log zerolog.Logger) <-chan response.Msg {
	messages := make(chan response.Msg)
	go func() {
		var err error
		for {
			var msg response.Msg
			msg.Username = u.Username

			msg.Text, err = u.Reader.ReadString('\n')
			if err != nil {
				log.Error().Msg(fmt.Sprintln("read:", err))
				close(messages)
				return
			}

			msg.Text = strings.TrimSuffix(msg.Text, "\n")
			if utf8.RuneCountInString(msg.Text) == 0 {
				continue
			}

			messages <- msg
		}
	}()
	return messages
}

func setUsername(reader *bufio.Reader) (string, error) {
	var username string
	var err error
	for {
		fmt.Print("Enter your username: ")

		username, err = reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		username = strings.TrimSuffix(username, "\n")
		username = strings.ReplaceAll(username, " ", "_")

		length := utf8.RuneCountInString(username)
		if length == 0 {
			fmt.Println("ERROR: username is not specified. Try again")
			continue
		}

		if length > 50 {
			fmt.Println("ERROR: username is larger than 50 characters. Try again")
			continue
		}
		break
	}
	return username, nil
}
