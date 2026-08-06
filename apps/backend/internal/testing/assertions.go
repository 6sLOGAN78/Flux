package testing

import (
	"reflect"
	"testing"
)

// AssertNoError fails the test if err is not nil.
func AssertNoError(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: expected no error, got: %v", msg, err)
	}
}

// AssertError fails the test if err is nil.
func AssertError(t *testing.T, err error, msg string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: expected error, got nil", msg)
	}
}

// AssertEqual fails the test if expected != actual.
func AssertEqual(t *testing.T, expected, actual interface{}, msg string) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("%s: expected '%v', got '%v'", msg, expected, actual)
	}
}

// AssertStatus fails the test if expected HTTP status != actual status.
func AssertStatus(t *testing.T, expectedStatus, actualStatus int) {
	t.Helper()
	if expectedStatus != actualStatus {
		t.Fatalf("HTTP status mismatch: expected %d, got %d", expectedStatus, actualStatus)
	}
}
