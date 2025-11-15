package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"

	"github.com/ZenoN-Cloud/zeno-auth/internal/app"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		log.Info().Msg("Shutting down gracefully...")
		cancel()
	}()

	application, err := app.New()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create app")
	}
	defer application.Close()

	if err := application.Run(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to run app")
	}
}
