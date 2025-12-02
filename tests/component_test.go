package tests_test

import (
	"bytes"
	"context"
	"encoding/json"
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

	go func() {
		svc := service.New(
			db,
			redisClient,
			spreadsheetsAPI,
			receiptsService,
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
