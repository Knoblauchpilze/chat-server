package messages

import (
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUnit_Parser_WhenReadFails_ExpectNothingPublishedToTheQueue(t *testing.T) {
	var queue OutgoingQueue
	parser := NewParser(queue, logger.New(os.Stdout))

	// Should hang in case the queue is used as the queue is unbuffered.
	parser.OnReadData(uuid.New(), []byte("not-a-message"))
}

func TestUnit_Parser_WhenReadFails_ExpectRequestToStayAlive(t *testing.T) {
	var queue OutgoingQueue
	parser := NewParser(queue, logger.New(os.Stdout))

	_, actual := parser.OnReadData(uuid.New(), []byte("not-a-message"))

	assert.True(t, actual)
}

func TestUnit_Parser_WhenReadFails_ExpectNoBytesProcessed(t *testing.T) {
	var queue OutgoingQueue
	parser := NewParser(queue, logger.New(os.Stdout))

	actual, _ := parser.OnReadData(uuid.New(), []byte("not-a-message"))

	assert.Equal(t, 0, actual)
}

func TestUnit_Parser_WhenReadSucceeds_ExpectMessagePushedToTheQueue(t *testing.T) {
	encoded := []byte{
		// CLIENT_CONNECTED
		0x0, 0x0, 0x0, 0x0,
		// UUID
		0x2d, 0xbf, 0x26, 0x22, 0x2a, 0x95, 0x4b, 0xd1, 0x9b, 0x38, 0x2f, 0x7b, 0x4c, 0xe6, 0x5f, 0xfe,
	}

	queue := make(chan Message, 1)
	parser := NewParser(queue, logger.New(os.Stdout))

	parser.OnReadData(uuid.New(), encoded)

	msg := <-queue
	assert.Equal(t, CLIENT_CONNECTED, msg.Type())
}

func TestUnit_Parser_WhenReadSucceeds_ExpectConnectionToStayOpen(t *testing.T) {
	encoded := []byte{
		// CLIENT_CONNECTED
		0x0, 0x0, 0x0, 0x0,
		// UUID
		0x2d, 0xbf, 0x26, 0x22, 0x2a, 0x95, 0x4b, 0xd1, 0x9b, 0x38, 0x2f, 0x7b, 0x4c, 0xe6, 0x5f, 0xfe,
	}

	queue := make(OutgoingQueue, 1)
	parser := NewParser(queue, logger.New(os.Stdout))

	_, actual := parser.OnReadData(uuid.New(), encoded)

	assert.True(t, actual)
}

func TestUnit_Parser_WhenReadSucceeds_ExpectSomeBytesToBeProcessed(t *testing.T) {
	encoded := []byte{
		// CLIENT_CONNECTED
		0x0, 0x0, 0x0, 0x0,
		// UUID
		0x2d, 0xbf, 0x26, 0x22, 0x2a, 0x95, 0x4b, 0xd1, 0x9b, 0x38, 0x2f, 0x7b, 0x4c, 0xe6, 0x5f, 0xfe,
	}

	queue := make(OutgoingQueue, 1)
	parser := NewParser(queue, logger.New(os.Stdout))

	actual, _ := parser.OnReadData(uuid.New(), encoded)

	assert.Equal(t, len(encoded), actual)
}

func TestUnit_Parser_WhenMoreDataThanOneMessage_ExpectNotAllBytesToBeProcessed(t *testing.T) {
	encoded := []byte{
		// CLIENT_CONNECTED
		0x0, 0x0, 0x0, 0x0,
		// UUID
		0x2d, 0xbf, 0x26, 0x22, 0x2a, 0x95, 0x4b, 0xd1, 0x9b, 0x38, 0x2f, 0x7b, 0x4c, 0xe6, 0x5f, 0xfe,
		// Additional garbage data
		0x45, 0x23, 0x12, 0x78,
	}

	queue := make(OutgoingQueue, 1)
	parser := NewParser(queue, logger.New(os.Stdout))

	actual, _ := parser.OnReadData(uuid.New(), encoded)

	// 4 additional bytes of garbage
	expectedProcessed := len(encoded) - 4
	assert.Equal(t, expectedProcessed, actual)
}

func TestUnit_Parser_WhenReadSucceeds_ExpectMessageCorrectlyDecoded(t *testing.T) {
	encoded := []byte{
		// CLIENT_CONNECTED
		0x0, 0x0, 0x0, 0x0,
		// UUID
		0x2d, 0xbf, 0x26, 0x22, 0x2a, 0x95, 0x4b, 0xd1, 0x9b, 0x38, 0x2f, 0x7b, 0x4c, 0xe6, 0x5f, 0xfe,
	}

	queue := make(chan Message, 1)
	parser := NewParser(queue, logger.New(os.Stdout))

	processed, keepAlive := parser.OnReadData(uuid.New(), encoded)
	assert.Equal(t, len(encoded), processed)
	assert.True(t, keepAlive)

	msg := <-queue

	assert.Equal(t, CLIENT_CONNECTED, msg.Type())
	inMsg, ok := msg.(ClientConnectedMessage)
	assert.True(t, ok)
	assert.Equal(t, sampleUuid, inMsg.Client)
}

func TestUnit_Parser_WhenQueueIsFull_ExpectReadBlocks(t *testing.T) {
	encoded := []byte{
		// CLIENT_CONNECTED
		0x0, 0x0, 0x0, 0x0,
		// UUID
		0x2d, 0xbf, 0x26, 0x22, 0x2a, 0x95, 0x4b, 0xd1, 0x9b, 0x38, 0x2f, 0x7b, 0x4c, 0xe6, 0x5f, 0xfe,
	}

	queue := make(chan Message)
	parser := NewParser(queue, logger.New(os.Stdout))

	var wg sync.WaitGroup
	wg.Add(1)

	var done atomic.Bool

	go func() {
		defer wg.Done()
		defer done.Store(true)

		processed, keepAlive := parser.OnReadData(uuid.New(), encoded)
		assert.Equal(t, len(encoded), processed)
		assert.True(t, keepAlive)
	}()

	time.Sleep(300 * time.Millisecond)
	assert.False(t, done.Load())

	<-queue
	wg.Wait()

	assert.True(t, done.Load())
}
