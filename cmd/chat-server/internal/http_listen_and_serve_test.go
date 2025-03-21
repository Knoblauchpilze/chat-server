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
	"github.com/Knoblauchpilze/chat-server/internal/service"
	"github.com/Knoblauchpilze/chat-server/pkg/repositories"
	"github.com/stretchr/testify/assert"
)

const (
	success = "SUCCESS"
	failure = "ERROR"
)

func TestIT_HttpListenAndServe_StartAndStopWithContext(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())
	asyncCancelContext(200*time.Millisecond, cancel)

	dbConn := newTestDbConnection(t)
	defer dbConn.Close(context.Background())

	props := newTestHttpProps(7200, dbConn)

	err := httpListenAndServe(cancellable, props)

	assert.Nil(t, err, "Actual err: %v", err)
}

func TestIT_HttpListenAndServe_WhenDbConnectionWorks_ExpectHealthcheckSucceeds(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())

	dbConn := newTestDbConnection(t)
	defer dbConn.Close(context.Background())

	props := newTestHttpProps(7201, dbConn)

	wg := asyncHttpListenAndServe(t, props, cancellable)

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

func TestUnit_HttpListenAndServe_WhenDbConnectionFails_ExpectHealthcheckFails(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())

	props := newTestHttpProps(7202, &mockDbConn{})

	wg := asyncHttpListenAndServe(t, props, cancellable)

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

func newTestHttpConfig(port uint16) Configuration {
	conf := DefaultConfig()
	conf.Server.Port = port
	return conf
}

func newTestHttpProps(port uint16, dbConn db.Connection) HttpServerProps {
	log := logger.New(os.Stdout)
	repos := repositories.New(dbConn)

	return HttpServerProps{
		Config:   newTestHttpConfig(port),
		DbConn:   dbConn,
		Services: service.New(dbConn, repos, log),
		Log:      log,
	}
}

func asyncHttpListenAndServe(
	t *testing.T,
	props HttpServerProps,
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
		err = httpListenAndServe(ctx, props)
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
