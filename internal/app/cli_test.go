package app

import (
	"flag"
	"testing"
)

func TestEffectiveLogLevel(t *testing.T) {
	if got := effectiveLogLevel("info", true); got != "fatal" {
		t.Fatalf("effectiveLogLevel(info, true) = %q, want %q", got, "fatal")
	}

	if got := effectiveLogLevel("debug", false); got != "debug" {
		t.Fatalf("effectiveLogLevel(debug, false) = %q, want %q", got, "debug")
	}
}

func TestIsFlagProvided(t *testing.T) {
	original := flag.CommandLine
	t.Cleanup(func() {
		flag.CommandLine = original
	})

	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
	flag.String("log-level", "info", "")

	if err := flag.CommandLine.Parse([]string{"-log-level", "debug"}); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if !isFlagProvided("log-level") {
		t.Fatal("expected log-level to be marked as provided")
	}
	if isFlagProvided("worker") {
		t.Fatal("did not expect worker to be marked as provided")
	}
}

func TestSanitizeMongoURI(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "password in standard uri",
			raw:  "mongodb://user:secret@localhost:27017/admin?authSource=admin",
			want: "mongodb://user:xxxxx@localhost:27017/admin?authSource=admin",
		},
		{
			name: "password in srv uri",
			raw:  "mongodb+srv://user:secret@cluster.example.net/test?retryWrites=true",
			want: "mongodb+srv://user:xxxxx@cluster.example.net/test?retryWrites=true",
		},
		{
			name: "no password remains unchanged",
			raw:  "mongodb://localhost:27017/admin",
			want: "mongodb://localhost:27017/admin",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := sanitizeMongoURI(test.raw)
			if got != test.want {
				t.Fatalf("sanitizeMongoURI(%q) = %q, want %q", test.raw, got, test.want)
			}
		})
	}
}
