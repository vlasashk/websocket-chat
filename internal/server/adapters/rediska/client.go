package rediska

import (
	"context"
	"encoding/json"
	"net"

	"github.com/redis/go-redis/v9"
	"github.com/vlasashk/websocket-chat/config"
	"github.com/vlasashk/websocket-chat/pkg/response"
	"github.com/vlasashk/websocket-chat/pkg/utils"
)

type Rediska struct {
	Client     *redis.Client
	MaxRecords int64
	HeadSize   int64
}

func NewClient(cfg config.RedisAddr) (*Rediska, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     net.JoinHostPort(cfg.Host, cfg.Port),
		Username: "",
		Password: "",
	})

	return &Rediska{
		Client:     client,
		MaxRecords: cfg.MaxRecords,
		HeadSize:   cfg.HeadSize,
	}, nil
}

func (r Rediska) AddMessage(ctx context.Context, data []byte) error {
	pipe := r.Client.Pipeline()

	pipe.LPush(ctx, "chat", data)
	pipe.LTrim(ctx, "chat", 0, r.MaxRecords)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r Rediska) GetLastTen(ctx context.Context) ([]response.Msg, error) {
	res := make([]response.Msg, 0, 10)

	data, err := r.Client.LRange(ctx, "chat", 0, r.HeadSize-1).Result()
	if err != nil {
		return nil, err
	}
	for _, v := range data {
		var msg response.Msg
		err = json.Unmarshal([]byte(v), &msg)
		if err != nil {
			return nil, err
		}
		res = append(res, msg)
	}

	utils.FlipMessageOrder(res)

	return res, nil
}
