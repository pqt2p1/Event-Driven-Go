package tests_test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lithammer/shortuuid/v3"
	"net/http"
	"os"
	"testing"
	"tickets/adapters"
	"tickets/entities"
	ticketsHttp "tickets/http"
	"tickets/message"
	"tickets/service"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComponent(t *testing.T) {
	redisClient := message.NewRedisClient("REDIS_ADDR")

	defer redisClient.Close()

	db, err := sqlx.Open("postgres", os.Getenv("POSTGRES_URL"))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	spreadsheetsAPI := &adapters.SpreadsheetsAPIStub{}
	receiptsService := &adapters.ReceiptsServiceStub{}
	filesAPI := &adapters.FilesAPIStub{}
	deadNationClient := &adapters.DeadNationStub{}

	go func() {
		svc := service.New(
			db,
			redisClient,
			spreadsheetsAPI,
			receiptsService,
			filesAPI,
			deadNationClient,
		)
		err := svc.Run(ctx)
		assert.NoError(t, err)
	}()
	waitForHttpServer(t)

	ticket := ticketsHttp.TicketStatus{
		TicketID:      "test-ticket",
		Status:        "confirmed",
		CustomerEmail: "test@example.com",
		Price: entities.Money{
			Amount:   "100",
			Currency: "USD",
		},
	}

	ticketCancel := ticketsHttp.TicketStatus{
		TicketID:      "test-ticket",
		Status:        "canceled",
		CustomerEmail: "test@example.com",
		Price: entities.Money{
			Amount:   "100",
			Currency: "USD",
		},
	}

	request := ticketsHttp.TicketsStatusRequest{
		Tickets: []ticketsHttp.TicketStatus{ticket, ticketCancel},
	}
	sendTicketsStatus(t, request)
	assertReceiptForTicketIssued(t, receiptsService, ticket)
	assertRowInTicketsToPrint(t, spreadsheetsAPI, ticket)
	assertRowInticketsToRefund(t, spreadsheetsAPI, ticketCancel)
	assertTicketPrinted(t, filesAPI, ticket)

	showID := createShow(t, db)
	bookingID := sendBookTicketsRequest(t, showID, "customer@example.com", 3)
	assertBookingCreated(t, db, bookingID, showID, "customer@example.com", 3)
	assertBookingInDeadNation(t, deadNationClient, bookingID, "customer@example.com", 3)

}

