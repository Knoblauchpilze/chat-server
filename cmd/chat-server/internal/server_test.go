package internal

import (
	"context"
	"fmt"
	"net/http"
	"os"
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
	insertUserInRoom(t, dbConn, user.Id, room.Id)

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
	insertUserInRoom(t, dbConn, user.Id, room.Id)

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
