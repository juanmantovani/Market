package gs1argentina

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"market/backend/internal/business/domain"
)

// Manual check against the real portal. Not part of CI.
func TestLivePortal(t *testing.T) {
	if os.Getenv("GS1_LIVE") != "1" {
		t.Skip("set GS1_LIVE=1 to call the real GS1 Argentina portal")
	}
	username, password := os.Getenv("GS1_USERNAME"), os.Getenv("GS1_PASSWORD")
	if username == "" || password == "" {
		t.Fatal("GS1_USERNAME and GS1_PASSWORD are required")
	}

	client := NewClient(nil, "", username, password, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	started := time.Now()
	got, err := client.Lookup(ctx, "7790895000997")
	t.Logf("lookup took %s", time.Since(started))
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if got.Name == "" {
		t.Fatal("empty name")
	}
	description := ""
	if got.Description != nil {
		description = *got.Description
	}
	t.Logf("name=%q description=%q", got.Name, description)

	_, err = client.Lookup(ctx, "7790000000010")
	if !errors.Is(err, domain.ErrExternalProductNotFound) {
		t.Fatalf("unknown GTIN error = %v", err)
	}
}
