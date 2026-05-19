package app

import "testing"

func TestFilterDatabasesWithIncludeAndExcludePatterns(t *testing.T) {
	dbs := []string{"admin", "dev", "dev_team", "prod", "prod_archive", "misc"}

	filtered := FilterDatabases(dbs, []string{"dev*", "prod*"}, true, []string{"*archive", "dev_team"})

	if len(filtered) != 2 {
		t.Fatalf("expected 2 databases, got %d", len(filtered))
	}
	if filtered[0].Name != "dev" {
		t.Fatalf("expected first database to be dev, got %q", filtered[0].Name)
	}
	if filtered[1].Name != "prod" {
		t.Fatalf("expected second database to be prod, got %q", filtered[1].Name)
	}
}

func TestParseListFlagTrimsAndSkipsEmptyPatterns(t *testing.T) {
	patterns := parseListFlag(" dev* , ,prod ")

	if len(patterns) != 2 {
		t.Fatalf("expected 2 patterns, got %d", len(patterns))
	}
	if patterns[0] != "dev*" {
		t.Fatalf("expected first pattern to be dev*, got %q", patterns[0])
	}
	if patterns[1] != "prod" {
		t.Fatalf("expected second pattern to be prod, got %q", patterns[1])
	}
}
