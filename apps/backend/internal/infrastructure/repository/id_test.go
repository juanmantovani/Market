package repository

import (
	"regexp"
	"testing"
)

func TestNewIDIsAUUIDv7(t *testing.T) {
	uuidV7 := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

	first, second := NewID(), NewID()

	if !uuidV7.MatchString(first) {
		t.Fatalf("NewID() = %q, want a UUIDv7", first)
	}
	if first == second {
		t.Fatal("NewID() repeated an ID")
	}
}
