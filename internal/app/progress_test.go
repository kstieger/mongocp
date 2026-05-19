package app

import "testing"

func TestFormatTaskWorkerLabel(t *testing.T) {
	got := formatTaskWorkerLabel(2, "example_db", "example_coll")
	want := "2: example_db.example_coll"
	if got != want {
		t.Fatalf("formatTaskWorkerLabel() = %q, want %q", got, want)
	}
}

func TestFormatIdleWorkerLabel(t *testing.T) {
	got := formatIdleWorkerLabel(4)
	want := "4: idle"
	if got != want {
		t.Fatalf("formatIdleWorkerLabel() = %q, want %q", got, want)
	}
}
