package internal

import (
	"context"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/Knoblauchpilze/chat-server/internal/service"
)

func RunTcpServer(
	ctx context.Context,
	conf Configuration,
	services service.Services,
	log logger.Logger,
) error {
	chat := services.Chat
	conf.Callbacks = chat.GenerateCallbacks()

	chat.Start()
	defer chat.Stop()

	return tcpListenAndServe(ctx, conf, log)
}
