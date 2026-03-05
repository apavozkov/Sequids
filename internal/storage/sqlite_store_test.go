package storage

import (
	"context"
	"path/filepath"
	"testing"
)

func TestSQLiteStoreScenarioRoundTrip(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "sequids.db")
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = store.Close() }()

	ctx := context.Background()
	if err := store.Init(ctx); err != nil {
		t.Fatal(err)
	}

	id := "scn_test_roundtrip"
	dsl := "name: demo\ndevices:\n  - id: d1\n    type: temperature\n    topic: iot/demo\n    frequency_hz: 1\n    formula_ref: temp_daily_sine\n"
	if err := store.SaveScenario(ctx, id, "demo", dsl); err != nil {
		t.Fatal(err)
	}

	got, err := store.GetScenario(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if got != dsl {
		t.Fatalf("unexpected dsl: got %q want %q", got, dsl)
	}
}
