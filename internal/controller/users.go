package controller

import (
	"net/http"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/Knoblauchpilze/chat-server/internal/service"
	"github.com/Knoblauchpilze/chat-server/pkg/communication"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func UserEndpoints(service service.UserService) Routes {
	var out Routes

	postHandler := createComponentAwareHttpHandler(createUser, service)
	post := rest.NewRoute(http.MethodPost, "/users", postHandler)
	out = append(out, post)

	getHandler := createComponentAwareHttpHandler(getUser, service)
	get := rest.NewRoute(http.MethodGet, "/users/:id", getHandler)
	out = append(out, get)

	listHandler := createComponentAwareHttpHandler(listUsers, service)
	list := rest.NewRoute(http.MethodGet, "/users", listHandler)
	out = append(out, list)

	listForUserHandler := createComponentAwareHttpHandler(listForUser, service)
	listForUser := rest.NewRoute(http.MethodGet, "/users/:id/rooms", listForUserHandler)
	out = append(out, listForUser)

	deleteHandler := createComponentAwareHttpHandler(deleteUser, service)
	delete := rest.NewRoute(http.MethodDelete, "/users/:id", deleteHandler)
	out = append(out, delete)

	return out
}

func createUser(c *gin.Context, s service.UserService) {
	var userDtoRequest communication.UserDtoRequest
	err := c.ShouldBindJSON(&userDtoRequest)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, "Invalid user syntax")
		return
	}

	out, err := s.Create(c.Request.Context(), userDtoRequest)
	if err != nil {
		if codedErr, ok := errors.AsErrorWithCode(err); ok && codedErr.Code == service.ErrInvalidName {
			c.AbortWithStatusJSON(http.StatusBadRequest, "Invalid user name")
			return
		}
		if dbErr, ok := db.AsDatabaseError(err); ok && dbErr.Code == db.ErrUniqueConstraintViolation {
			c.AbortWithStatusJSON(http.StatusConflict, "User name already in use")
			return
		}

		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, out)
}

func getUser(c *gin.Context, s service.UserService) {
	maybeId := c.Param("id")
	id, err := uuid.Parse(maybeId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, "Invalid id syntax")
		return
	}

	out, err := s.Get(c.Request.Context(), id)
	if err != nil {
		if err == db.ErrNoMatchingRows {
			c.AbortWithStatusJSON(http.StatusNotFound, "No such user")
			return
		}

		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, out)
}

func listUsers(c *gin.Context, s service.UserService) {
	const userNameKey = "name"
	maybeName := c.Query(userNameKey)
	exists := (maybeName != "")

	var users []communication.UserDtoResponse
	var err error

	if exists {
		var user communication.UserDtoResponse
		user, err = s.GetByName(c.Request.Context(), maybeName)
		if err == nil {
			users = append(users, user)
		}
	} else {
		c.JSON(
			http.StatusBadRequest,
			"Please provide a user name as filtering parameter",
		)
		return
	}

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	out, err := marshalNilToEmptySlice(users)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	c.Data(http.StatusOK, "application/json", out)
}

func listForUser(c *gin.Context, s service.UserService) {
	maybeId := c.Param("id")
	id, err := uuid.Parse(maybeId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, "Invalid id syntax")
		return
	}

	rooms, err := s.ListForUser(c.Request.Context(), id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	out, err := marshalNilToEmptySlice(rooms)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	c.Data(http.StatusOK, "application/json", out)
}

func deleteUser(c *gin.Context, s service.UserService) {
	maybeId := c.Param("id")
	id, err := uuid.Parse(maybeId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, "Invalid id syntax")
		return
	}

	err = s.Delete(c.Request.Context(), id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusNoContent)
}
