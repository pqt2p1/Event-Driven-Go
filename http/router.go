package http

import (
	libHttp "github.com/ThreeDotsLabs/go-event-driven/v2/common/http"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/labstack/echo/v4"
	"net/http"
)

func NewHttpRouter(eventBus *cqrs.EventBus, commandBus *cqrs.CommandBus, ticketsRepo TicketsRepository, showsRepo ShowsRepository, bookingRepo BookingRepository) *echo.Echo {
	e := libHttp.NewEcho()

	handler := Handler{
		eventBus:    eventBus,
		commandBus:  commandBus,
		ticketsRepo: ticketsRepo,
		showsRepo:   showsRepo,
		bookingRepo: bookingRepo,
	}

	e.POST("/tickets-status", handler.PostTicketsStatus)
	e.GET("/tickets", handler.GetTickets)
	e.POST("/shows", handler.PostShows)
	e.POST("/book-tickets", handler.PostBookTickets)
	e.PUT("/ticket-refund/:ticket_id", handler.PutTicketRefund)

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	return e
}
