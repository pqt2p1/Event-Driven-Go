package entities

type TicketPrinted struct {
	Header   MessageHeader `json:"header"`
	TicketID string        `json:"ticket_id"`
	FileName string        `json:"file_name"`
}
