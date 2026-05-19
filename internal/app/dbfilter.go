package app

import (
	"context"
	"path/filepath"
	"slices"

	"github.com/kstieger/mongocp/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var systemDBs = []string{"admin", "local", "config"}

// FilterDatabases applies include and exclusion filters to the list of database names.
func FilterDatabases(dbs []string, includeList []string, excludeSystem bool, excludeList []string) []domain.Database {
	var result []domain.Database
	for _, db := range dbs {
		if excludeSystem && slices.Contains(systemDBs, db) {
			continue
		}
		if len(includeList) > 0 && !matchesAnyPattern(db, includeList) {
			continue
		}
		if matchesAnyPattern(db, excludeList) {
			continue
		}
		result = append(result, domain.Database{Name: db})
	}
	return result
}

func matchesAnyPattern(name string, patterns []string) bool {
	for _, pattern := range patterns {
		matched, err := filepath.Match(pattern, name)
		if err == nil && matched {
			return true
		}
		if name == pattern {
			return true
		}
	}

	return false
}

// ListDatabaseNames returns all database names from the MongoDB server.
func ListDatabaseNames(ctx context.Context, client *mongo.Client) ([]string, error) {
	return client.ListDatabaseNames(ctx, bson.D{})
}
