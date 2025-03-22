package communication

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var someTime = time.Date(2024, 11, 12, 19, 9, 36, 0, time.UTC)

func TestUnit_RoomDtoRequest_MarshalsToCamelCase(t *testing.T) {
	dto := RoomDtoRequest{
		Name: "my-room",
	}

	out, err := json.Marshal(dto)

	assert.Nil(t, err)
	expectedJson := `
	{
		"name": "my-room"
	}`
	assert.JSONEq(t, expectedJson, string(out))
}

func TestUnit_FromRoomDtoRequest(t *testing.T) {
	beforeConversion := time.Now()

	dto := RoomDtoRequest{
		Name: "my-room",
	}

	actual := FromRoomDtoRequest(dto)

	assert.Equal(t, "my-room", actual.Name)
	assert.True(t, actual.CreatedAt.After(beforeConversion))
	assert.Equal(t, actual.CreatedAt, actual.UpdatedAt)
}

func TestUnit_RoomDtoResponse_MarshalsToCamelCase(t *testing.T) {
	dto := RoomDtoResponse{
		Id:        uuid.MustParse("a590b448-d3cd-4dbc-a9e3-8d642b1a5814"),
		Name:      "my-room",
		CreatedAt: someTime,
	}

	out, err := json.Marshal(dto)

	assert.Nil(t, err)
	expectedJson := `
	{
		"id": "a590b448-d3cd-4dbc-a9e3-8d642b1a5814",
		"name": "my-room",
		"createdAt": "2024-11-12T19:09:36Z"
	}`
	assert.JSONEq(t, expectedJson, string(out))
}

func TestUnit_ToRoomDtoResponse(t *testing.T) {
	entity := persistence.Room{
		Id:   uuid.New(),
		Name: "my-room",

		CreatedAt: someTime,
	}

	actual := ToRoomDtoResponse(entity)

	assert.Equal(t, entity.Id, actual.Id)
	assert.Equal(t, "my-room", actual.Name)
	assert.Equal(t, someTime, actual.CreatedAt)
}
