package messages

import (
	"sync/atomic"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/process"
	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
)

type processorImpl struct {
	queue     chan persistence.Message
	callbacks Callbacks

	running atomic.Bool
	quit    chan struct{}
	done    chan struct{}
}

func NewProcessor(
	messageQueueSize int, callbacks Callbacks,
) Processor {
	return &processorImpl{
		queue:     make(chan persistence.Message, messageQueueSize),
		callbacks: callbacks,

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

func (p *processorImpl) Enqueue(msg persistence.Message) {
	p.queue <- msg
}

func (p *processorImpl) activeLoop() error {
	running := true

	var err error

	defer func() {
		p.done <- struct{}{}
	}()

	if p.callbacks.Start != nil {
		err = process.SafeRunSync(process.RunFunc(p.callbacks.Start))
		if err != nil {
			return err
		}
	}

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

	if p.callbacks.Finish != nil {
		finishErr := process.SafeRunSync(process.RunFunc(p.callbacks.Finish))
		if finishErr != nil {
			return finishErr
		}
	}

	return err
}

func (p *processorImpl) processMessage(msg persistence.Message) error {
	return process.SafeRunSync(
		func() error {
			// Note: this is technically unsafe as we don't verify that the
			// callback is set, unlike the other ones
			return p.callbacks.Message(msg)
		},
	)
}
