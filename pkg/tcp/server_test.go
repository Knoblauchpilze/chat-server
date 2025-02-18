package tcp

import (
	"context"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/stretchr/testify/assert"
)

const reasonableWaitTimeForServerToBeUp = 200 * time.Millisecond

func TestUnit_Server_StartAndStopWithContext(t *testing.T) {
	s, err := NewServer(newTestServerConfig(6000), logger.New(os.Stdout))
	assert.Nil(t, err, "Actual err: %v", err)
	cancellable, cancel := context.WithCancel(context.Background())

	asyncCancelContext(reasonableWaitTimeForServerToBeUp, cancel)

	err = s.Start(cancellable)

	assert.Nil(t, err, "Actual err: %v", err)
}

func TestUnit_Server_WhenStartedMultipleTimes_ExpectFailure(t *testing.T) {
	s, err := NewServer(newTestServerConfig(6001), logger.New(os.Stdout))
	assert.Nil(t, err, "Actual err: %v", err)
	cancellable, cancel := context.WithCancel(context.Background())

	// Start the first server: it should run without error.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := s.Start(cancellable)
		assert.Nil(t, err)
	}()

	// Wait a bit to be sure the first one is up.
	time.Sleep(100 * time.Millisecond)

	// Start the second server: this should fail.
	err = s.Start(context.Background())
	assert.True(t, errors.IsErrorWithCode(err, ErrAlreadyRunning))

	cancel()
	wg.Wait()
}

func TestUnit_Server_WhenPortIsNotFree_ExpectStartReturnsInitializationFailure(t *testing.T) {
	log := logger.New(os.Stdout)
	_, err1 := NewServer(newTestServerConfig(6002), log)
	_, err2 := NewServer(newTestServerConfig(6002), log)

	assert.Nil(t, err1, "Actual err: %v", err1)
	assert.True(
		t,
		errors.IsErrorWithCode(err2, ErrTcpInitialization),
		"Actual err: %v",
		err2,
	)
}

func TestUnit_Server_ConnectDisconnect(t *testing.T) {
	s, err := NewServer(newTestServerConfig(6004), logger.New(os.Stdout))
	assert.Nil(t, err, "Actual err: %v", err)
	cancellable, cancel := context.WithCancel(context.Background())

	wg := asyncRunServerAndWaitForItToBeUp(t, s, cancellable)

	conn, err := net.Dial("tcp", ":6004")
	assert.Nil(t, err, "Unexpected dial error: %v", err)

	conn.Close()

	cancel()
	wg.Wait()
}

func TestUnit_Server_WhenServerIsClosed_ExpectConnectionToBeClosed(t *testing.T) {
	s, err := NewServer(newTestServerConfig(6005), logger.New(os.Stdout))
	assert.Nil(t, err, "Actual err: %v", err)
	cancellable, cancel := context.WithCancel(context.Background())

	wg := asyncRunServerAndWaitForItToBeUp(t, s, cancellable)

	conn, err := net.Dial("tcp", ":6005")
	assert.Nil(t, err, "Unexpected dial error: %v", err)

	cancel()
	wg.Wait()

	assertConnectionIsClosed(t, conn)
}

func TestUnit_Server_WhenServerIsClosed_ConnectionAreNotAcceptedAnymore(t *testing.T) {
	s, err := NewServer(newTestServerConfig(6005), logger.New(os.Stdout))
	assert.Nil(t, err, "Actual err: %v", err)
	cancellable, cancel := context.WithCancel(context.Background())

	wg := asyncRunServerAndWaitForItToBeUp(t, s, cancellable)

	cancel()
	wg.Wait()

	_, err = net.Dial("tcp", ":6005")
	assert.Regexp(t, "dial.* connection refused", err.Error(), "Actual err: %v", err)
}

func newTestServerConfig(port uint16) ServerConfiguration {
	return ServerConfiguration{
		Port: port,
	}
}

func asyncCancelContext(delay time.Duration, cancel context.CancelFunc) {
	go func() {
		time.Sleep(delay)
		cancel()
	}()
}

func asyncRunServerAndWaitForItToBeUp(
	t *testing.T,
	s Server,
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
		err = s.Start(ctx)
		assert.Nil(t, err, "Actual err: %v", err)
	}()

	time.Sleep(reasonableWaitTimeForServerToBeUp)

	return &wg
}
