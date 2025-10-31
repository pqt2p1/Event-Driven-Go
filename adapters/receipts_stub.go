package adapters

import (
	"context"
	"sync"
	"tickets/entities"
)

type ReceiptsServiceStub struct {
	lock           sync.Mutex
	IssuedReceipts []entities.IssueReceiptRequest
}

func (r *ReceiptsServiceStub) IssueReceipt(ctx context.Context, request entities.IssueReceiptRequest) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.IssuedReceipts = append(r.IssuedReceipts, request)
	return nil
}
