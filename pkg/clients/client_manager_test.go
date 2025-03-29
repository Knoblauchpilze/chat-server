package clients

import (
	"net"
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
	manager := newTestClientManager(queue)

	actual, clientId := manager.OnConnect(nil)

	msg := <-queue

	assert.True(t, actual)
	assert.Equal(t, sampleUuid, clientId)

	actualMsg, ok := msg.(messages.ClientConnectedMessage)
	assert.True(t, ok)
	assert.Equal(t, messages.CLIENT_CONNECTED, actualMsg.Type())
	assert.Equal(t, clientId, actualMsg.Client)
}

func TestUnit_ClientManager_WhenClientDisconnects_ExpectMessageToBeSent(t *testing.T) {
	queue := make(chan messages.Message, 1)
	manager := newTestClientManager(queue)

	_, clientId := manager.OnConnect(nil)
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
	manager := newTestClientManager(queue)

	_, clientId := manager.OnConnect(nil)
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
	manager := newTestClientManager(queue)

	clientConn, serverConn := newTestConnection(t, 7500)
	_, clientId := manager.OnConnect(serverConn)
	<-queue

	manager.OnReadError(clientId, errSample)

	dummyId := uuid.New()
	assert.NotEqual(t, dummyId, clientId)
	msg := messages.NewClientConnectedMessage(dummyId)
	manager.Broadcast(msg)

	time.Sleep(100 * time.Millisecond)

	assertNoDataReceived(t, clientConn)
}

func TestUnit_ClientManager_WhenMessageIsBroadcast_ExpectMessagesToBeSent(t *testing.T) {
	queue := make(chan messages.Message, 1)
	manager := newTestClientManagerWithHandshakeFunc(
		queue, newHandshakeFuncWithRandomUuid(),
	)

	clientConn1, serverConn1 := newTestConnection(t, 7501)
	manager.OnConnect(serverConn1)
	<-queue

	clientConn2, serverConn2 := newTestConnection(t, 7501)
	manager.OnConnect(serverConn2)
	<-queue

	client3 := sampleUuid
	msg := messages.NewClientConnectedMessage(client3)
	manager.Broadcast(msg)

	time.Sleep(100 * time.Millisecond)

	expected, err := messages.Encode(msg)
	assert.Nil(t, err, "Actual err: %v", err)

	actual1 := readFromConnection(t, clientConn1)
	assert.Equal(t, expected, actual1)
	actual2 := readFromConnection(t, clientConn2)
	assert.Equal(t, expected, actual2)
}

func TestUnit_ClientManager_WhenMessageIsBroadcastExceptToOneClient_ExpectNothingReceivedForThisClient(t *testing.T) {
	queue := make(chan messages.Message, 1)
	manager := newTestClientManagerWithHandshakeFunc(
		queue, newHandshakeFuncWithRandomUuid(),
	)

	clientConn1, serverConn1 := newTestConnection(t, 7502)
	manager.OnConnect(serverConn1)
	<-queue

	clientConn2, serverConn2 := newTestConnection(t, 7502)
	_, client2 := manager.OnConnect(serverConn2)
	<-queue

	msg := messages.NewClientConnectedMessage(sampleUuid)
	manager.BroadcastExcept(client2, msg)

	time.Sleep(100 * time.Millisecond)

	actual1 := readFromConnection(t, clientConn1)

	expected, err := messages.Encode(msg)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, expected, actual1)
	assertNoDataReceived(t, clientConn2)
}

func TestUnit_ClientManager_WhenMessageIsSentToClient_ExpectMessageIsSent(t *testing.T) {
	queue := make(chan messages.Message, 1)
	manager := newTestClientManager(queue)

	clientConn, serverConn := newTestConnection(t, 7503)
	_, client := manager.OnConnect(serverConn)
	<-queue

	msg := messages.NewClientConnectedMessage(sampleUuid)
	manager.SendTo(client, msg)

	time.Sleep(100 * time.Millisecond)

	actual1 := readFromConnection(t, clientConn)

	expected, err := messages.Encode(msg)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, expected, actual1)
}

func TestUnit_ClientManager_WhenMessageIsSentToClient_ExpectNobodyElseReceivesIt(t *testing.T) {
	queue := make(chan messages.Message, 1)
	manager := newTestClientManagerWithHandshakeFunc(
		queue, newHandshakeFuncWithRandomUuid(),
	)

	_, serverConn1 := newTestConnection(t, 7504)
	_, client1 := manager.OnConnect(serverConn1)
	<-queue

	clientConn2, serverConn2 := newTestConnection(t, 7504)
	manager.OnConnect(serverConn2)
	<-queue

	msg := messages.NewClientConnectedMessage(uuid.New())
	manager.SendTo(client1, msg)

	time.Sleep(100 * time.Millisecond)

	assertNoDataReceived(t, clientConn2)
}

func TestUnit_ClientManager_WhenClientDisconnects_ExpectMessageIsNotReceivedAnymore(t *testing.T) {
	queue := make(chan messages.Message, 1)
	manager := newTestClientManager(queue)

	clientConn1, serverConn1 := newTestConnection(t, 7505)
	_, client1 := manager.OnConnect(serverConn1)
	<-queue

	manager.OnDisconnect(client1)

	msg := messages.NewClientConnectedMessage(sampleUuid)
	manager.Broadcast(msg)

	time.Sleep(100 * time.Millisecond)

	assertNoDataReceived(t, clientConn1)
}

func TestUnit_ClientManager_WhenHandshakeFails_ExpectClientIsDenied(t *testing.T) {
	hanshake := func(net.Conn, time.Duration) (uuid.UUID, error) {
		return uuid.Nil, errSample
	}
	// In case a message is published this will hang
	queue := make(chan messages.Message)
	manager := newTestClientManagerWithHandshakeFunc(queue, hanshake)

	actual, _ := manager.OnConnect(nil)

	assert.False(t, actual)
}

func TestUnit_ClientManager_WhenHandshakePanics_ExpectClientIsDenied(t *testing.T) {
	hanshake := func(net.Conn, time.Duration) (uuid.UUID, error) {
		panic(errSample)
	}
	// In case a message is published this will hang
	queue := make(chan messages.Message)
	manager := newTestClientManagerWithHandshakeFunc(queue, hanshake)

	actual, _ := manager.OnConnect(nil)

	assert.False(t, actual)
}

func newTestHandshakeFunc(id uuid.UUID) HandshakeFunc {
	return func(net.Conn, time.Duration) (uuid.UUID, error) {
		return id, nil
	}
}

func newHandshakeFuncWithRandomUuid() HandshakeFunc {
	return func(net.Conn, time.Duration) (uuid.UUID, error) {
		return uuid.New(), nil
	}
}

func newTestClientManager(queue chan messages.Message) Manager {
	return newTestClientManagerWithHandshakeFunc(queue, newTestHandshakeFunc(sampleUuid))
}

func newTestClientManagerWithHandshakeFunc(
	queue chan messages.Message, handshake HandshakeFunc,
) Manager {
	props := ManagerProps{
		Queue:          queue,
		ConnectTimeout: 100 * time.Millisecond,
		Handshake:      handshake,
		Log:            logger.New(os.Stdout),
	}

	return NewManager(props)
}
