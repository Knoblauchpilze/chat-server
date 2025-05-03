package messages

import (
	"context"
	"sync/atomic"

	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
	"github.com/Knoblauchpilze/chat-server/pkg/repositories"
)

type Processor interface {
	Start() error
	Stop() error

	Enqueue(msg persistence.Message) error
}

type processorImpl struct {
	queue       chan persistence.Message
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
		queue:       make(chan persistence.Message, messageQueueSize),
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

func (p *processorImpl) Enqueue(msg persistence.Message) error {
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

func (p *processorImpl) processMessage(msg persistence.Message) error {
	_, err := p.messageRepo.Create(context.Background(), msg)
	if err != nil {
		return err
	}

	out := NewRoomMessage(msg.ChatUser, msg.Room, msg.Message)
	p.dispatcher.BroadcastExcept(msg.ChatUser, out)

	return nil
}
