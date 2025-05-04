package messages

import (
	"sync/atomic"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/process"
	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
)

type Processor interface {
	Start() error
	Stop() error

	Enqueue(msg persistence.Message) error
}

type MessageCallback func(msg persistence.Message) error
type FinishCallback func() error

type processorImpl struct {
	queue          chan persistence.Message
	msgCallback    MessageCallback
	finishCallback FinishCallback

	running atomic.Bool
	quit    chan struct{}
	done    chan struct{}
}

func newProcessor(
	messageQueueSize int,
	msgCallback MessageCallback,
	finishCallback FinishCallback,
) Processor {
	return &processorImpl{
		queue:          make(chan persistence.Message, messageQueueSize),
		msgCallback:    msgCallback,
		finishCallback: finishCallback,

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

	if p.finishCallback != nil {
		finishErr := process.SafeRunSync(process.RunFunc(p.finishCallback))
		if finishErr != nil {
			return finishErr
		}
	}

	return err
}

func (p *processorImpl) processMessage(msg persistence.Message) error {
	return process.SafeRunSync(
		func() error {
			return p.msgCallback(msg)
		},
	)
}
