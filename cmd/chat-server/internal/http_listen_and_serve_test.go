package internal

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/stretchr/testify/assert"
)

const (
	success = "SUCCESS"
	failure = "ERROR"
)

func TestIT_HttpListenAndServer_StartAndStopWithContext(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())
	asyncCancelContext(200*time.Millisecond, cancel)

	conn := newTestDbConnection(t)
	defer conn.Close(context.Background())

	conf := newHttpTestConfig(7200)

	err := httpListenAndServe(cancellable, conf, conn, logger.New(os.Stdout))

	assert.Nil(t, err, "Actual err: %v", err)
}

func TestIT_HttpListenAndServer_WhenDbConnectionWorks_ExpectHealthcheckSucceeds(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())

	dbConn := newTestDbConnection(t)
	defer dbConn.Close(context.Background())

	conf := newHttpTestConfig(7201)

	wg := asyncHttpListenAndServe(t, conf, dbConn, cancellable)

	client := &http.Client{}
	url := "http://localhost:7201/v1/chats/healthcheck"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	assert.Nil(t, err, "Actual err: %v", err)

	rw, err := client.Do(req)
	assert.Nil(t, err, "Actual err: %v", err)

	cancel()
	wg.Wait()

	assert.Equal(t, http.StatusOK, rw.StatusCode)
	assertResponseContainsDetails(t, rw, success, `"OK"`)
}

type mockDbConn struct {
	db.Connection
}

func (m *mockDbConn) Ping(ctx context.Context) error {
	return errSample
}

func TestUnit_HttpListenAndServer_WhenDbConnectionFails_ExpectHealthcheckFails(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())

	conf := newHttpTestConfig(7202)

	wg := asyncHttpListenAndServe(t, conf, &mockDbConn{}, cancellable)

	client := &http.Client{}
	url := "http://localhost:7202/v1/chats/healthcheck"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	assert.Nil(t, err, "Actual err: %v", err)

	rw, err := client.Do(req)
	assert.Nil(t, err, "Actual err: %v", err)

	cancel()
	wg.Wait()

	assert.Equal(t, http.StatusServiceUnavailable, rw.StatusCode)
	assertResponseContainsDetails(t, rw, failure, "{}")
}

func newHttpTestConfig(port uint16) Configuration {
	conf := DefaultConfig()
	conf.Server.Port = port
	return conf
}

func asyncHttpListenAndServe(
	t *testing.T,
	config Configuration,
	conn db.Connection,
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
		err = httpListenAndServe(ctx, config, conn, logger.New(os.Stdout))
		assert.Nil(t, err, "Actual err: %v", err)
	}()

	time.Sleep(reasonableWaitTimeForServerToBeUp)

	return &wg
}

func assertResponseContainsDetails(t *testing.T, rw *http.Response, status string, expectedContent string) {
	body, err := io.ReadAll(rw.Body)
	assert.Nil(t, err, "Actual err: %v", err)

	const uuidRegexTemplate = `[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`
	const requestBodyRegexTemplate = `{"requestId":"%s","status":"%s","details":%s}`

	regexText := fmt.Sprintf(requestBodyRegexTemplate, uuidRegexTemplate, status, expectedContent)
	expected := regexp.MustCompile(regexText)

	assert.Regexp(t, expected, string(body), "Actual response body: %v", string(body))
}
