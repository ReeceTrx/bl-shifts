package file

import (
	"bl-shifts/store"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type fileStore struct {
	filename string
}

func NewStore(filename string) store.Store {
	return &fileStore{
		filename: filename,
	}
}

func (s *fileStore) FilterAndSaveCodes(ctx context.Context, codes []string) ([]string, error) {
	file, err := os.OpenFile(s.filename, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	fileOpen := true
	defer func() {
		if fileOpen {
			file.Close()
		}
	}()
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read from file: %w", err)
	}
	file.Close()
	fileOpen = false
	var existingCodes map[string]bool
	err = json.Unmarshal(data, &existingCodes)
	if err != nil && len(data) > 0 {
		return nil, fmt.Errorf("failed to unmarshal existing codes: %w", err)
	}
	if err != nil {
		existingCodes = map[string]bool{}
	}
	codesToSend := []string{}
	for _, code := range codes {
		if existingCodes[code] {
			continue
		}
		codesToSend = append(codesToSend, code)
		existingCodes[code] = true
	}
	newData, err := json.Marshal(existingCodes)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal codes: %w", err)
	}
	err = os.WriteFile(s.filename, newData, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to write to file: %w", err)
	}
	return codesToSend, nil
}
