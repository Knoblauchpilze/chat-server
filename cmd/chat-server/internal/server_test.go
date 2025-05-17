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

	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/Knoblauchpilze/chat-server/pkg/communication"
	eassert "github.com/Knoblauchpilze/easy-assert/assert"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestIT_RunServer_Room_CreateGetDeleteWorkflow(t *testing.T) {
	props := newTestServerConfig(7601)
	cancellable, cancel := context.WithCancel(context.Background())

	wg := asyncRunServerAndAssertNoError(t, cancellable, props)

	// Create a new room
	requestDto := communication.RoomDtoRequest{
		Name: fmt.Sprintf("my-room-%v", uuid.New()),
	}

	url := "http://localhost:7601/v1/chats/rooms"
	rw := doRequestWithData(t, http.MethodPost, url, requestDto)

	responseDto := assertResponseAndExtractDetails[communication.RoomDtoResponse](
		t, rw, success,
	)

	assert.Equal(t, http.StatusCreated, rw.StatusCode)
	assert.Equal(t, requestDto.Name, responseDto.Name)

	// Fetch it
	url = fmt.Sprintf("http://localhost:7601/v1/chats/rooms/%s", responseDto.Id)
	rw = doRequest(t, http.MethodGet, url)

	getResponseDto := assertResponseAndExtractDetails[communication.RoomDtoResponse](
		t, rw, success,
	)

	assert.Equal(t, http.StatusOK, rw.StatusCode)
	assert.Equal(t, responseDto, getResponseDto)

	// Delete it
	url = fmt.Sprintf("http://localhost:7601/v1/chats/rooms/%s", responseDto.Id)
	rw = doRequest(t, http.MethodDelete, url)

	assert.Equal(t, http.StatusNoContent, rw.StatusCode)

	// Get it again
	url = fmt.Sprintf("http://localhost:7601/v1/chats/rooms/%s", responseDto.Id)
	rw = doRequest(t, http.MethodGet, url)

	assert.Equal(t, http.StatusNotFound, rw.StatusCode)

	cancel()
	wg.Wait()
}

func TestIT_RunServer_User_CreateGetDeleteWorkflow(t *testing.T) {
	props := newTestServerConfig(7602)
	cancellable, cancel := context.WithCancel(context.Background())

	wg := asyncRunServerAndAssertNoError(t, cancellable, props)

	// Create a new user
	requestDto := communication.UserDtoRequest{
		Name: fmt.Sprintf("my-user-%v", uuid.New()),
	}

	url := "http://localhost:7602/v1/chats/users"
	rw := doRequestWithData(t, http.MethodPost, url, requestDto)

	responseDto := assertResponseAndExtractDetails[communication.UserDtoResponse](
		t, rw, success,
	)

	assert.Equal(t, http.StatusCreated, rw.StatusCode)
	assert.Equal(t, requestDto.Name, responseDto.Name)

	// Fetch it
	url = fmt.Sprintf("http://localhost:7602/v1/chats/users/%s", responseDto.Id)
	rw = doRequest(t, http.MethodGet, url)

	getResponseDto := assertResponseAndExtractDetails[communication.UserDtoResponse](
		t, rw, success,
	)

	assert.Equal(t, http.StatusOK, rw.StatusCode)
	assert.Equal(t, responseDto, getResponseDto)

	// Delete it
	url = fmt.Sprintf("http://localhost:7602/v1/chats/users/%s", responseDto.Id)
	rw = doRequest(t, http.MethodDelete, url)

	assert.Equal(t, http.StatusNoContent, rw.StatusCode)

	// Get it again
	url = fmt.Sprintf("http://localhost:7602/v1/chats/users/%s", responseDto.Id)
	rw = doRequest(t, http.MethodGet, url)

	assert.Equal(t, http.StatusNotFound, rw.StatusCode)

	cancel()
	wg.Wait()
}

