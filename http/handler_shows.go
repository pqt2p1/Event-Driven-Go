package http

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
	"tickets/entities"
	"time"
)

type showRequest struct {
	DeadNationID   string    `json:"dead_nation_id"`
	NumberOfTicket int       `json:"number_of_ticket"`
	StartTime      time.Time `json:"start_time"`
	Title          string    `json:"title"`
	Venue          string    `json:"venue"`
}

func (h Handler) PostShows(c echo.Context) error {
	var request showRequest
	err := c.Bind(&request)
	if err != nil {
		return err
	}

	showID := uuid.New().String()

	show := entities.Show{
		ShowID:          showID,
		DeadNationID:    request.DeadNationID,
		NumberOfTickets: request.NumberOfTicket,
		StartTime:       request.StartTime,
		Title:           request.Title,
		Venue:           request.Venue,
	}

	err = h.showsRepo.Add(c.Request().Context(), show)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"show_id": showID,
	})

}
