package service

import (
	"os"
	"testing"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/Knoblauchpilze/chat-server/pkg/clients"
	"github.com/Knoblauchpilze/chat-server/pkg/messages"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type clientManagerMock struct {
	clients.Manager

	broadcastMsgs       []messages.Message
	broadcastExceptMsgs []messages.Message
	broadcastExceptIds  []uuid.UUID
	directMsgs          []messages.Message
	directIds           []uuid.UUID
}

func TestUnit_MessageProcessingService_StartStop(t *testing.T) {
	var queue messages.IncomingQueue
	var manager clients.Manager
	service := NewMessageProcessingService(queue, manager, logger.New(os.Stdout))

	wg := asyncRunMessageProcessingService(t, service)

	service.Stop()
	wg.Wait()
}

func TestUnit_MessageProcessingService_WhenClientConnectedMessageReceived_ExpectBroadcastExcept(t *testing.T) {
	log := logger.New(os.Stdout)
	queue := make(chan messages.Message, 1)
	managerMock := &clientManagerMock{}
	service := NewMessageProcessingService(queue, managerMock, log)

	wg := asyncRunMessageProcessingService(t, service)

	clientId := uuid.New()
	msg := messages.NewClientConnectedMessage(clientId)
	queue <- msg

	service.Stop()
	wg.Wait()

	assert.Equal(t, 1, len(managerMock.broadcastExceptMsgs))
	actual, ok := managerMock.broadcastExceptMsgs[0].(messages.ClientConnectedMessage)
	assert.True(t, ok)
	assert.Equal(t, clientId, actual.Client)
	assert.Equal(t, 1, len(managerMock.broadcastExceptIds))
	assert.Equal(t, clientId, managerMock.broadcastExceptIds[0])
}

func TestUnit_MessageProcessingService_WhenClientDisonnectedMessageReceived_ExpectBroadcast(t *testing.T) {
	log := logger.New(os.Stdout)
	queue := make(chan messages.Message, 1)
	managerMock := &clientManagerMock{}
	service := NewMessageProcessingService(queue, managerMock, log)

	wg := asyncRunMessageProcessingService(t, service)

	clientId := uuid.New()
	msg := messages.NewClientDisconnectedMessage(clientId)
	queue <- msg

	service.Stop()
	wg.Wait()

	assert.Equal(t, 1, len(managerMock.broadcastMsgs))
	actual, ok := managerMock.broadcastMsgs[0].(messages.ClientDisconnectedMessage)
	assert.True(t, ok)
	assert.Equal(t, clientId, actual.Client)
}

func TestUnit_MessageProcessingService_WhenDirectMessageReceived_ExpectBroadcast(t *testing.T) {
	log := logger.New(os.Stdout)
	queue := make(chan messages.Message, 1)
	managerMock := &clientManagerMock{}
	service := NewMessageProcessingService(queue, managerMock, log)

	wg := asyncRunMessageProcessingService(t, service)

	client1Id := uuid.New()
	client2Id := uuid.New()
	msg := messages.NewDirectMessage(client1Id, client2Id, "Hello")
	queue <- msg

	service.Stop()
	wg.Wait()

	assert.Equal(t, 1, len(managerMock.directMsgs))
	actual, ok := managerMock.directMsgs[0].(messages.DirectMessage)
	assert.True(t, ok)
	assert.Equal(t, client1Id, actual.Emitter)
	assert.Equal(t, client2Id, actual.Receiver)
	assert.Equal(t, "Hello", actual.Content)
	assert.Equal(t, 1, len(managerMock.directIds))
	assert.Equal(t, client2Id, managerMock.directIds[0])
}

func (m *clientManagerMock) Broadcast(msg messages.Message) {
	m.broadcastMsgs = append(m.broadcastMsgs, msg)
}

func (m *clientManagerMock) BroadcastExcept(id uuid.UUID, msg messages.Message) {
	m.broadcastExceptMsgs = append(m.broadcastExceptMsgs, msg)
	m.broadcastExceptIds = append(m.broadcastExceptIds, id)
}

func (m *clientManagerMock) SendTo(id uuid.UUID, msg messages.Message) {
	m.directMsgs = append(m.directMsgs, msg)
	m.directIds = append(m.directIds, id)
}
