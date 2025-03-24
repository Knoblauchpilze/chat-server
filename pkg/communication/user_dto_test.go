package communication

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUnit_UserDtoRequest_MarshalsToCamelCase(t *testing.T) {
	dto := UserDtoRequest{
		Name:    "my-user",
		ApiUser: uuid.MustParse("7fe0e1a3-4e8d-4a1d-ab4e-1e617f2c5666"),
	}

	out, err := json.Marshal(dto)

	assert.Nil(t, err)
	expectedJson := `
	{
		"name": "my-user",
		"api_user": "7fe0e1a3-4e8d-4a1d-ab4e-1e617f2c5666"
	}`
	assert.JSONEq(t, expectedJson, string(out))
}

func TestUnit_FromUserDtoRequest(t *testing.T) {
	beforeConversion := time.Now()

	dto := UserDtoRequest{
		Name:    "my-user",
		ApiUser: uuid.New(),
	}

	actual := FromUserDtoRequest(dto)

	assert.Equal(t, "my-user", actual.Name)
	assert.Equal(t, dto.ApiUser, actual.ApiUser)
	assert.True(t, actual.CreatedAt.After(beforeConversion))
	assert.Equal(t, actual.CreatedAt, actual.UpdatedAt)
}

func TestUnit_UserDtoResponse_MarshalsToCamelCase(t *testing.T) {
	dto := UserDtoResponse{
		Id:        uuid.MustParse("a590b448-d3cd-4dbc-a9e3-8d642b1a5814"),
		Name:      "my-user",
		ApiUser:   uuid.MustParse("3038a794-bbb6-4b7b-bd87-009baf08d211"),
		CreatedAt: someTime,
	}

	out, err := json.Marshal(dto)

	assert.Nil(t, err)
	expectedJson := `
	{
		"id": "a590b448-d3cd-4dbc-a9e3-8d642b1a5814",
		"name": "my-user",
		"api_user": "3038a794-bbb6-4b7b-bd87-009baf08d211",
		"created_at": "2024-11-12T19:09:36Z"
	}`
	assert.JSONEq(t, expectedJson, string(out))
}

func TestUnit_ToUserDtoResponse(t *testing.T) {
	entity := persistence.User{
		Id:      uuid.New(),
		Name:    "my-user",
		ApiUser: uuid.New(),

		CreatedAt: someTime,
	}

	actual := ToUserDtoResponse(entity)

	assert.Equal(t, entity.Id, actual.Id)
	assert.Equal(t, "my-user", actual.Name)
	assert.Equal(t, entity.ApiUser, actual.ApiUser)
	assert.Equal(t, someTime, actual.CreatedAt)
}
