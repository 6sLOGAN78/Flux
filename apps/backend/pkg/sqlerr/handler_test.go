package sqlerr_test

import (
	"errors"
	"testing"

	"flux/apps/backend/internal/errs"
	"flux/apps/backend/pkg/sqlerr"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestMapCode(t *testing.T) {
	tests := []struct {
		sqlState string
		expected sqlerr.Code
	}{
		{"23502", sqlerr.NotNullViolation},
		{"23503", sqlerr.ForeignKeyViolation},
		{"23505", sqlerr.UniqueViolation},
		{"23514", sqlerr.CheckViolation},
		{"23P01", sqlerr.ExcludeViolation},
		{"25P02", sqlerr.TransactionFailed},
		{"40P01", sqlerr.DeadlockDetected},
		{"53300", sqlerr.TooManyConnections},
		{"99999", sqlerr.Other},
	}

	for _, tt := range tests {
		got := sqlerr.MapCode(tt.sqlState)
		if got != tt.expected {
			t.Errorf("MapCode(%s) = %s, expected %s", tt.sqlState, got, tt.expected)
		}
	}
}

func TestMapSeverity(t *testing.T) {
	tests := []struct {
		severity string
		expected sqlerr.Severity
	}{
		{"ERROR", sqlerr.SeverityError},
		{"FATAL", sqlerr.SeverityFatal},
		{"PANIC", sqlerr.SeverityPanic},
		{"WARNING", sqlerr.SeverityWarning},
		{"NOTICE", sqlerr.SeverityNotice},
		{"UNKNOWN", sqlerr.SeverityError},
	}

	for _, tt := range tests {
		got := sqlerr.MapSeverity(tt.severity)
		if got != tt.expected {
			t.Errorf("MapSeverity(%s) = %s, expected %s", tt.severity, got, tt.expected)
		}
	}
}

func TestConvertPgError(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code:           "23505",
		Severity:       "ERROR",
		Message:        "duplicate key value violates unique constraint",
		TableName:      "links",
		ConstraintName: "unique_links_short_code",
	}

	converted := sqlerr.ConvertPgError(pgErr)
	if converted == nil {
		t.Fatalf("expected non-nil Error")
	}

	if converted.Code != sqlerr.UniqueViolation {
		t.Errorf("expected UniqueViolation code, got %s", converted.Code)
	}

	if converted.Unwrap() != pgErr {
		t.Errorf("expected unwrapped error to match pgErr")
	}
}

func TestHandleError_UniqueViolation(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code:           "23505",
		Severity:       "ERROR",
		Message:        "duplicate key value violates unique constraint",
		TableName:      "links",
		ConstraintName: "unique_links_short_code",
	}

	err := sqlerr.HandleError(pgErr)
	var httpErr *errs.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *errs.HTTPError, got %T", err)
	}

	if httpErr.StatusCode != 400 {
		t.Errorf("expected status code 400, got %d", httpErr.StatusCode)
	}
}

func TestHandleError_ErrNoRows(t *testing.T) {
	err := sqlerr.HandleError(pgx.ErrNoRows)
	var httpErr *errs.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *errs.HTTPError, got %T", err)
	}

	if httpErr.StatusCode != 404 {
		t.Errorf("expected status code 404, got %d", httpErr.StatusCode)
	}
}
