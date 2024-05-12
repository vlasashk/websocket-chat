package main

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/vlasashk/websocket-chat/config"
	"github.com/vlasashk/websocket-chat/internal/server"
)

func main() {
	ctx := context.Background()

	cfg, err := config.NewServerCfg()
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	if err = server.Run(ctx, cfg); err != nil {
		log.Fatal().Err(err).Send()
	}
}
