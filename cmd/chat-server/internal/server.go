package internal

import (
	"context"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/Knoblauchpilze/chat-server/pkg/service"
)

func RunServer(ctx context.Context, conf Configuration, log logger.Logger) error {
	chat := service.NewChatService(log)
	conf.Callbacks = chat.GenerateCallbacks()

	chat.Start()
	defer chat.Stop()

	return listenAndServe(ctx, conf, log)
}
