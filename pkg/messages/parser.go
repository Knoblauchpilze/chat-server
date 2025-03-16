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
	queue OutgoingQueue
}

func NewParser(queue OutgoingQueue, log logger.Logger) Parser {
	return &parserImpl{
		log:   log,
		queue: queue,
	}
}

func (p *parserImpl) OnReadData(id uuid.UUID, data []byte) bool {
	msg, err := Decode(data)
	if err != nil {
		p.log.Warnf("Unable to decode %d byte(s) received from %v: %v", len(data), id, err)
		// Still return true as it can be that the data is just incomplete.
		return true
	}

	p.queue <- msg
	return true
}
