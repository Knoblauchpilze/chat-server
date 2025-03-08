package service

import (
	"os"
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/Knoblauchpilze/chat-server/pkg/clients"
	"github.com/Knoblauchpilze/chat-server/pkg/messages"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const reasonableWaitTimeForOnConnectMessageToBeProcessed = 100 * time.Millisecond

func TestUnit_ChatService_OnConnect_SendsMessagesToOthers(t *testing.T) {
	service, callbacks := newTestChatService()
	client1, server1 := newTestConnection(t, 6000)
	_, server2 := newTestConnection(t, 6000)
	wg := asyncRunChatService(t, service)

	client1Id := uuid.New()
	accepted := callbacks.OnConnect(client1Id, server1)
	assert.True(t, accepted)

	time.Sleep(reasonableWaitTimeForOnConnectMessageToBeProcessed)

	client2Id := uuid.New()
	accepted = callbacks.OnConnect(client2Id, server2)
	assert.True(t, accepted)

	data := readFromConnection(t, client1)
	msg, err := messages.Decode(data)
	assert.Nil(t, err, "Actual err: %v", err)
	actual, ok := msg.(messages.ClientConnectedMessage)
	assert.True(t, ok)
	assert.Equal(t, client2Id, actual.Client)

	service.Stop()
	wg.Wait()
}

func TestUnit_ChatService_OnConnect_DoesNotSendMessageToSelf(t *testing.T) {
	service, callbacks := newTestChatService()
	_, server1 := newTestConnection(t, 6001)
	client2, server2 := newTestConnection(t, 6001)
	wg := asyncRunChatService(t, service)

	client1Id := uuid.New()
	accepted := callbacks.OnConnect(client1Id, server1)
	assert.True(t, accepted)

	time.Sleep(reasonableWaitTimeForOnConnectMessageToBeProcessed)

	client2Id := uuid.New()
	accepted = callbacks.OnConnect(client2Id, server2)
	assert.True(t, accepted)

	assertNoDataReceived(t, client2)

	service.Stop()
	wg.Wait()
}

func TestUnit_ChatService_OnDisconnect_SendsMessagesToOthers(t *testing.T) {
	service, callbacks := newTestChatService()
	_, server1 := newTestConnection(t, 6001)
	client2, server2 := newTestConnection(t, 6001)
	wg := asyncRunChatService(t, service)

	client1Id := uuid.New()
	accepted := callbacks.OnConnect(client1Id, server1)
	assert.True(t, accepted)

	time.Sleep(reasonableWaitTimeForOnConnectMessageToBeProcessed)

	client2Id := uuid.New()
	accepted = callbacks.OnConnect(client2Id, server2)
	assert.True(t, accepted)

	callbacks.OnDisconnect(client1Id)

	// Wait for the message to be processed
	time.Sleep(100 * time.Millisecond)

	data := readFromConnection(t, client2)
	msg, err := messages.Decode(data)
	assert.Nil(t, err, "Actual err: %v", err)
	actual, ok := msg.(messages.ClientDisconnectedMessage)
	assert.True(t, ok)
	assert.Equal(t, client1Id, actual.Client)

	service.Stop()
	wg.Wait()
}

func TestUnit_ChatService_OnDirectMessage_RoutesMessageToCorrectClient(t *testing.T) {
	service, callbacks := newTestChatService()
	client1, server1 := newTestConnection(t, 6002)
	client2, server2 := newTestConnection(t, 6002)
	client3, server3 := newTestConnection(t, 6002)
	wg := asyncRunChatService(t, service)

	client1Id := uuid.New()
	accepted := callbacks.OnConnect(client1Id, server1)
	assert.True(t, accepted)

	client2Id := uuid.New()
	accepted = callbacks.OnConnect(client2Id, server2)
	assert.True(t, accepted)

	client3Id := uuid.New()
	accepted = callbacks.OnConnect(client3Id, server3)
	assert.True(t, accepted)

	time.Sleep(reasonableWaitTimeForOnConnectMessageToBeProcessed)

	// Drain connections
	drainConnection(t, client1)
	drainConnection(t, client2)
	drainConnection(t, client3)

	msg := messages.NewDirectMessage(client1Id, client2Id, "Hello, world!")
	encoded, err := messages.Encode(msg)
	assert.Nil(t, err, "Actual err: %v", err)
	callbacks.OnReadData(client1Id, encoded)

	time.Sleep(100 * time.Millisecond)

	data := readFromConnection(t, client2)
	actual, err := messages.Decode(data)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, msg, actual)

	assertNoDataReceived(t, client3)

	service.Stop()
	wg.Wait()
}

func newTestChatService() (ChatService, clients.Callbacks) {
	service := NewChatService(logger.New(os.Stdout))
	return service, service.GenerateCallbacks()
}
