package main

import (
	"context"
	"os"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/config"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/Knoblauchpilze/chat-server/cmd/chat-server/internal"
	"github.com/Knoblauchpilze/chat-server/pkg/service"
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

	chat := service.NewChatService()
	conf.Callbacks = chat.GenerateCallbacks()

	chat.Start()
	defer chat.Stop()

	err = internal.ListenAndServe(context.Background(), conf, log)
	if err != nil {
		log.Errorf("Error while serving TCP: %+v", err)
		os.Exit(1)
	}
}
