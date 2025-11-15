package main

import (
	"github.com/ZenoN-Cloud/zeno-auth/internal/app"
	"github.com/rs/zerolog/log"
)

func main() {
	app, err := app.New()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create app")
	}

	if err := app.Run(); err != nil {
		log.Fatal().Err(err).Msg("Failed to run app")
	}
}