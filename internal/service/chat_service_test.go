package service

import (
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/Knoblauchpilze/chat-server/pkg/clients"
	"github.com/Knoblauchpilze/chat-server/pkg/messages"
	"github.com/stretchr/testify/assert"
)

const (
	reasonableWaitTimeForServiceToBeUp                 = 50 * time.Millisecond
	reasonableWaitTimeForOnConnectMessageToBeProcessed = 100 * time.Millisecond
	reasonableReadTimeout                              = 100 * time.Millisecond
	reasonableReadSizeInBytes                          = 1024
	reasonableConnectTimeout                           = 100 * time.Millisecond
)

func TestUnit_ChatService_OnConnect_SendsMessagesToOthers(t *testing.T) {
	service, callbacks := newTestChatService()
	client1, server1 := newTestConnection(t, 6000)
	_, server2 := newTestConnection(t, 6000)
	wg := asyncRunChatService(t, service)

	accepted, _ := callbacks.OnConnect(server1)
	assert.True(t, accepted)

	time.Sleep(reasonableWaitTimeForOnConnectMessageToBeProcessed)

	accepted, client2Id := callbacks.OnConnect(server2)
	assert.True(t, accepted)

	data := readFromConnection(t, client1)
	msg, decoded, err := messages.Decode(data)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, len(data), decoded)
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

	accepted, _ := callbacks.OnConnect(server1)
	assert.True(t, accepted)

	time.Sleep(reasonableWaitTimeForOnConnectMessageToBeProcessed)

	accepted, _ = callbacks.OnConnect(server2)
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

	accepted, client1Id := callbacks.OnConnect(server1)
	assert.True(t, accepted)

	time.Sleep(reasonableWaitTimeForOnConnectMessageToBeProcessed)

	accepted, _ = callbacks.OnConnect(server2)
	assert.True(t, accepted)

	callbacks.OnDisconnect(client1Id)

	// Wait for the message to be processed
	time.Sleep(100 * time.Millisecond)

	data := readFromConnection(t, client2)
	msg, decoded, err := messages.Decode(data)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, len(data), decoded)
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

	accepted, client1Id := callbacks.OnConnect(server1)
	assert.True(t, accepted)

	accepted, client2Id := callbacks.OnConnect(server2)
	assert.True(t, accepted)

	accepted, _ = callbacks.OnConnect(server3)
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
	actual, decoded, err := messages.Decode(data)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, len(data), decoded)
	assert.Equal(t, msg, actual)

	assertNoDataReceived(t, client3)

	service.Stop()
	wg.Wait()
}

func newTestChatService() (ChatService, clients.Callbacks) {
	service := NewChatService(reasonableConnectTimeout, logger.New(os.Stdout))
	return service, service.GenerateCallbacks()
}

func asyncRunChatService(
	t *testing.T, service ChatService,
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

func newTestConnection(
	t *testing.T,
	port uint16,
) (client net.Conn, server net.Conn) {
	address := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", address)
	assert.Nil(t, err, "Actual err: %v", err)

	var wg sync.WaitGroup
	wg.Add(1)
	asyncConnect := func() {
		defer wg.Done()

		// Wait for the listener to be started in the main thread.
		time.Sleep(50 * time.Millisecond)

		var dialErr error
		client, dialErr = net.Dial("tcp", address)
		assert.Nil(t, dialErr, "Actual err: %v", dialErr)
	}

	go asyncConnect()

	server, err = listener.Accept()
	assert.Nil(t, err, "Actual err: %v", err)

	wg.Wait()

	listener.Close()

	return
}

func readFromConnection(t *testing.T, conn net.Conn) []byte {
	conn.SetReadDeadline(time.Now().Add(reasonableReadTimeout))

	out := make([]byte, reasonableReadSizeInBytes)
	n, err := conn.Read(out)
	assert.Nil(t, err, "Actual err: %v", err)

	return out[:n]
}

func drainConnection(t *testing.T, conn net.Conn) []byte {
	conn.SetReadDeadline(time.Now().Add(reasonableReadTimeout))

	out := make([]byte, reasonableReadSizeInBytes)
	n, err := conn.Read(out)
	if err != nil && err != io.EOF {
		assert.Nil(t, err, "Actual err: %v", err)
	}

	return out[:n]
}

func assertNoDataReceived(t *testing.T, conn net.Conn) {
	conn.SetReadDeadline(time.Now().Add(reasonableReadTimeout))

	oneByte := make([]byte, 1)
	_, err := conn.Read(oneByte)

	opErr, ok := err.(*net.OpError)
	assert.True(t, ok)
	if ok {
		assert.True(t, opErr.Timeout())
	}
}
