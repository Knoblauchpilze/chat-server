package messages

import (
	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/google/uuid"
)

type Parser interface {
	OnReadData(id uuid.UUID, data []byte) bool
}

type parserImpl struct {
	log   logger.Logger
	queue Queue
}

func NewParser(queue Queue, log logger.Logger) Parser {
	return &parserImpl{
		log:   log,
		queue: queue,
	}
}

func (p *parserImpl) OnReadData(id uuid.UUID, data []byte) bool {
	msg, err := Decode(data)
	if err != nil {
		p.log.Warnf("Unable to decode %d byte(s) received from %v: %v", len(data), id, err)
		return false
	}

	p.queue <- msg
	return true
}
