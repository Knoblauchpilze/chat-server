package messages

import (
	"context"
	"sync/atomic"

	"github.com/Knoblauchpilze/chat-server/pkg/communication"
	"github.com/Knoblauchpilze/chat-server/pkg/repositories"
)

type Processor interface {
	Start() error
	Stop() error

	Enqueue(msg communication.MessageDtoRequest) error
}

type processorImpl struct {
	queue       chan communication.MessageDtoRequest
	messageRepo repositories.MessageRepository
	dispatcher  Dispatcher

	running atomic.Bool
	quit    chan struct{}
	done    chan struct{}
}

func NewProcessor(
	messageQueueSize int,
	dispatcher Dispatcher,
	repos repositories.Repositories,
) Processor {
	return &processorImpl{
		queue:       make(chan communication.MessageDtoRequest, messageQueueSize),
		messageRepo: repos.Message,
		dispatcher:  dispatcher,

		quit: make(chan struct{}, 1),
		done: make(chan struct{}, 1),
	}
}

func (p *processorImpl) Start() error {
	if !p.running.CompareAndSwap(false, true) {
		return nil
	}

	return p.activeLoop()
}

func (p *processorImpl) Stop() error {
	if !p.running.CompareAndSwap(true, false) {
		return nil
	}

	p.quit <- struct{}{}
	<-p.done

	return nil
}

func (p *processorImpl) Enqueue(msg communication.MessageDtoRequest) error {
	p.queue <- msg
	return nil
}

func (p *processorImpl) activeLoop() error {
	running := true

	var err error

	defer func() {
		p.done <- struct{}{}
	}()

	for running {
		select {
		case <-p.quit:
			running = false
		case msg := <-p.queue:
			err = p.processMessage(msg)
		}

		if err != nil {
			running = false
		}
	}

	return err
}

func (p *processorImpl) processMessage(msg communication.MessageDtoRequest) error {
	dbMsg := communication.FromMessageDtoRequest(msg)

	_, err := p.messageRepo.Create(context.Background(), dbMsg)
	if err != nil {
		return err
	}

	out := NewRoomMessage(msg.User, msg.Room, msg.Message)
	p.dispatcher.BroadcastExcept(msg.User, out)

	return nil
}
