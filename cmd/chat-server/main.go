package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/config"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/Knoblauchpilze/chat-server/cmd/chat-server/internal"
)

func determineConfigName() string {
	if len(os.Args) < 2 {
		return "config-prod.yml"
	}

	return os.Args[1]
}

func main() {
	log := logger.New(os.Stdout)

	conf, err := config.Load(determineConfigName(), internal.DefaultConfig())
	if err != nil {
		log.Error("Failed to load configuration", slog.Any("error", err))
		os.Exit(1)
	}

	fmt.Printf("log: %+v\n", conf)

	if err := internal.RunServer(context.Background(), conf, log); err != nil {
		log.Error("Error while serving", slog.Any("error", err))
		os.Exit(1)
	}
}
