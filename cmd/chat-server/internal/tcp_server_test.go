package internal

import (
	"context"
	"net"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/Knoblauchpilze/chat-server/pkg/messages"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const reasonableTimeForConnectionToBeProcessed = 100 * time.Millisecond

func TestUnit_RunTcpServer_OnConnect_ShouldBeAccepted(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())
	config := newTcpTestConfig(7100)

	wg := asyncRunTcpServer(t, config, cancellable)

	conn, err := net.Dial("tcp", ":7100")
	assert.Nil(t, err, "Actual err: %v", err)

	time.Sleep(100 * time.Millisecond)
	assertConnectionIsStillOpen(t, conn)

	cancel()
	wg.Wait()
}

func TestUnit_RunTcpServer_WhenServerCloses_ExpectConnectionToBeClosed(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())
	config := newTcpTestConfig(7101)

	wg := asyncRunTcpServer(t, config, cancellable)

	conn, err := net.Dial("tcp", ":7101")
	assert.Nil(t, err, "Actual err: %v", err)

	cancel()
	wg.Wait()
	assertConnectionIsClosed(t, conn)
}

func TestUnit_RunTcpServer_OnConnect_ExpectOthersAreNotified(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())
	config := newTcpTestConfig(7103)

	wg := asyncRunTcpServer(t, config, cancellable)

	conn1, err := net.Dial("tcp", ":7103")
	assert.Nil(t, err, "Actual err: %v", err)

	time.Sleep(reasonableTimeForConnectionToBeProcessed)

	conn2, err := net.Dial("tcp", ":7103")
	assert.Nil(t, err, "Actual err: %v", err)

	data := readFromConnection(t, conn1)
	msg, decoded, err := messages.Decode(data)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, len(data), decoded)
	assert.Equal(t, messages.CLIENT_CONNECTED, msg.Type())
	assertNoDataReceived(t, conn2)

	cancel()
	wg.Wait()
}

func TestUnit_RunTcpServer_OnDisconnect_ExpectOthersAreNotified(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())
	config := newTcpTestConfig(7104)

	wg := asyncRunTcpServer(t, config, cancellable)

	conn1, err := net.Dial("tcp", ":7104")
	assert.Nil(t, err, "Actual err: %v", err)
	conn2, err := net.Dial("tcp", ":7104")
	assert.Nil(t, err, "Actual err: %v", err)

	time.Sleep(reasonableTimeForConnectionToBeProcessed)
	drainConnection(t, conn1)
	drainConnection(t, conn2)

	conn1.Close()
	time.Sleep(reasonableTimeForConnectionToBeProcessed)

	data := readFromConnection(t, conn2)
	msg, decoded, err := messages.Decode(data)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, len(data), decoded)
	assert.Equal(t, messages.CLIENT_DISCONNECTED, msg.Type())

	cancel()
	wg.Wait()
}

func TestUnit_RunTcpServer_WhenSendingMessageToClient_ExpectOnlyItReceivesIt(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())
	config := newTcpTestConfig(7105)

	wg := asyncRunTcpServer(t, config, cancellable)

	// Connect client 1
	conn1, err := net.Dial("tcp", ":7105")
	assert.Nil(t, err, "Actual err: %v", err)

	time.Sleep(reasonableTimeForConnectionToBeProcessed)

	// Connect client 2
	conn2, err := net.Dial("tcp", ":7105")
	assert.Nil(t, err, "Actual err: %v", err)
	time.Sleep(reasonableTimeForConnectionToBeProcessed)

	// Fetch id of client 2
	data := readFromConnection(t, conn1)
	msg, decoded, err := messages.Decode(data)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, len(data), decoded)
	connected, err := messages.ToMessageStruct[messages.ClientConnectedMessage](msg)
	assert.Nil(t, err, "Actual err: %v", err)
	clientId2 := connected.Client

	assertNoDataReceived(t, conn2)

	// Send message to client 2 from client 1's connection
	dummyIdForClient1 := uuid.New()
	msg = messages.NewDirectMessage(dummyIdForClient1, clientId2, "Hello, client 2")
	out, err := messages.Encode(msg)

	assert.Nil(t, err, "Actual err: %v", err)
	n, err := conn1.Write(out)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, len(out), n)

	time.Sleep(100 * time.Millisecond)

	// Read message from client 2
	data = readFromConnection(t, conn2)
	msg, decoded, err = messages.Decode(data)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, len(data), decoded)
	actual, err := messages.ToMessageStruct[messages.DirectMessage](msg)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, dummyIdForClient1, actual.Emitter)
	assert.Equal(t, clientId2, actual.Receiver)
	assert.Equal(t, "Hello, client 2", actual.Content)

	cancel()
	wg.Wait()
}

func TestUnit_RunTcpServer_WhenSendingGarbage_ExpectConnectionToStayOpen(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())
	config := newTcpTestConfig(7102)

	wg := asyncRunTcpServer(t, config, cancellable)

	conn, err := net.Dial("tcp", ":7102")
	assert.Nil(t, err, "Actual err: %v", err)
	_, err = conn.Write([]byte("garbage"))
	assert.Nil(t, err, "Actual err: %v", err)

	// Wait long enough for the read timeout to expire and connection
	// to be effectively closed if it does terminate.
	time.Sleep(2 * time.Second)

	assertConnectionIsStillOpen(t, conn)

	cancel()
	wg.Wait()
}

func TestUnit_RunTcpServer_WhenClientIsSendingTooMuchGarbage_ExpectDisconnected(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())
	config := newTcpTestConfig(7106)

	wg := asyncRunTcpServer(t, config, cancellable)

	conn, err := net.Dial("tcp", ":7106")
	assert.Nil(t, err, "Actual err: %v", err)

	time.Sleep(reasonableTimeForConnectionToBeProcessed)

	// Generate a garbage string that can't be interpreted as a valid message.
	out := []byte(strings.Repeat("abc123def456ghi789kl", 60))

	n, err := conn.Write(out)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, len(out), n)

	// Wait long enough for the connection to be terminated.
	time.Sleep(200 * time.Millisecond)

	assertConnectionIsClosed(t, conn)

	cancel()
	wg.Wait()
}

func asyncRunTcpServer(
	t *testing.T,
	config Configuration,
	ctx context.Context,
) *sync.WaitGroup {
	var err error
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if panicErr := recover(); panicErr != nil {
				assert.Failf(t, "Server panicked", "Panic details: %v", panicErr)
			}
		}()

		err = RunTcpServer(ctx, config, logger.New(os.Stdout))
		assert.Nil(t, err, "Actual err: %v", err)
	}()

	time.Sleep(reasonableWaitTimeForServerToBeUp)

	return &wg
}
