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
	"github.com/Knoblauchpilze/chat-server/pkg/clients"
	"github.com/Knoblauchpilze/chat-server/pkg/communication"
	"github.com/Knoblauchpilze/chat-server/pkg/messages"
	"github.com/Knoblauchpilze/chat-server/pkg/repositories"
	eassert "github.com/Knoblauchpilze/easy-assert/assert"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const reasonableConnectTimeout = 100 * time.Millisecond

func TestIT_RunHttpServer_StartAndStopWithContext(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())
	asyncCancelContext(200*time.Millisecond, cancel)

	dbConn := newTestDbConnection(t)
	defer dbConn.Close(context.Background())

	props, _ := newTestHttpProps(7200, dbConn)

	err := RunHttpServer(cancellable, props)

	assert.Nil(t, err, "Actual err: %v", err)
}

func TestIT_RunHttpServer_WhenDbConnectionWorks_ExpectHealthcheckSucceeds(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())

	dbConn := newTestDbConnection(t)
	defer dbConn.Close(context.Background())

	props, _ := newTestHttpProps(7201, dbConn)

	wg := asyncRunHttpServer(t, props, cancellable)

	url := "http://localhost:7201/v1/chats/healthcheck"
	rw := doRequest(t, http.MethodGet, url)

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

func TestUnit_RunHttpServer_WhenDbConnectionFails_ExpectHealthcheckFails(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())

	props, _ := newTestHttpProps(7202, &mockDbConn{})

	wg := asyncRunHttpServer(t, props, cancellable)

	url := "http://localhost:7202/v1/chats/healthcheck"
	rw := doRequest(t, http.MethodGet, url)

	cancel()
	wg.Wait()

	assert.Equal(t, http.StatusServiceUnavailable, rw.StatusCode)
	assertResponseContainsDetails(t, rw, failure, "{}")
}

func TestIT_RunHttpServer_Room_CreateGetDeleteWorkflow(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())
	dbConn := newTestDbConnection(t)
	defer dbConn.Close(context.Background())
	props, _ := newTestHttpProps(7203, dbConn)

	wg := asyncRunHttpServer(t, props, cancellable)

	// Create a new room
	requestDto := communication.RoomDtoRequest{
		Name: fmt.Sprintf("my-room-%v", uuid.New()),
	}

	url := "http://localhost:7203/v1/chats/rooms"
	rw := doRequestWithData(t, http.MethodPost, url, requestDto)

	responseDto := assertResponseAndExtractDetails[communication.RoomDtoResponse](
		t, rw, success,
	)

	assert.Equal(t, http.StatusCreated, rw.StatusCode)
	assert.Equal(t, requestDto.Name, responseDto.Name)

	// Fetch it
	url = fmt.Sprintf("http://localhost:7203/v1/chats/rooms/%s", responseDto.Id)
	rw = doRequest(t, http.MethodGet, url)

	getResponseDto := assertResponseAndExtractDetails[communication.RoomDtoResponse](
		t, rw, success,
	)

	assert.Equal(t, http.StatusOK, rw.StatusCode)
	assert.Equal(t, responseDto, getResponseDto)

	// Delete it
	url = fmt.Sprintf("http://localhost:7203/v1/chats/rooms/%s", responseDto.Id)
	rw = doRequest(t, http.MethodDelete, url)

	assert.Equal(t, http.StatusNoContent, rw.StatusCode)

	// Get it again
	url = fmt.Sprintf("http://localhost:7203/v1/chats/rooms/%s", responseDto.Id)
	rw = doRequest(t, http.MethodGet, url)

	assert.Equal(t, http.StatusNotFound, rw.StatusCode)

	cancel()
	wg.Wait()
}

func TestIT_RunHttpServer_User_CreateGetDeleteWorkflow(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())
	dbConn := newTestDbConnection(t)
	defer dbConn.Close(context.Background())
	props, _ := newTestHttpProps(7203, dbConn)

	wg := asyncRunHttpServer(t, props, cancellable)

	// Create a new user
	requestDto := communication.UserDtoRequest{
		Name: fmt.Sprintf("my-user-%v", uuid.New()),
	}

	url := "http://localhost:7203/v1/chats/users"
	rw := doRequestWithData(t, http.MethodPost, url, requestDto)

	responseDto := assertResponseAndExtractDetails[communication.UserDtoResponse](
		t, rw, success,
	)

	assert.Equal(t, http.StatusCreated, rw.StatusCode)
	assert.Equal(t, requestDto.Name, responseDto.Name)

	// Fetch it
	url = fmt.Sprintf("http://localhost:7203/v1/chats/users/%s", responseDto.Id)
	rw = doRequest(t, http.MethodGet, url)

	getResponseDto := assertResponseAndExtractDetails[communication.UserDtoResponse](
		t, rw, success,
	)

	assert.Equal(t, http.StatusOK, rw.StatusCode)
	assert.Equal(t, responseDto, getResponseDto)

	// Delete it
	url = fmt.Sprintf("http://localhost:7203/v1/chats/users/%s", responseDto.Id)
	rw = doRequest(t, http.MethodDelete, url)

	assert.Equal(t, http.StatusNoContent, rw.StatusCode)

	// Get it again
	url = fmt.Sprintf("http://localhost:7203/v1/chats/users/%s", responseDto.Id)
	rw = doRequest(t, http.MethodGet, url)

	assert.Equal(t, http.StatusNotFound, rw.StatusCode)

	cancel()
	wg.Wait()
}

