package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/config"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/Knoblauchpilze/chat-server/cmd/websocket/internal"
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

	tcpConf := tcp.AcceptorConfig{
		BasePath:        fmt.Sprintf("%s/ws", conf.Server.BasePath),
		Port:            conf.TcpPort,
		ShutdownTimeout: 1 * time.Second,
	}
	a := tcp.NewWebsocketAcceptor(tcpConf, log)

	notifyCtx, stop := signal.NotifyContext(
		context.Background(), os.Interrupt, syscall.SIGTERM,
	)
	defer stop()

	done := make(chan error, 1)
	go func() {
		var err error
		defer func() {
			done <- err
		}()

		err = a.Accept()
	}()

	var serveErr error
	select {
	case <-notifyCtx.Done():
		log.Infof("Received shutdown signal, shutting down...")
	case serveErr = <-done:
		log.Infof("Server has shut down")
	}

	if serveErr != nil {
		log.Infof("Serve error: %v", serveErr)
	}

	err = a.Close()
	if err != nil {
		log.Errorf("Failed to close acceptor: %v", err)
		os.Exit(1)
	}

	log.Infof("Gracefully shutdown")

	// opts := websocket.AcceptOptions{
	// 	OriginPatterns: []string{"localhost:*"},
	// }

	// http.HandleFunc("/chats/ws", func(w http.ResponseWriter, r *http.Request) {
	// 	printHost(r)

	// 	c, err := websocket.Accept(w, r, &opts)
	// 	if err != nil {
	// 		log.Errorf("failed to accept connection: %v", err)
	// 		return
	// 	}

	// 	defer c.CloseNow()

	// 	// Set the context as needed. Use of r.Context() is not recommended
	// 	// to avoid surprising behavior (see http.Hijacker).
	// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	// 	defer cancel()
	// 	msg, data, err := c.Read(ctx)
	// 	if err != nil {
	// 		log.Errorf("failed to read from connection: %v", err)
	// 		c.Close(websocket.StatusInternalError, "failed to read")
	// 		return
	// 	}

	// 	id, err := uuid.FromBytes(data)
	// 	if err != nil {
	// 		log.Errorf("failed to interpret handshake as uuid: %v", err)
	// 		c.Close(websocket.StatusUnsupportedData, "wrong handshake")
	// 		return
	// 	}

	// 	log.Infof("received msg %v (%d byte(s)): \"%s\", id: %s", msg, len(data), string(data), id)

	// 	c.Close(websocket.StatusNormalClosure, "all good")
	// })

	// address := fmt.Sprintf(":%d", conf.TcpPort)
	// err = http.ListenAndServe(address, nil)
	// if err != nil {
	// 	log.Errorf("failed to listen and serve: %v", err)
	// 	os.Exit(1)
	// }
}

func printHost(req *http.Request) {
	origin := req.Header.Get("Origin")
	if origin == "" {
		fmt.Printf("nil origin\n")
		return
	}

	u, err := url.Parse(origin)
	if err != nil {
		fmt.Printf("failed to parse origin: %v\n", err)
		return
	}

	if strings.EqualFold(req.Host, u.Host) {
		fmt.Printf("url host is equal to request host: %s\n", u.Host)
		return
	}

	fmt.Printf("url host: %v\n", u.Host)
}
