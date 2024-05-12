package main

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/vlasashk/websocket-chat/config"
	"github.com/vlasashk/websocket-chat/internal/client"
)

func main() {
	ctx := context.Background()

	cfg, err := config.NewClientCfg()
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	if err = client.Run(ctx, cfg); err != nil {
		log.Fatal().Err(err).Send()
	}
}
