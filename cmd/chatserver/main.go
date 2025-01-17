package main

import (
	"context"
	"os"

	"github.com/KnoblauchPilze/backend-toolkit/pkg/config"
	"github.com/KnoblauchPilze/backend-toolkit/pkg/logger"
	"github.com/Knoblauchpilze/chat-server/cmd/chatserver/internal"
	"github.com/Knoblauchpilze/chat-server/pkg/tcp"
)

func determineConfigName() string {
	if len(os.Args) < 2 {
		return "config-prod.yml"
	}

	return os.Args[1]
}

func main() {
	log := logger.New(logger.NewPrettyWriter(os.Stdout))

	conf, err := config.Load(determineConfigName(), internal.DefaultConfig())
	if err != nil {
		log.Errorf("Failed to load configuration: %v", err)
		os.Exit(1)
	}

	s := tcp.NewServer(conf.Server, log)

	err = s.Start(context.Background())
	if err != nil {
		log.Errorf("Error while serving TCP: %v", err)
		os.Exit(1)
	}
}
