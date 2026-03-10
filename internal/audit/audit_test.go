package audit

import (
	"os"
	"testing"
	"time"
)

func TestAuditStore_Record(t *testing.T) {
	dbPath := "test_audit.db"
	defer os.Remove(dbPath)

	store, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	err = store.Record("test.tool", 500*time.Millisecond, true)
	if err != nil {
		t.Fatalf("failed to record execution: %v", err)
	}

	var count int64
	store.db.Model(&ToolExecutionRecord{}).Count(&count)
	if count != 1 {
		t.Errorf("expected 1 record, got %d", count)
	}

	var record ToolExecutionRecord
	store.db.First(&record)
	if record.Tool != "test.tool" {
		t.Errorf("expected tool test.tool, got %s", record.Tool)
	}
	if record.Status != "SUCCESS" {
		t.Errorf("expected status SUCCESS, got %s", record.Status)
	}
}
