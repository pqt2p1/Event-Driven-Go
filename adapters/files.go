package adapters

import (
	"context"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients/files"
)

type FilesAPIClient struct {
	clients *clients.Clients
}

func NewFilesAPIClient(clients *clients.Clients) *FilesAPIClient {
	return &FilesAPIClient{
		clients: clients,
	}
}

func (f *FilesAPIClient) PutFilesFileIdContentWithTextBodyWithResponse(ctx context.Context, fileID string, body string) (*files.PutFilesFileIdContentResponse, error) {
	p, err := f.clients.Files.PutFilesFileIdContentWithTextBodyWithResponse(ctx, fileID, body)
	return p, err
}
