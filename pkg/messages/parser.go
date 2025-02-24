package messages

import "github.com/google/uuid"

type Parser interface {
	OnReadData(id uuid.UUID, data []byte) bool
}

type parserImpl struct {
	queue Queue
}

func NewParser(queue Queue) Parser {
	return &parserImpl{
		queue: queue,
	}
}

func (p *parserImpl) OnReadData(id uuid.UUID, data []byte) bool {
	msg := NewMessage(DATA_RECEIVED)
	p.queue <- msg
	return true
}
