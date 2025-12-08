package adapters

import (
	"context"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients/files"
	"net/http"
	"sync"
)

type FilesAPIStub struct {
	mu sync.Mutex

	calls []FileCall
}

type FileCall struct {
	FileID  string
	Content string
}

func NewFilesAPIStub() *FilesAPIStub {
	return &FilesAPIStub{
		calls: []FileCall{},
	}
}

func (s *FilesAPIStub) PutFilesFileIdContentWithTextBodyWithResponse(
	ctx context.Context,
	fileID string,
	body string) (*files.PutFilesFileIdContentResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.calls = append(s.calls, FileCall{
		FileID:  fileID,
		Content: body,
	})

	return &files.PutFilesFileIdContentResponse{
		HTTPResponse: &http.Response{
			StatusCode: http.StatusOK,
		},
	}, nil
}

func (s *FilesAPIStub) WasCalled() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.calls) > 0
}

func (s *FilesAPIStub) GetCalls() []FileCall {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calls
}

func (s *FilesAPIStub) GetCallCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.calls)
}

func (s *FilesAPIStub) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = []FileCall{}
}
