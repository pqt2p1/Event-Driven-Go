package event

import (
	"context"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients/files"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"tickets/db"

	"tickets/entities"
)

type Handler struct {
	spreadsheetsAPI  SpreadsheetsAPI
	receiptsService  ReceiptsService
	ticketsRepo      *db.TicketsRepository
	filesApi         FilesAPI
	eventBus         *cqrs.EventBus
	showsRepo        *db.ShowsRepository
	deadNationClient DeadNationClient
}

func NewHandler(
	spreadsheetsAPI SpreadsheetsAPI,
	receiptsService ReceiptsService,
	ticketsRepo *db.TicketsRepository,
	filesApi FilesAPI,
	eventBus *cqrs.EventBus,
	showsRepo *db.ShowsRepository,
	deadNationClient DeadNationClient,
) Handler {
	if spreadsheetsAPI == nil {
		panic("missing spreadsheetsAPI")
	}
	if receiptsService == nil {
		panic("missing receiptsService")
	}

	return Handler{
		spreadsheetsAPI:  spreadsheetsAPI,
		receiptsService:  receiptsService,
		ticketsRepo:      ticketsRepo,
		filesApi:         filesApi,
		eventBus:         eventBus,
		showsRepo:        showsRepo,
		deadNationClient: deadNationClient,
	}
}

type SpreadsheetsAPI interface {
	AppendRow(ctx context.Context, sheetName string, row []string) error
}

type ReceiptsService interface {
	IssueReceipt(ctx context.Context, request entities.IssueReceiptRequest) error
}

type FilesAPI interface {
	PutFilesFileIdContentWithTextBodyWithResponse(
		ctx context.Context,
		fileID string,
		body string,
	) (*files.PutFilesFileIdContentResponse, error)
}

type DeadNationClient interface {
	PostTicketBooking(
		ctx context.Context,
		bookingID string,
		eventID string,
		numberOfTickets int,
		customerEmail string,
	) error
}

func (h Handler) EventHandlers() []cqrs.EventHandler {
	return []cqrs.EventHandler{
		cqrs.NewEventHandler("issue_receipt", h.IssueReceipt),
		cqrs.NewEventHandler("append_to_tracker", h.AppendToTracker),
		cqrs.NewEventHandler("cancel_ticket", h.CancelTicket),
		cqrs.NewEventHandler("store_tickets", h.StoreTickets),
	}
}
