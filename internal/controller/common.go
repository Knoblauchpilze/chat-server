package controller

import (
	"github.com/labstack/echo/v4"
)

type componentAwareHttpHandler[T any] func(echo.Context, T) error

func createComponentAwareHttpHandler[T any](handler componentAwareHttpHandler[T], component T) echo.HandlerFunc {
	return func(c echo.Context) error {
		return handler(c, component)
	}
}
