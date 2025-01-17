package tcp

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/KnoblauchPilze/backend-toolkit/pkg/errors"
	"github.com/KnoblauchPilze/backend-toolkit/pkg/logger"
	"github.com/KnoblauchPilze/backend-toolkit/pkg/server"
)

type Server interface {
	Start(ctx context.Context) error
}

type serverImpl struct {
	basePath        string
	port            uint16
	shutdownTimeout time.Duration
	log             logger.Logger
	listener        net.Listener

	lock     sync.Mutex
	handlers []ConnectionHandler
}

func NewServer(config server.Config, log logger.Logger) Server {
	return &serverImpl{
		basePath:        config.BasePath,
		port:            config.Port,
		shutdownTimeout: 5 * time.Second,
		log:             log,
	}
}

func (s *serverImpl) Start(ctx context.Context) error {
	address := fmt.Sprintf(":%d", s.port)

	var err error
	s.listener, err = net.Listen("tcp", address)
	if err != nil {
		return errors.WrapCode(err, TcpInitializationFailure)
	}

	return s.acceptLoop()
}

func (s *serverImpl) acceptLoop() error {
	s.log.Infof("Starting server on port %v", s.port)

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			s.log.Errorf("Failed to accept connection: %v", err)
			continue
		}

		go s.handleConnection(conn)
	}
}

func (s *serverImpl) handleConnection(conn net.Conn) {
	s.lock.Lock()
	defer s.lock.Unlock()

	h := newHandler(s.log)

	s.handlers = append(s.handlers, h)

	if err := h.Handle(conn); err != nil {
		s.log.Errorf("Failure while handling connection: %v", err)
	}
}
