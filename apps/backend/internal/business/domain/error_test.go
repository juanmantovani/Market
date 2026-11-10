package domain

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrorExposesMessageAndUnwrapsCause(t *testing.T) {
	cause := errors.New("connection refused")
	err := ErrStorageUnavailable.WithCause(cause)

	if err.Error() != ErrStorageUnavailable.Message {
		t.Fatalf("Error() = %q, want %q", err.Error(), ErrStorageUnavailable.Message)
	}
	if !errors.Is(err, cause) {
		t.Fatal("errors.Is(err, cause) = false, want true")
	}
	if !errors.Is(err, ErrStorageUnavailable) {
		t.Fatal("errors.Is(err, ErrStorageUnavailable) = false, want true")
	}
	if ErrStorageUnavailable.Cause != nil {
		t.Fatal("WithCause mutated the shared sentinel")
	}
}

func TestErrorIsComparesByCode(t *testing.T) {
	if errors.Is(ErrStorageUnavailable, ErrForbidden) {
		t.Fatal("errors with different codes matched")
	}
	if errors.Is(ErrForbidden, errors.New("forbidden")) {
		t.Fatal("a plain error matched a domain error")
	}
}

func TestFromStorage(t *testing.T) {
	if FromStorage(nil) != nil {
		t.Fatal("FromStorage(nil) != nil")
	}

	wrappedBusiness := fmt.Errorf("lookup: %w", ErrForbidden)
	if got := FromStorage(wrappedBusiness); got != wrappedBusiness {
		t.Fatalf("FromStorage changed a business error: %v", got)
	}

	cause := errors.New("disk full")
	got := FromStorage(cause)
	var domainErr *Error
	if !errors.As(got, &domainErr) || domainErr.Code != ErrStorageUnavailable.Code {
		t.Fatalf("FromStorage(%v) = %v, want storage_unavailable", cause, got)
	}
	if !errors.Is(got, cause) {
		t.Fatal("FromStorage lost the cause")
	}
}
