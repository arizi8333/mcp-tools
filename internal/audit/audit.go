package audit

import (
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ToolExecutionRecord represents a single tool execution in the audit trail.
type ToolExecutionRecord struct {
	ID        uint   `gorm:"primaryKey"`
	Tool      string `gorm:"index"`
	Status    string
	Duration  float64
	Timestamp time.Time
}

// Store handles persistence of audit records.
type Store struct {
	db *gorm.DB
}

// NewStore initializes a new audit store using SQLite.
func NewStore(dbPath string) (*Store, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto Migration
	if err := db.AutoMigrate(&ToolExecutionRecord{}); err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

// Record saves an execution record to the database.
func (s *Store) Record(tool string, duration time.Duration, success bool) error {
	status := "SUCCESS"
	if !success {
		status = "FAILED"
	}

	record := ToolExecutionRecord{
		Tool:      tool,
		Status:    status,
		Duration:  duration.Seconds(),
		Timestamp: time.Now(),
	}

	return s.db.Create(&record).Error
}
