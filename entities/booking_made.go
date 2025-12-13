package entities

type BookingMade struct {
	Header MessageHeader `json:"header"`

	NumberOfTickets int    `json:"number_of_tickets"`
	BookingID       string `json:"booking_id"`
	CustomerEmail   string `json:"customer_email"`
	ShowID          string `json:"show_id"`
}
