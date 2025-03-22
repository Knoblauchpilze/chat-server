package internal

import (
	"context"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/Knoblauchpilze/chat-server/internal/service"
)

func RunTcpServer(ctx context.Context, conf Configuration, log logger.Logger) error {
	chat := service.NewChat(log)
	conf.Callbacks = chat.GenerateCallbacks()

	chat.Start()
	defer chat.Stop()

	return tcpListenAndServe(ctx, conf, log)
}
