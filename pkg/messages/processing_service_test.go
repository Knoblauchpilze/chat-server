package messages

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const reasonableWaitTimeForServiceToBeUp = 50 * time.Millisecond

type dispatcherMock struct {
	MessageDispatcher

	broadcastMsgs       []Message
	broadcastExceptMsgs []Message
	broadcastExceptIds  []uuid.UUID
	directMsgs          []Message
	directIds           []uuid.UUID
}

func TestUnit_ProcessingService_StartStop(t *testing.T) {
	var queue IncomingQueue
	var manager MessageDispatcher
	service := NewProcessingService(queue, manager, logger.New(os.Stdout))

	wg := asyncRunProcessingService(t, service)

	service.Stop()
	wg.Wait()
}

func TestUnit_ProcessingService_WhenClientConnectedMessageReceived_ExpectBroadcastExcept(t *testing.T) {
	log := logger.New(os.Stdout)
	queue := make(chan Message, 1)
	dispatcherMock := &dispatcherMock{}
	service := NewProcessingService(queue, dispatcherMock, log)

	wg := asyncRunProcessingService(t, service)

	clientId := uuid.New()
	msg := NewClientConnectedMessage(clientId)
	queue <- msg

	service.Stop()
	wg.Wait()

	assert.Equal(t, 1, len(dispatcherMock.broadcastExceptMsgs))
	actual, ok := dispatcherMock.broadcastExceptMsgs[0].(ClientConnectedMessage)
	assert.True(t, ok)
	assert.Equal(t, clientId, actual.Client)
	assert.Equal(t, 1, len(dispatcherMock.broadcastExceptIds))
	assert.Equal(t, clientId, dispatcherMock.broadcastExceptIds[0])
}

func TestUnit_ProcessingService_WhenClientDisonnectedMessageReceived_ExpectBroadcast(t *testing.T) {
	log := logger.New(os.Stdout)
	queue := make(chan Message, 1)
	dispatcherMock := &dispatcherMock{}
	service := NewProcessingService(queue, dispatcherMock, log)

	wg := asyncRunProcessingService(t, service)

	clientId := uuid.New()
	msg := NewClientDisconnectedMessage(clientId)
	queue <- msg

	service.Stop()
	wg.Wait()

	assert.Equal(t, 1, len(dispatcherMock.broadcastMsgs))
	actual, ok := dispatcherMock.broadcastMsgs[0].(ClientDisconnectedMessage)
	assert.True(t, ok)
	assert.Equal(t, clientId, actual.Client)
}

func TestUnit_ProcessingService_WhenDirectMessageReceived_ExpectBroadcast(t *testing.T) {
	log := logger.New(os.Stdout)
	queue := make(chan Message, 1)
	dispatcherMock := &dispatcherMock{}
	service := NewProcessingService(queue, dispatcherMock, log)

	wg := asyncRunProcessingService(t, service)

	client1Id := uuid.New()
	client2Id := uuid.New()
	msg := NewDirectMessage(client1Id, client2Id, "Hello")
	queue <- msg

	service.Stop()
	wg.Wait()

	assert.Equal(t, 1, len(dispatcherMock.directMsgs))
	actual, ok := dispatcherMock.directMsgs[0].(DirectMessage)
	assert.True(t, ok)
	assert.Equal(t, client1Id, actual.Emitter)
	assert.Equal(t, client2Id, actual.Receiver)
	assert.Equal(t, "Hello", actual.Content)
	assert.Equal(t, 1, len(dispatcherMock.directIds))
	assert.Equal(t, client2Id, dispatcherMock.directIds[0])
}

func asyncRunProcessingService(
	t *testing.T, service ProcessingService,
) *sync.WaitGroup {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if panicErr := recover(); panicErr != nil {
				assert.Failf(t, "Server panicked", "Panic details: %v", panicErr)
			}
		}()
		service.Start()
	}()

	time.Sleep(reasonableWaitTimeForServiceToBeUp)

	return &wg
}

func (m *dispatcherMock) Broadcast(msg Message) {
	m.broadcastMsgs = append(m.broadcastMsgs, msg)
}

func (m *dispatcherMock) BroadcastExcept(id uuid.UUID, msg Message) {
	m.broadcastExceptMsgs = append(m.broadcastExceptMsgs, msg)
	m.broadcastExceptIds = append(m.broadcastExceptIds, id)
}

func (m *dispatcherMock) SendTo(id uuid.UUID, msg Message) {
	m.directMsgs = append(m.directMsgs, msg)
	m.directIds = append(m.directIds, id)
}