func sendTicketsStatus(t *testing.T, req ticketsHttp.TicketsStatusRequest) {
	t.Helper()

	payload, err := json.Marshal(req)
	require.NoError(t, err)

	correlationID := shortuuid.New()

	httpReq, err := http.NewRequest(
		http.MethodPost,
		"http://localhost:8080/tickets-status",
		bytes.NewBuffer(payload),
	)
	require.NoError(t, err)

	httpReq.Header.Set("Correlation-ID", correlationID)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func assertReceiptForTicketIssued(t *testing.T, receiptsService *adapters.ReceiptsServiceStub, ticket ticketsHttp.TicketStatus) {
	t.Helper()

	parentT := t

	assert.EventuallyWithT(
		t,
		func(t *assert.CollectT) {
			issuedReceipts := len(receiptsService.IssuedReceipts)
			parentT.Log("issued receipts", issuedReceipts)

			assert.Greater(t, issuedReceipts, 0, "no receipts issued")
		},
		10*time.Second,
		100*time.Millisecond,
	)

	var receipt entities.IssueReceiptRequest
	var ok bool
	for _, issuedReceipt := range receiptsService.IssuedReceipts {
		if issuedReceipt.TicketID != ticket.TicketID {
			continue
		}
		receipt = issuedReceipt
		ok = true
		break
	}

	require.Truef(t, ok, "receipt for ticket %s not found", ticket.TicketID)
	assert.Equal(t, ticket.TicketID, receipt.TicketID)
	assert.Equal(t, ticket.Price.Amount, receipt.Price.Amount)
	assert.Equal(t, ticket.Price.Currency, receipt.Price.Currency)
}

func assertRowInTicketsToPrint(t *testing.T, spreadsheetsAPI *adapters.SpreadsheetsAPIStub, ticket ticketsHttp.TicketStatus) {
	t.Helper()

	parentT := t

	assert.EventuallyWithT(
		t,
		func(t *assert.CollectT) {
			rowsLen := len(spreadsheetsAPI.Rows)
			parentT.Log("rows", rowsLen)
			assert.Greater(t, rowsLen, 0, "no rows")
		},
		10*time.Second,
		100*time.Millisecond,
	)

	var rowIndex = -1
	for i, sheetName := range spreadsheetsAPI.SpreadsheetNames {
		if sheetName == "tickets-to-print" {
			row := spreadsheetsAPI.Rows[i]
			for _, cell := range row {
				if cell == ticket.TicketID {
					rowIndex = i
					break
				}
			}
		}
		if rowIndex != -1 {
			break
		}
	}

	require.NotEqual(t, -1, rowIndex, "row for ticket %s not found in tickets-to-print", ticket.TicketID)
	assert.Equal(t, "tickets-to-print", spreadsheetsAPI.SpreadsheetNames[rowIndex])
}

func assertRowInticketsToRefund(t *testing.T, spreadsheetsAPI *adapters.SpreadsheetsAPIStub, ticket ticketsHttp.TicketStatus) {
	t.Helper()

	parentT := t

	assert.EventuallyWithT(
		t,
		func(t *assert.CollectT) {
			rowsLen := len(spreadsheetsAPI.Rows)
			parentT.Log("rows", rowsLen)
			assert.Greater(t, rowsLen, 0, "no rows")
		},
		10*time.Second,
		100*time.Millisecond,
	)

	var rowIndex = -1
	for i, sheetName := range spreadsheetsAPI.SpreadsheetNames {
		if sheetName == "tickets-to-refund" {
			row := spreadsheetsAPI.Rows[i]
			for _, cell := range row {
				if cell == ticket.TicketID {
					rowIndex = i
					break
				}
			}
		}
		if rowIndex != -1 {
			break
		}
	}

	require.NotEqual(t, -1, rowIndex, "row for ticket %s not found in tickets-to-refund", ticket.TicketID)
	assert.Equal(t, "tickets-to-refund", spreadsheetsAPI.SpreadsheetNames[rowIndex])

}

func assertTicketPrinted(t *testing.T, filesAPI *adapters.FilesAPIStub, ticket ticketsHttp.TicketStatus) {
	t.Helper()

	assert.EventuallyWithT(
		t,
		func(t *assert.CollectT) {
			assert.True(t, filesAPI.WasCalled(), "Files API should be called")
		},
		10*time.Second,
		100*time.Millisecond,
	)

	calls := filesAPI.GetCalls()
	require.Greater(t, len(calls), 0, "No files created")

	expectedFileID := ticket.TicketID + "-ticket.html"
	found := false
	for _, call := range calls {
		if call.FileID == expectedFileID {
			found = true
			break
		}
	}
	assert.True(t, found, "File %s not created", expectedFileID)
}

func sendBookTicketsRequest(t *testing.T, showID, customerEmail string, numberOfTickets int) string {
	t.Helper()

	request := map[string]interface{}{
		"show_id":           showID,
		"number_of_tickets": numberOfTickets,
		"customer_email":    customerEmail,
	}

	payload, err := json.Marshal(request)
	require.NoError(t, err)

	httpReq, err := http.NewRequest(
		http.MethodPost,
		"http://localhost:8080/book-tickets",
		bytes.NewBuffer(payload),
	)
	require.NoError(t, err)

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var response map[string]string
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	bookingID, ok := response["booking_id"]
	require.True(t, ok, "booking_id not found in response")
	require.NotEmpty(t, bookingID, "booking_id is empty")

	return bookingID
}

func assertBookingCreated(t *testing.T, db *sqlx.DB, bookingID, showID, customerEmail string, numberOfTickets int) {
	t.Helper()

	assert.EventuallyWithT(
		t,
		func(t *assert.CollectT) {
			// Query booking từ DB
			var booking entities.Booking
			err := db.Get(&booking, "SELECT * FROM bookings WHERE booking_id = $1", bookingID)

			// Assert booking tồn tại
			if !assert.NoError(t, err, "booking not found in database") {
				return
			}

			// Assert các fields
			assert.Equal(t, bookingID, booking.BookingID)
			assert.Equal(t, showID, booking.ShowID)
			assert.Equal(t, customerEmail, booking.CustomerEmail)
			assert.Equal(t, numberOfTickets, booking.NumberOfTickets)
		},
		10*time.Second,
		100*time.Millisecond,
	)
}

func createShow(t *testing.T, db *sqlx.DB) string {
	t.Helper()

	showID := uuid.New().String()

	_, err := db.Exec(`
          INSERT INTO shows (show_id, dead_nation_id, number_of_tickets, start_time, title, venue)
          VALUES ($1, $2, $3, $4, $5, $6)
      `, showID, uuid.New().String(), 100, time.Now(), "Test Concert", "Test Venue")

	require.NoError(t, err)

	return showID
}

func assertBookingInDeadNation(t *testing.T, deadNationClient *adapters.DeadNationStub, bookingID, customerEmail string, numberOfTickets int) {
	t.Helper()

	assert.EventuallyWithT(
		t,
		func(t *assert.CollectT) {
			bookingsLen := len(deadNationClient.DeadNationBookings)
			if !assert.Greater(t, bookingsLen, 0, "no bookings sent to Dead Nation") {
				return
			}

			// Find booking by ID
			var found bool
			for _, booking := range deadNationClient.DeadNationBookings {
				if booking.BookingID.String() == bookingID {
					found = true
					assert.Equal(t, customerEmail, booking.CustomerEmail)
					assert.Equal(t, numberOfTickets, booking.NumberOfTickets)
					assert.NotEqual(t, uuid.Nil, booking.DeadNationEventID, "DeadNationEventID should not be nil")
					break
				}
			}

			assert.True(t, found, "booking %s not found in Dead Nation bookings", bookingID)
		},
		10*time.Second,
		100*time.Millisecond,
	)
}

func waitForHttpServer(t *testing.T) {
	t.Helper()

	require.EventuallyWithT(
		t,
		func(t *assert.CollectT) {
			resp, err := http.Get("http://localhost:8080/health")
			if !assert.NoError(t, err) {
				return
			}
			defer resp.Body.Close()

			if assert.Less(t, resp.StatusCode, 300, "API not ready, http status: %d", resp.StatusCode) {
				return
			}
		},
		time.Second*10,
		time.Millisecond*50,
	)
}
