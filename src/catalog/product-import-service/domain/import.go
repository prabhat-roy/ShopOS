package domain

import "time"

// ImportStatus represents the lifecycle state of an import job.
type ImportStatus string

const (
	ImportPending    ImportStatus = "PENDING"
	ImportProcessing ImportStatus = "PROCESSING"
	ImportCompleted  ImportStatus = "COMPLETED"
	ImportFailed     ImportStatus = "FAILED"
)

// ImportFormat is the file format of the uploaded import data.
type ImportFormat string

const (
	FormatCSV  ImportFormat = "CSV"
	FormatJSON ImportFormat = "JSON"
)

// ImportError captures a single row-level validation failure.
type ImportError struct {
	Row     int    `json:"row"`
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ImportJob is the root aggregate for a product import operation.
type ImportJob struct {
	ID            string       `json:"id"`
	FileName      string       `json:"file_name"`
	Format        ImportFormat `json:"format"`
	Status        ImportStatus `json:"status"`
	TotalRows     int          `json:"total_rows"`
	ProcessedRows int          `json:"processed_rows"`
	ErrorRows     int          `json:"error_rows"`
	Errors        []ImportError `json:"errors"`
	CreatedAt     time.Time    `json:"created_at"`
	CompletedAt   *time.Time   `json:"completed_at,omitempty"`
}