func TestIT_RunServer_ListForRoom(t *testing.T) {
	props := newTestServerConfig(7603)
	cancellable, cancel := context.WithCancel(context.Background())
	dbConn := newTestDbConnection(t)
	defer dbConn.Close(context.Background())

	user := insertTestUser(t, dbConn)
	room := insertTestRoom(t, dbConn)
	registerUserInRoom(t, dbConn, user.Id, room.Id)

	wg := asyncRunServerAndAssertNoError(t, cancellable, props)

	url := fmt.Sprintf("http://localhost:7603/v1/chats/rooms/%s/users", room.Id)
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

func TestIT_RunServer_Message_PostGetWorkflow(t *testing.T) {
	props := newTestServerConfig(7604)
	cancellable, cancel := context.WithCancel(context.Background())
	dbConn := newTestDbConnection(t)
	defer dbConn.Close(context.Background())

	user := insertTestUser(t, dbConn)
	room := insertTestRoom(t, dbConn)
	registerUserInRoom(t, dbConn, user.Id, room.Id)

	wg := asyncRunServerAndAssertNoError(t, cancellable, props)

	// Create a new message
	requestDto := communication.MessageDtoRequest{
		User:    user.Id,
		Room:    room.Id,
		Message: fmt.Sprintf("%s says hello to %s", user.Name, room.Id),
	}

	url := fmt.Sprintf("http://localhost:7604/v1/chats/rooms/%s/messages", room.Id)
	rw := doRequestWithData(t, http.MethodPost, url, requestDto)

	assert.Equal(t, http.StatusAccepted, rw.StatusCode)

	// Wait a bit for the processor to persist the message
	time.Sleep(50 * time.Millisecond)

	// Fetch it
	url = fmt.Sprintf("http://localhost:7604/v1/chats/rooms/%s/messages", room.Id)
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
	wg.Wait()
}

func TestIT_RunServer_SubscribeToMessage_ReceivesPostedMessage(t *testing.T) {
	props := newTestServerConfig(7605)
	cancellable, cancelServer := context.WithCancel(context.Background())
	dbConn := newTestDbConnection(t)
	defer dbConn.Close(context.Background())

	user1 := insertTestUser(t, dbConn)
	user2 := insertTestUser(t, dbConn)
	room := insertTestRoom(t, dbConn)
	registerUserInRoom(t, dbConn, user1.Id, room.Id)
	registerUserInRoom(t, dbConn, user2.Id, room.Id)

	wg := asyncRunServerAndAssertNoError(t, cancellable, props)

	requestDto := communication.MessageDtoRequest{
		User:    user2.Id,
		Room:    room.Id,
		Message: fmt.Sprintf("%s says hello to %s", user2.Name, room.Id),
	}

	// Post a message with a delay to let the user the time to subscribe
	go func() {
		time.Sleep(100 * time.Millisecond)

		url := fmt.Sprintf("http://localhost:7605/v1/chats/rooms/%s/messages", room.Id)
		rw := doRequestWithData(t, http.MethodPost, url, requestDto)

		assert.Equal(t, http.StatusAccepted, rw.StatusCode)
	}()

	// Give the message a bit of time to be processed
	reqCtx, cancelReq := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancelReq()

	url := fmt.Sprintf("http://localhost:7605/v1/chats/users/%s/subscribe", user1.Id)
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
	assert.Nil(t, err, "Actual err: %v", err)

	client := &http.Client{}
	rw, err := client.Do(req)
	assert.Nil(t, err, "Actual err: %v", err)

	// Wait a bit for the message to be processed and sent to the client
	time.Sleep(200 * time.Millisecond)

	cancelServer()
	wg.Wait()

	assert.Equal(t, http.StatusOK, rw.StatusCode)

	// We don't know the id of the message so we need to use a regex
	const uuidPattern = "[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}"
	const timePatten = "[0-9-]+T[0-9:.]+Z"
	expectedBodyRegex := fmt.Sprintf(
		`id: %s
data: {"id":"%s","user":"%s","room":"%s","message":"%s","created_at":"%s"}

`,
		uuidPattern,
		uuidPattern,
		user2.Id.String(),
		room.Id.String(),
		requestDto.Message,
		timePatten,
	)
	body, err := io.ReadAll(rw.Body)
	assert.Nil(t, err, "Actual err: %v", err)
	actual := string(body)
	assert.Regexp(
		t,
		regexp.MustCompile(expectedBodyRegex),
		actual,
		"Actual body: %s",
		actual,
	)
}

func TestIT_RunServer_SubscribeToMessage_DoesNotReceiveMessageFromAnotherRoom(t *testing.T) {
	props := newTestServerConfig(7606)
	cancellable, cancelServer := context.WithCancel(context.Background())
	dbConn := newTestDbConnection(t)
	defer dbConn.Close(context.Background())

	user1 := insertTestUser(t, dbConn)
	user2 := insertTestUser(t, dbConn)
	room := insertTestRoom(t, dbConn)
	registerUserInRoom(t, dbConn, user2.Id, room.Id)

	wg := asyncRunServerAndAssertNoError(t, cancellable, props)

	requestDto := communication.MessageDtoRequest{
		User:    user2.Id,
		Room:    room.Id,
		Message: fmt.Sprintf("%s says hello to %s", user2.Name, room.Id),
	}

	// Post a message with a delay to let the user the time to subscribe
	go func() {
		time.Sleep(100 * time.Millisecond)

		url := fmt.Sprintf("http://localhost:7606/v1/chats/rooms/%s/messages", room.Id)
		rw := doRequestWithData(t, http.MethodPost, url, requestDto)

		assert.Equal(t, http.StatusAccepted, rw.StatusCode)
	}()

	// Give the message a bit of time to be processed
	reqCtx, cancelReq := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancelReq()

	url := fmt.Sprintf("http://localhost:7606/v1/chats/users/%s/subscribe", user1.Id)
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
	assert.Nil(t, err, "Actual err: %v", err)

	client := &http.Client{}
	rw, err := client.Do(req)
	assert.Nil(t, err, "Actual err: %v", err)

	// Wait a bit for the message to be processed and sent to the client
	time.Sleep(200 * time.Millisecond)

	cancelServer()
	wg.Wait()

	assert.Equal(t, http.StatusOK, rw.StatusCode)

	_, err = io.ReadAll(rw.Body)
	assert.Equal(t, io.ErrUnexpectedEOF, err, "Actual err: %v", err)
}

func newTestServerConfig(httpPort uint16) Configuration {
	baseConfig := DefaultConfig()
	baseConfig.Server.Port = httpPort

	return Configuration{
		Server:   baseConfig.Server,
		Database: dbTestConfig,
	}
}

func asyncRunServerAndAssertNoError(
	t *testing.T,
	ctx context.Context,
	config Configuration,
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
		err = RunServer(ctx, config, logger.New(os.Stdout))
		assert.Nil(t, err, "Actual err: %v", err)
	}()

	time.Sleep(reasonableWaitTimeForServerToBeUp)

	return &wg
}
