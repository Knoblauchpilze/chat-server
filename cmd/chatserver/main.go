package main

import (
	"fmt"
	"os"

	"github.com/KnoblauchPilze/backend-toolkit/pkg/config"
	"github.com/KnoblauchPilze/backend-toolkit/pkg/logger"
	"github.com/KnoblauchPilze/backend-toolkit/pkg/rest"
	"github.com/Knoblauchpilze/chat-server/cmd/chatserver/internal"
	"github.com/Knoblauchpilze/chat-server/pkg/tcp"
	"github.com/labstack/echo/v4"
)

func determineConfigName() string {
	if len(os.Args) < 2 {
		return "users-prod.yml"
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

	// https://github.com/labstack/echo?tab=readme-ov-file#example
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	route := rest.ConcatenateEndpoints("/", conf.Server.BasePath)
	e.GET(route, tcp.NewHandler(log))

	log.Infof("Starting server at \"%v\" on port %v", route, conf.Server.Port)

	address := fmt.Sprintf(":%d", conf.Server.Port)
	err = e.Start(address)
	if err != nil {
		log.Errorf("Failure while serving TCP: %v", err)
		os.Exit(1)
	}
}
