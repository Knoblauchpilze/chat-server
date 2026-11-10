package controller

import (
	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/gin-gonic/gin"
)

type Routes []*rest.Route

type componentAwareHttpHandler[T any] func(*gin.Context, T)

func createComponentAwareHttpHandler[T any](handler componentAwareHttpHandler[T], component T) gin.HandlerFunc {
	return func(c *gin.Context) {
		handler(c, component)
	}
}
