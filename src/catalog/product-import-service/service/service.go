package service

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/shopos/product-import-service/domain"
)

// Storer is the subset of store.Storer the service layer needs.
type Storer interface {
	Create(fileName string, format domain.ImportFormat) (*domain.ImportJob, error)
	Get(id string) (*domain.ImportJob, error)
	List() ([]*domain.ImportJob, error)
	UpdateProgress(id string, processed, errorRows int, errs []domain.ImportError) error
	Complete(id string) error
	Fail(id string, errMsg string) error
}

// ImportService orchestrates import job creation and processing.
type ImportService struct {
	store Storer
}

// New returns an initialised ImportService.
func New(store Storer) *ImportService {
	return &ImportService{store: store}
}

// CreateJob creates a new import job record.
func (s *ImportService) CreateJob(fileName string, format domain.ImportFormat) (*domain.ImportJob, error) {
	return s.store.Create(fileName, format)
}

// GetJob retrieves a single import job by ID.
func (s *ImportService) GetJob(id string) (*domain.ImportJob, error) {
	return s.store.Get(id)
}

// ListJobs returns all import jobs.
func (s *ImportService) ListJobs() ([]*domain.ImportJob, error) {
	return s.store.List()
}

// ProcessCSV parses raw CSV bytes and records per-row validation results.
// Required columns: sku, name, price.
func (s *ImportService) ProcessCSV(jobID string, data []byte) error {
	r := csv.NewReader(bytes.NewReader(data))
	records, err := r.ReadAll()
	if err != nil {
		if ferr := s.store.Fail(jobID, fmt.Sprintf("CSV parse error: %v", err)); ferr != nil {
			return ferr
		}
		return fmt.Errorf("service.ProcessCSV parse: %w", err)
	}

	if len(records) < 2 {
		return s.store.Fail(jobID, "CSV has no data rows")
	}

	header := records[0]
	colIdx := make(map[string]int, len(header))
	for i, h := range header {
		colIdx[strings.ToLower(strings.TrimSpace(h))] = i
	}

	dataRows := records[1:]
	total := len(dataRows)
	var errs []domain.ImportError

	for i, row := range dataRows {
		rowNum := i + 2 // 1-based, accounting for header
		rowErrs := validateCSVRow(rowNum, row, colIdx)
		errs = append(errs, rowErrs...)
	}

	errCount := len(errs)
	processed := total - errCount
	if processed < 0 {
		processed = 0
	}

	if err := s.store.UpdateProgress(jobID, processed, errCount, errs); err != nil {
		return err
	}
	return s.store.Complete(jobID)
}

// ProcessJSON parses a JSON array of product objects and validates each entry.
// Each object must have: sku (string), name (string), price (number > 0).
func (s *ImportService) ProcessJSON(jobID string, data []byte) error {
	var rows []map[string]any
	if err := json.Unmarshal(data, &rows); err != nil {
		if ferr := s.store.Fail(jobID, fmt.Sprintf("JSON parse error: %v", err)); ferr != nil {
			return ferr
		}
		return fmt.Errorf("service.ProcessJSON parse: %w", err)
	}

	total := len(rows)
	var errs []domain.ImportError

	for i, row := range rows {
		rowNum := i + 1
		rowErrs := validateJSONRow(rowNum, row)
		errs = append(errs, rowErrs...)
	}

	errCount := len(errs)
	processed := total - errCount
	if processed < 0 {
		processed = 0
	}

	if err := s.store.UpdateProgress(jobID, processed, errCount, errs); err != nil {
		return err
	}
	return s.store.Complete(jobID)
}

// ---- validation helpers -----------------------------------------------------

func validateCSVRow(rowNum int, row []string, colIdx map[string]int) []domain.ImportError {
	var errs []domain.ImportError

	get := func(col string) string {
		idx, ok := colIdx[col]
		if !ok || idx >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[idx])
	}

	if sku := get("sku"); sku == "" {
		errs = append(errs, domain.ImportError{Row: rowNum, Field: "sku", Message: "sku is required"})
	}
	if name := get("name"); name == "" {
		errs = append(errs, domain.ImportError{Row: rowNum, Field: "name", Message: "name is required"})
	}
	priceStr := get("price")
	if priceStr == "" {
		errs = append(errs, domain.ImportError{Row: rowNum, Field: "price", Message: "price is required"})
	} else {
		p, err := strconv.ParseFloat(priceStr, 64)
		if err != nil || p <= 0 {
			errs = append(errs, domain.ImportError{Row: rowNum, Field: "price", Message: "price must be a positive number"})
		}
	}
	return errs
}

func validateJSONRow(rowNum int, row map[string]any) []domain.ImportError {
	var errs []domain.ImportError

	skuVal, hasSku := row["sku"]
	if !hasSku || fmt.Sprintf("%v", skuVal) == "" {
		errs = append(errs, domain.ImportError{Row: rowNum, Field: "sku", Message: "sku is required"})
	}

	nameVal, hasName := row["name"]
	if !hasName || fmt.Sprintf("%v", nameVal) == "" {
		errs = append(errs, domain.ImportError{Row: rowNum, Field: "name", Message: "name is required"})
	}

	priceVal, hasPrice := row["price"]
	if !hasPrice {
		errs = append(errs, domain.ImportError{Row: rowNum, Field: "price", Message: "price is required"})
	} else {
		switch p := priceVal.(type) {
		case float64:
			if p <= 0 {
				errs = append(errs, domain.ImportError{Row: rowNum, Field: "price", Message: "price must be greater than 0"})
			}
		default:
			errs = append(errs, domain.ImportError{Row: rowNum, Field: "price", Message: "price must be a number"})
		}
	}
	return errs
}
