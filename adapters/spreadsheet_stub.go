package adapters

import (
	"context"
	"sync"
)

type SpreadsheetsAPIStub struct {
	lock             sync.Mutex
	SpreadsheetNames []string
	Rows             [][]string
}

func (s *SpreadsheetsAPIStub) AppendRow(ctx context.Context, spreadsheetName string, row []string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.SpreadsheetNames = append(s.SpreadsheetNames, spreadsheetName)
	s.Rows = append(s.Rows, row)
	return nil
}