func TestIT_RunHttpServer_ListForRoom(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())
	dbConn := newTestDbConnection(t)
	defer dbConn.Close(context.Background())
	props, _ := newTestHttpProps(7203, dbConn)

	user := insertTestUser(t, dbConn)
	room := insertTestRoom(t, dbConn)
	insertUserInRoom(t, dbConn, user.Id, room.Id)

	wg := asyncRunHttpServer(t, props, cancellable)

	url := fmt.Sprintf("http://localhost:7203/v1/chats/rooms/%s/users", room.Id)
	rw := doRequest(t, http.MethodGet, url)

	responseDto := assertResponseAndExtractDetails[[]communication.UserDtoResponse](
		t, rw, success,
	)

	assert.Equal(t, http.StatusOK, rw.StatusCode)
	assert.Len(t, responseDto, 1)
	expected := communication.ToUserDtoResponse(user)
	assert.True(
		t,
		eassert.EqualsIgnoringFields(expected, responseDto[0]),
		"Expected: %v, actual: %v",
		expected,
		responseDto,
	)

	cancel()
	wg.Wait()
}

func TestIT_RunHttpServer_Message_PostGetWorkflow(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())
	dbConn := newTestDbConnection(t)
	defer dbConn.Close(context.Background())
	props, processor := newTestHttpProps(7204, dbConn)

	user := insertTestUser(t, dbConn)
	room := insertTestRoom(t, dbConn)
	insertUserInRoom(t, dbConn, user.Id, room.Id)

	wg := asyncRunHttpServer(t, props, cancellable)
	wgProcessor := asyncRunMessageProcessor(t, processor)

	// Create a new message
	requestDto := communication.MessageDtoRequest{
		User:    user.Id,
		Room:    room.Id,
		Message: fmt.Sprintf("%s says hello to %s", user.Name, room.Id),
	}

	url := fmt.Sprintf("http://localhost:7204/v1/chats/rooms/%s/messages", room.Id)
	rw := doRequestWithData(t, http.MethodPost, url, requestDto)

	assert.Equal(t, http.StatusAccepted, rw.StatusCode)

	// Wait a bit for the processor to persist the message
	time.Sleep(50 * time.Millisecond)

	// Fetch it
	url = fmt.Sprintf("http://localhost:7204/v1/chats/rooms/%s/messages", room.Id)
	rw = doRequest(t, http.MethodGet, url)

	getResponseDto := assertResponseAndExtractDetails[[]communication.MessageDtoResponse](
		t, rw, success,
	)

	assert.Equal(t, http.StatusOK, rw.StatusCode)
	assert.Equal(t, 1, len(getResponseDto))
	assert.Len(t, getResponseDto, 1)
	actual := getResponseDto[0]
	assert.Equal(t, requestDto.User, actual.User)
	assert.Equal(t, requestDto.Room, actual.Room)
	assert.Equal(t, requestDto.Message, actual.Message)

	cancel()
	err := processor.Stop()
	wg.Wait()
	wgProcessor.Wait()

	assert.Nil(t, err, "Actual err: %v", err)
}

func newTestHttpConfig(port uint16) Configuration {
	conf := DefaultConfig()
	conf.Server.Port = port
	conf.ConnectTimeout = reasonableConnectTimeout
	return conf
}

func newTestHttpProps(
	port uint16, dbConn db.Connection,
) (HttpServerProps, messages.Processor) {
	conf := newTestHttpConfig(port)
	log := logger.New(os.Stdout)
	repos := repositories.New(dbConn)

	clientManager := clients.NewManager()
	processor := messages.NewMessageProcessor(1, clientManager, repos)

	props := HttpServerProps{
		Config: conf,
		DbConn: dbConn,
		Services: service.New(
			conf.ConnectTimeout,
			dbConn,
			repos,
			processor,
			clientManager,
			log,
		),
		Log: log,
	}

	return props, processor
}

func asyncRunHttpServer(
	t *testing.T,
	props HttpServerProps,
	ctx context.Context,
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
		err := RunHttpServer(ctx, props)
		assert.Nil(t, err, "Actual err: %v", err)
	}()

	time.Sleep(reasonableWaitTimeForServerToBeUp)

	return &wg
}

func asyncRunMessageProcessor(
	t *testing.T, processor messages.Processor,
) *sync.WaitGroup {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if panicErr := recover(); panicErr != nil {
				assert.Failf(t, "Processor panicked", "Panic details: %v", panicErr)
			}
		}()
		err := processor.Start()
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
