package clients

import (
	"os"
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/Knoblauchpilze/chat-server/pkg/messages"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var sampleUuid = uuid.MustParse("2dbf2622-2a95-4bd1-9b38-2f7b4ce65ffe")

func TestUnit_ClientManager_WhenClientConnects_ExpectMessageToBeSent(t *testing.T) {
	queue := make(chan messages.Message, 1)
	manager := NewManager(queue, logger.New(os.Stdout))

	clientId := uuid.New()
	actual := manager.OnConnect(clientId, nil)

	msg := <-queue

	assert.True(t, actual)
	connectMsg, ok := msg.(messages.ClientConnectedMessage)
	assert.True(t, ok)
	assert.Equal(t, messages.CLIENT_CONNECTED, msg.Type())
	assert.Equal(t, clientId, connectMsg.Client)
}

func TestUnit_ClientManager_WhenClientDisconnects_ExpectMessageToBeSent(t *testing.T) {
	queue := make(chan messages.Message, 1)
	manager := NewManager(queue, logger.New(os.Stdout))

	clientId := uuid.New()
	manager.OnConnect(clientId, nil)
	<-queue

	manager.OnDisconnect(clientId)
	msg := <-queue

	disconnectMsg, ok := msg.(messages.ClientDisconnectedMessage)
	assert.True(t, ok)
	assert.Equal(t, messages.CLIENT_DISCONNECTED, msg.Type())
	assert.Equal(t, clientId, disconnectMsg.Client)
}

func TestUnit_ClientManager_WhenReadErrorDetected_ExpectDisconnectMessageToBeSent(t *testing.T) {
	queue := make(chan messages.Message, 1)
	manager := NewManager(queue, logger.New(os.Stdout))

	clientId := uuid.New()
	manager.OnConnect(clientId, nil)
	<-queue

	manager.OnReadError(clientId, errSample)
	msg := <-queue

	disconnectMsg, ok := msg.(messages.ClientDisconnectedMessage)
	assert.True(t, ok)
	assert.Equal(t, messages.CLIENT_DISCONNECTED, msg.Type())
	assert.Equal(t, clientId, disconnectMsg.Client)
}

func TestUnit_ClientManager_WhenReadErrorDetected_ExpectMessageIsNotReceivedAnymore(t *testing.T) {
	queue := make(chan messages.Message, 1)
	manager := NewManager(queue, logger.New(os.Stdout))

	client1 := uuid.New()
	clientConn1, serverConn1 := newTestConnection(t, 7500)
	manager.OnConnect(client1, serverConn1)
	<-queue

	manager.OnReadError(client1, errSample)

	msg := messages.NewClientConnectedMessage(sampleUuid)
	manager.Broadcast(msg)

	time.Sleep(100 * time.Millisecond)

	assertNoDataReceived(t, clientConn1)
}

func TestUnit_ClientManager_WhenMessageIsBroadcast_ExpectMessagesToBeSent(t *testing.T) {
	queue := make(chan messages.Message, 1)
	manager := NewManager(queue, logger.New(os.Stdout))

	client1 := uuid.New()
	clientConn1, serverConn1 := newTestConnection(t, 7501)
	manager.OnConnect(client1, serverConn1)
	<-queue

	client2 := uuid.New()
	clientConn2, serverConn2 := newTestConnection(t, 7501)
	manager.OnConnect(client2, serverConn2)
	<-queue

	client3 := sampleUuid
	msg := messages.NewClientConnectedMessage(client3)
	manager.Broadcast(msg)

	time.Sleep(100 * time.Millisecond)

	expectedMessage := []byte{
		// CLIENT_CONNECTED
		0x0, 0x0, 0x0, 0x0,
		// Client id
		0x2d, 0xbf, 0x26, 0x22, 0x2a, 0x95, 0x4b, 0xd1, 0x9b, 0x38, 0x2f, 0x7b, 0x4c, 0xe6, 0x5f, 0xfe,
	}

	actual1 := readFromConnection(t, clientConn1)
	assert.Equal(t, expectedMessage, actual1)

	actual2 := readFromConnection(t, clientConn2)
	assert.Equal(t, expectedMessage, actual2)
}

func TestUnit_ClientManager_WhenMessageIsBroadcastExceptToOneClient_ExpectNothingReceivedForThisClient(t *testing.T) {
	queue := make(chan messages.Message, 1)
	manager := NewManager(queue, logger.New(os.Stdout))

	client1 := uuid.New()
	clientConn1, serverConn1 := newTestConnection(t, 7502)
	manager.OnConnect(client1, serverConn1)
	<-queue

	client2 := uuid.New()
	clientConn2, serverConn2 := newTestConnection(t, 7502)
	manager.OnConnect(client2, serverConn2)
	<-queue

	msg := messages.NewClientConnectedMessage(sampleUuid)
	manager.BroadcastExcept(client2, msg)

	time.Sleep(100 * time.Millisecond)

	actual1 := readFromConnection(t, clientConn1)
	expectedMessage := []byte{
		// CLIENT_CONNECTED
		0x0, 0x0, 0x0, 0x0,
		// Client id
		0x2d, 0xbf, 0x26, 0x22, 0x2a, 0x95, 0x4b, 0xd1, 0x9b, 0x38, 0x2f, 0x7b, 0x4c, 0xe6, 0x5f, 0xfe,
	}
	assert.Equal(t, expectedMessage, actual1)

	assertNoDataReceived(t, clientConn2)
}

func TestUnit_ClientManager_WhenMessageIsSentToClient_ExpectMessageIsSent(t *testing.T) {
	queue := make(chan messages.Message, 1)
	manager := NewManager(queue, logger.New(os.Stdout))

	client1 := uuid.New()
	clientConn1, serverConn1 := newTestConnection(t, 7503)
	manager.OnConnect(client1, serverConn1)
	<-queue

	msg := messages.NewClientConnectedMessage(sampleUuid)
	manager.SendTo(client1, msg)

	time.Sleep(100 * time.Millisecond)

	actual1 := readFromConnection(t, clientConn1)
	expectedMessage := []byte{
		// CLIENT_CONNECTED
		0x0, 0x0, 0x0, 0x0,
		// Client id
		0x2d, 0xbf, 0x26, 0x22, 0x2a, 0x95, 0x4b, 0xd1, 0x9b, 0x38, 0x2f, 0x7b, 0x4c, 0xe6, 0x5f, 0xfe,
	}
	assert.Equal(t, expectedMessage, actual1)
}

func TestUnit_ClientManager_WhenMessageIsSentToClient_ExpectNobodyElseReceivesIt(t *testing.T) {
	queue := make(chan messages.Message, 1)
	manager := NewManager(queue, logger.New(os.Stdout))

	client1 := uuid.New()
	_, serverConn1 := newTestConnection(t, 7504)
	manager.OnConnect(client1, serverConn1)
	<-queue

	client2 := uuid.New()
	clientConn2, serverConn2 := newTestConnection(t, 7504)
	manager.OnConnect(client2, serverConn2)
	<-queue

	msg := messages.NewClientConnectedMessage(uuid.New())
	manager.SendTo(client1, msg)

	time.Sleep(100 * time.Millisecond)

	assertNoDataReceived(t, clientConn2)
}

func TestUnit_ClientManager_WhenClientDisconnects_ExpectMessageIsNotReceivedAnymore(t *testing.T) {
	queue := make(chan messages.Message, 1)
	manager := NewManager(queue, logger.New(os.Stdout))

	client1 := uuid.New()
	clientConn1, serverConn1 := newTestConnection(t, 7505)
	manager.OnConnect(client1, serverConn1)
	<-queue

	manager.OnDisconnect(client1)

	msg := messages.NewClientConnectedMessage(sampleUuid)
	manager.Broadcast(msg)

	time.Sleep(100 * time.Millisecond)

	assertNoDataReceived(t, clientConn1)
}
