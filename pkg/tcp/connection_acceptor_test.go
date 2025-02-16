package tcp

import (
	"net"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/stretchr/testify/assert"
)

const reasonableWaitTimeForAcceptorToBeUp = 200 * time.Millisecond

func TestUnit_ConnectionAcceptor_ListenWithContext(t *testing.T) {
	ca := NewConnectionAcceptor(newTestAcceptorConfig(6000), logger.New(os.Stdout))

	wg := asyncRunAcceptorAndWaitForItToBeUp(t, ca)

	closeAcceptorAndAssertNoError(t, ca, wg)
}

func TestUnit_ConnectionAcceptor_WhenStartedMultipleTimes_ExpectFailure(t *testing.T) {
	ca := NewConnectionAcceptor(newTestAcceptorConfig(6001), logger.New(os.Stdout))

	// Start the first acceptor: it should run without error.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := ca.Accept()
		assert.Nil(t, err)
	}()

	// Wait a bit to be sure the first one is up.
	time.Sleep(100 * time.Millisecond)

	// Start the second acceptor: this should fail.
	err := ca.Accept()
	assert.True(t, errors.IsErrorWithCode(err, ErrAlreadyListening))

	closeAcceptorAndAssertNoError(t, ca, &wg)
}

func TestUnit_ConnectionAcceptor_WhenNotStarted_StopDoesNotFail(t *testing.T) {
	ca := NewConnectionAcceptor(newTestAcceptorConfig(6001), logger.New(os.Stdout))

	err := ca.Close()

	assert.Nil(t, err, "Actual err: %v", err)
}

func TestUnit_ConnectionAcceptor_WhenPortIsNotFree_ExpectStartReturnsInitializationFailure(t *testing.T) {
	log := logger.New(os.Stdout)
	ca1 := NewConnectionAcceptor(newTestAcceptorConfig(6002), log)
	ca2 := NewConnectionAcceptor(newTestAcceptorConfig(6002), log)

	wg := asyncRunAcceptorAndWaitForItToBeUp(t, ca1)

	acceptErr := ca2.Accept()

	closeAcceptorAndAssertNoError(t, ca1, wg)
	assert.True(
		t,
		errors.IsErrorWithCode(acceptErr, ErrTcpInitialization),
		"Actual err: %v",
		acceptErr,
	)
}

func TestUnit_ConnectionAcceptor_WhenPortIsNotFree_ExpectStopDoesNotCrash(t *testing.T) {
	log := logger.New(os.Stdout)
	ca1 := NewConnectionAcceptor(newTestAcceptorConfig(6002), log)
	ca2 := NewConnectionAcceptor(newTestAcceptorConfig(6002), log)

	wg := asyncRunAcceptorAndWaitForItToBeUp(t, ca1)

	acceptErr := ca2.Accept()
	assert.NotNil(t, acceptErr)

	closeAcceptorAndAssertNoError(t, ca1, wg)
	err := ca2.Close()

	assert.Nil(t, err, "Actual err: %v", err)
}

func TestUnit_ConnectionAcceptor_WhenAcceptorIsStopped_ExpectConnectionToNotBeClosed(t *testing.T) {
	ca := NewConnectionAcceptor(newTestAcceptorConfig(6003), logger.New(os.Stdout))

	wg := asyncRunAcceptorAndWaitForItToBeUp(t, ca)

	conn, err := net.Dial("tcp", ":6003")
	assert.Nil(t, err, "Actual err: %v", err)

	closeAcceptorAndAssertNoError(t, ca, wg)

	assertConnectionIsStillOpen(t, conn)
}

func TestUnit_ConnectionAcceptor_ConnectDisconnect(t *testing.T) {
	ca := NewConnectionAcceptor(newTestAcceptorConfig(6004), logger.New(os.Stdout))

	wg := asyncRunAcceptorAndWaitForItToBeUp(t, ca)

	conn, err := net.Dial("tcp", ":6004")
	assert.Nil(t, err, "Unexpected dial error: %v", err)

	conn.Close()

	closeAcceptorAndAssertNoError(t, ca, wg)
}

func TestUnit_ConnectionAcceptor_OnConnect_ExpectCallbackNotified(t *testing.T) {
	config := newTestAcceptorConfig(6005)
	var called int
	config.Callbacks.ConnectCallback = func(conn net.Conn) {
		called++
	}

	ca := NewConnectionAcceptor(config, logger.New(os.Stdout))

	wg := asyncRunAcceptorAndWaitForItToBeUp(t, ca)

	conn, err := net.Dial("tcp", ":6005")
	assert.Nil(t, err, "Unexpected dial error: %v", err)

	conn.Close()

	closeAcceptorAndAssertNoError(t, ca, wg)
	assert.Equal(t, 1, called)
}

