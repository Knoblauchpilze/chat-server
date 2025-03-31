package internal

import (
	"context"
	"net"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/Knoblauchpilze/chat-server/internal/service"
	"github.com/Knoblauchpilze/chat-server/pkg/clients"
	"github.com/Knoblauchpilze/chat-server/pkg/messages"
	"github.com/Knoblauchpilze/chat-server/pkg/repositories"
	"github.com/stretchr/testify/assert"
)

const reasonableTimeForConnectionToBeProcessed = 100 * time.Millisecond

func TestIT_RunTcpServer_OnConnect_ShouldBeAccepted(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())
	config := newTcpTestConfig(7100)
	dbConn := newTestDbConnection(t)
	defer dbConn.Close(context.Background())

	wg := asyncRunTcpServer(t, config, dbConn, cancellable)

	conn, err := net.Dial("tcp", ":7100")
	assert.Nil(t, err, "Actual err: %v", err)

	time.Sleep(100 * time.Millisecond)
	assertConnectionIsStillOpen(t, conn)

	cancel()
	wg.Wait()
}

func TestIT_RunTcpServer_WhenServerCloses_ExpectConnectionToBeClosed(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())
	config := newTcpTestConfig(7101)
	dbConn := newTestDbConnection(t)
	defer dbConn.Close(context.Background())

	wg := asyncRunTcpServer(t, config, dbConn, cancellable)

	conn, err := net.Dial("tcp", ":7101")
	assert.Nil(t, err, "Actual err: %v", err)

	cancel()
	wg.Wait()
	assertConnectionIsClosed(t, conn)
}

func TestIT_RunTcpServer_OnConnect_ExpectOthersAreNotified(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())
	config := newTcpTestConfig(7103)
	dbConn := newTestDbConnection(t)
	defer dbConn.Close(context.Background())

	wg := asyncRunTcpServer(t, config, dbConn, cancellable)

	conn1, _ := connectToServerAndSendHandshake(t, 7103)

	time.Sleep(reasonableTimeForConnectionToBeProcessed)

	conn2, client2Id := connectToServerAndSendHandshake(t, 7103)

	data := readFromConnection(t, conn1)
	msg, decoded, err := messages.Decode(data)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, len(data), decoded)
	actualMsg, err := messages.ToMessageStruct[messages.ClientConnectedMessage](msg)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, client2Id, actualMsg.Client)
	assertNoDataReceived(t, conn2)

	cancel()
	wg.Wait()
}

func TestIT_RunTcpServer_OnDisconnect_ExpectOthersAreNotified(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())
	config := newTcpTestConfig(7104)
	dbConn := newTestDbConnection(t)
	defer dbConn.Close(context.Background())

	wg := asyncRunTcpServer(t, config, dbConn, cancellable)

	conn1, client1Id := connectToServerAndSendHandshake(t, 7104)
	conn2, _ := connectToServerAndSendHandshake(t, 7104)

	time.Sleep(reasonableTimeForConnectionToBeProcessed)
	drainConnection(t, conn1)
	drainConnection(t, conn2)

	conn1.Close()
	time.Sleep(reasonableTimeForConnectionToBeProcessed)

	data := readFromConnection(t, conn2)
	msg, decoded, err := messages.Decode(data)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, len(data), decoded)
	actualMsg, err := messages.ToMessageStruct[messages.ClientDisconnectedMessage](msg)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, client1Id, actualMsg.Client)

	cancel()
	wg.Wait()
}

func TestIT_RunTcpServer_WhenSendingMessageToClient_ExpectOnlyItReceivesIt(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())
	config := newTcpTestConfig(7105)
	dbConn := newTestDbConnection(t)
	defer dbConn.Close(context.Background())

	wg := asyncRunTcpServer(t, config, dbConn, cancellable)

	// Connect client 1
	conn1, client1Id := connectToServerAndSendHandshake(t, 7105)

	time.Sleep(reasonableTimeForConnectionToBeProcessed)

	// Connect client 2
	conn2, client2Id := connectToServerAndSendHandshake(t, 7105)
	time.Sleep(reasonableTimeForConnectionToBeProcessed)

	// Drain connection 1 so that no data is pending
	drainConnection(t, conn1)

	assertNoDataReceived(t, conn2)

	// Send message to client 2 from client 1's connection
	msg := messages.NewDirectMessage(client1Id, client2Id, "Hello, client 2")
	out, err := messages.Encode(msg)

	assert.Nil(t, err, "Actual err: %v", err)
	n, err := conn1.Write(out)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, len(out), n)

	time.Sleep(100 * time.Millisecond)

	// Read message from client 2
	data := readFromConnection(t, conn2)
	msg, decoded, err := messages.Decode(data)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, len(data), decoded)
	actual, err := messages.ToMessageStruct[messages.DirectMessage](msg)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, client1Id, actual.Emitter)
	assert.Equal(t, client2Id, actual.Receiver)
	assert.Equal(t, "Hello, client 2", actual.Content)

	cancel()
	wg.Wait()
}

func TestIT_RunTcpServer_WhenSendingGarbage_ExpectConnectionToStayOpen(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())
	config := newTcpTestConfig(7102)
	dbConn := newTestDbConnection(t)
	defer dbConn.Close(context.Background())

	wg := asyncRunTcpServer(t, config, dbConn, cancellable)

	conn, _ := connectToServerAndSendHandshake(t, 7102)
	_, err := conn.Write([]byte("garbage"))
	assert.Nil(t, err, "Actual err: %v", err)

	// Wait long enough for the read timeout to expire and connection
	// to be effectively closed if it does terminate.
	time.Sleep(2 * time.Second)

	assertConnectionIsStillOpen(t, conn)

	cancel()
	wg.Wait()
}

func TestIT_RunTcpServer_WhenClientIsSendingTooMuchGarbage_ExpectDisconnected(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())
	config := newTcpTestConfig(7106)
	dbConn := newTestDbConnection(t)
	defer dbConn.Close(context.Background())

	wg := asyncRunTcpServer(t, config, dbConn, cancellable)

	conn, _ := connectToServerAndSendHandshake(t, 7106)

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
	dbConn db.Connection,
	ctx context.Context,
) *sync.WaitGroup {
	var err error
	var wg sync.WaitGroup

	userRepo := repositories.NewUserRepository(dbConn)

	log := logger.New(os.Stdout)
	services := service.Services{
		Chat: service.NewChatService(
			clients.NewHandshake(
				userRepo,
				config.ConnectTimeout,
			),
			log,
		),
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if panicErr := recover(); panicErr != nil {
				assert.Failf(t, "Server panicked", "Panic details: %v", panicErr)
			}
		}()

		err = RunTcpServer(ctx, config, services, log)
		assert.Nil(t, err, "Actual err: %v", err)
	}()

	time.Sleep(reasonableWaitTimeForServerToBeUp)

	return &wg
}
