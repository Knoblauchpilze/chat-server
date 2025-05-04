package messages

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestIT_Processor_ExpectStartCallbackCalled(t *testing.T) {
	var called int
	startCb := func() error {
		called++
		return nil
	}

	processor := newTestProcessorWithCallbacks(startCb, dummyMessageCallback, nil)

	wg := asyncStartProcessorAndAssertNoError(t, processor)

	err := processor.Stop()
	assert.Nil(t, err, "Actual err: %v", err)
	wg.Wait()

	assert.Equal(t, 1, called)
}

func TestIT_Processor_WhenStartCallbackFails_ExpectErrorIsReturned(t *testing.T) {
	testErr := fmt.Errorf("some error")
	startCb := func() error {
		return testErr
	}
	processor := newTestProcessorWithCallbacks(startCb, dummyMessageCallback, nil)

	wg := asyncStartProcessorAndAssertError(t, processor, testErr)

	msg := persistence.Message{}
	processor.Enqueue(msg)

	err := processor.Stop()
	assert.Nil(t, err, "Actual err: %v", err)
	wg.Wait()
}

func TestIT_Processor_EnqueueMessage_ExpectMessageCallbackCalled(t *testing.T) {
	var receivedMsg persistence.Message
	var called int
	msgCb := func(msg persistence.Message) error {
		called++
		receivedMsg = msg
		return nil
	}

	processor := newTestProcessorWithCallbacks(nil, msgCb, nil)

	wg := asyncStartProcessorAndAssertNoError(t, processor)

	roomId := uuid.New()
	msg := persistence.Message{
		Id:       uuid.New(),
		ChatUser: uuid.New(),
		Room:     roomId,
		Message:  fmt.Sprintf("hello %s", roomId),
	}
	processor.Enqueue(msg)

	err := processor.Stop()
	assert.Nil(t, err, "Actual err: %v", err)
	wg.Wait()

	assert.Equal(t, 1, called)
	assert.Equal(t, msg, receivedMsg)
}

func TestIT_Processor_WhenMessageQueueIsFull_ExpectCallBlocks(t *testing.T) {
	var block atomic.Bool
	block.Store(true)
	unblock := make(chan struct{}, 1)

	blockingMsgCb := func(msg persistence.Message) error {
		if block.Load() {
			<-unblock
		}

		return nil
	}
	processor := newTestProcessorWithCallbacks(nil, blockingMsgCb, nil)

	wg := asyncStartProcessorAndAssertNoError(t, processor)

	enqueueMessage := func() {
		msg := persistence.Message{}
		processor.Enqueue(msg)
	}

	// We have a queue of 1: the first message will be enqueued and stuck in
	// the repository. The second one will be enqueued and stay in the queue.
	enqueueMessage()
	enqueueMessage()

	// The third message will not be able to be enqueued
	msgEnqueued := make(chan struct{}, 1)
	go func() {
		defer func() {
			msgEnqueued <- struct{}{}

		}()

		enqueueMessage()
	}()

	timeout := time.After(100 * time.Millisecond)
	select {
	case <-timeout:
	case <-msgEnqueued:
		assert.Fail(t, "Message should not have been enqueued")
	}

	// We need to unblock the message repository
	block.Store(false)
	unblock <- struct{}{}

	err := processor.Stop()
	assert.Nil(t, err, "Actual err: %v", err)
	wg.Wait()
}

func TestIT_Processor_WhenMessageFailsToBeProcessed_ExpectProcessingStops(t *testing.T) {
	testErr := fmt.Errorf("some error")
	msgCb := func(msg persistence.Message) error {
		return testErr
	}
	processor := newTestProcessorWithCallbacks(nil, msgCb, nil)

	wg := asyncStartProcessorAndAssertError(t, processor, testErr)

	msg := persistence.Message{}
	processor.Enqueue(msg)

	err := processor.Stop()
	assert.Nil(t, err, "Actual err: %v", err)
	wg.Wait()
}

func TestIT_Processor_WhenStopped_ExpectFinishCallbackCalled(t *testing.T) {
	var called int
	finishCb := func() error {
		called++
		return nil
	}

	processor := newTestProcessorWithCallbacks(nil, dummyMessageCallback, finishCb)

	wg := asyncStartProcessorAndAssertNoError(t, processor)

	roomId := uuid.New()
	msg := persistence.Message{
		Id:       uuid.New(),
		ChatUser: uuid.New(),
		Room:     roomId,
		Message:  fmt.Sprintf("hello %s", roomId),
	}
	processor.Enqueue(msg)

	err := processor.Stop()
	assert.Nil(t, err, "Actual err: %v", err)
	wg.Wait()

	assert.Equal(t, 1, called)
}

func TestIT_Processor_WhenFinishCallbackFails_ExpectErrorIsReturned(t *testing.T) {
	testErr := fmt.Errorf("some error")
	finishCb := func() error {
		return testErr
	}
	processor := newTestProcessorWithCallbacks(nil, dummyMessageCallback, finishCb)

	wg := asyncStartProcessorAndAssertError(t, processor, testErr)

	msg := persistence.Message{}
	processor.Enqueue(msg)

	err := processor.Stop()
	assert.Nil(t, err, "Actual err: %v", err)
	wg.Wait()
}

func newTestProcessorWithCallbacks(
	startCallback StartCallback,
	msgCallback MessageCallback,
	finishCallback FinishCallback,
) Processor {
	cb := Callbacks{
		Start:   startCallback,
		Message: msgCallback,
		Finish:  finishCallback,
	}
	return NewProcessor(1, cb)
}