func TestUnit_ConnectionAcceptor_WhenOnConnectCallbackSucceeds_ExpectConnectionToStayOpen(t *testing.T) {
	config := newTestAcceptorConfig(6006)
	var called int
	config.Callbacks.ConnectCallback = func(conn net.Conn) {
		called++
	}

	ca := NewConnectionAcceptor(config, logger.New(os.Stdout))

	wg := asyncRunAcceptorAndWaitForItToBeUp(t, ca)

	conn, err := net.Dial("tcp", ":6006")
	assert.Nil(t, err, "Unexpected dial error: %v", err)

	assertConnectionIsStillOpen(t, conn)

	closeAcceptorAndAssertNoError(t, ca, wg)
	assert.Equal(t, 1, called)
}

func TestUnit_ConnectionAcceptor_WhenOnConnectCallbackFails_ExpectConnectionToBeClosed(t *testing.T) {
	config := newTestAcceptorConfig(6007)
	config.Callbacks.ConnectCallback = func(conn net.Conn) {
		panic(errSample)
	}

	ca := NewConnectionAcceptor(config, logger.New(os.Stdout))

	wg := asyncRunAcceptorAndWaitForItToBeUp(t, ca)

	conn, err := net.Dial("tcp", ":6007")
	assert.Nil(t, err, "Unexpected dial error: %v", err)

	// Allow the callback to be processed.
	time.Sleep(50 * time.Millisecond)

	assertConnectionIsClosed(t, conn)

	closeAcceptorAndAssertNoError(t, ca, wg)
}

func TestUnit_ConnectionAcceptor_WhenOnConnectCallbackFails_ExpectAcceptorStillAcceptsOtherConnections(t *testing.T) {
	config := newTestAcceptorConfig(6008)
	var called atomic.Int32
	var doPanic atomic.Bool
	doPanic.Store(true)
	config.Callbacks.ConnectCallback = func(conn net.Conn) {
		called.Add(1)
		if doPanic.Load() {
			doPanic.Store(false)
			panic(errSample)
		}
	}

	ca := NewConnectionAcceptor(config, logger.New(os.Stdout))

	wg := asyncRunAcceptorAndWaitForItToBeUp(t, ca)

	// First connection should panic and be closed
	conn, err := net.Dial("tcp", ":6008")
	assert.Nil(t, err, "Unexpected dial error: %v", err)
	time.Sleep(50 * time.Millisecond)
	assertConnectionIsClosed(t, conn)

	// Second connection should work
	conn, err = net.Dial("tcp", ":6008")
	assert.Nil(t, err, "Unexpected dial error: %v", err)
	time.Sleep(50 * time.Millisecond)
	assertConnectionIsStillOpen(t, conn)

	closeAcceptorAndAssertNoError(t, ca, wg)
	assert.Equal(t, int32(2), called.Load())
}

func TestUnit_ConnectionAcceptor_WhenOnConnectTakesLong_ExpectOtherConnectionsCanBeProcessedConcurrently(t *testing.T) {
	config := newTestAcceptorConfig(6009)
	var called atomic.Int32
	config.Callbacks.ConnectCallback = func(conn net.Conn) {
		time.Sleep(200 * time.Millisecond)
		called.Add(1)
	}

	ca := NewConnectionAcceptor(config, logger.New(os.Stdout))

	wg := asyncRunAcceptorAndWaitForItToBeUp(t, ca)

	wgConn1 := asyncOpenConnectionAndCloseIt(t, 6009)
	wgConn2 := asyncOpenConnectionAndCloseIt(t, 6009)

	start := time.Now()

	wgConn2.Wait()
	wgConn1.Wait()

	// Wait for connections to be processed before closing.
	time.Sleep(100 * time.Millisecond)

	closeAcceptorAndAssertNoError(t, ca, wg)

	elapsed := time.Since(start)

	assert.Equal(t, int32(2), called.Load())
	// If this would not be concurrent we would expect at least 400ms of
	// processing time for the callbacks.
	assert.LessOrEqual(t, elapsed, 300*time.Millisecond)
}

func newTestAcceptorConfig(port uint16) AcceptorConfig {
	return AcceptorConfig{
		Port: port,
	}
}

func asyncRunAcceptorAndWaitForItToBeUp(
	t *testing.T,
	ca ConnectionAcceptor,
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
		err := ca.Accept()
		assert.Nil(t, err, "Actual err: %v", err)
	}()

	time.Sleep(reasonableWaitTimeForAcceptorToBeUp)

	return &wg
}

func closeAcceptorAndAssertNoError(t *testing.T, ca ConnectionAcceptor, wg *sync.WaitGroup) {
	err := ca.Close()
	wg.Wait()
	assert.Nil(t, err, "Actual err: %v", err)
}
