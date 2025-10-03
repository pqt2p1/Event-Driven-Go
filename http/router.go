package http

import (
	libHttp "github.com/ThreeDotsLabs/go-event-driven/v2/common/http"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/labstack/echo/v4"
	"net/http"
)

func NewHttpRouter(publisher message.Publisher, spreadsheetsAPIClient SpreadsheetsAPI) *echo.Echo {
	e := libHttp.NewEcho()

	handler := Handler{
		publisher:             publisher,
		spreadsheetsAPIClient: spreadsheetsAPIClient,
	}

	e.POST("/tickets-status", handler.PostTicketsStatus)
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	return e
}
