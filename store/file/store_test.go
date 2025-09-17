package file_test

import (
	"bl-shifts/store/file"
	"os"
	"testing"
)

func TestExistingKeys(t *testing.T) {
	s := file.NewStore("test_store.json")
	defer func() {
		_ = os.Remove("test_store.json")
	}()
	codes := []string{"CODE1", "CODE2", "CODE3"}
	codesToSend, err := s.FilterAndSaveCodes(t.Context(), codes)
	if err != nil {
		t.Error("unexpected error:", err)
		t.FailNow()
	}
	if len(codesToSend) != 3 {
		t.Error("expected 3 new codes, got", len(codesToSend))
		t.FailNow()
	}
	newCodes := []string{"CODE2", "CODE3", "CODE4"}
	codesToSend, err = s.FilterAndSaveCodes(t.Context(), newCodes)
	if err != nil {
		t.Error("unexpected error:", err)
		t.FailNow()
	}
	if len(codesToSend) != 1 || codesToSend[0] != "CODE4" {
		t.Error("expected 1 new code (CODE4), got", codesToSend)
		t.FailNow()
	}
}
