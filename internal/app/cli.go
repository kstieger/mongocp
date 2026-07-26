package app

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
)

var (
	Version = "dev"
	BuiltAt = "unknown"
	GitHash = "unknown"
)

// Run is the main entry point for the CLI using Go stdlib flag package.
func Run() error {
	var srcURI, dstURI, includeDbsStr, excludeDbsStr, logLevel string
	var workerCount int
	var dryRun, excludeSystem, showVersion bool

	flag.StringVar(&srcURI, "src", "", "Source MongoDB URI")
	flag.StringVar(&dstURI, "dst", "", "Destination MongoDB URI")
	flag.StringVar(&includeDbsStr, "include-dbs", "", "Comma-separated list of database patterns to include")
	flag.StringVar(&excludeDbsStr, "exclude-dbs", "", "Comma-separated list of database patterns to exclude")
	flag.IntVar(&workerCount, "worker", 10, "Number of parallel workers")
	flag.BoolVar(&dryRun, "dry-run", false, "Show what would be copied, but do not modify destination")
	flag.BoolVar(&excludeSystem, "exclude-system_dbs", true, "Exclude system databases (admin, local, config)")
	flag.StringVar(&logLevel, "log-level", "info", "Set log level (info, debug, warn, error)")
	flag.StringVar(&logLevel, "loglevel", "info", "Alias for -log-level")
	flag.BoolVar(&showVersion, "version", false, "Print version information and exit")

	flag.Parse()

	if showVersion {
		fmt.Printf("mongocp, version %s\nbuilt at %s\ngit hash %s\n", Version, BuiltAt, GitHash)
		return nil
	}

	if srcURI == "" || dstURI == "" {
		flag.Usage()
		os.Exit(1)
	}

	progressEnabled := !isFlagProvided("log-level") && !isFlagProvided("loglevel")

	includeDbs := parseListFlag(includeDbsStr)
	excludeDbs := parseListFlag(excludeDbsStr)

	ctx := context.Background()
	logger := SetupLogger(effectiveLogLevel(logLevel, progressEnabled))

	logger.Info("Using database filters", "include_dbs", strings.Join(includeDbs, ","), "exclude_dbs", strings.Join(excludeDbs, ","))
	logger.Info("Connecting to source MongoDB", "uri", sanitizeMongoURI(srcURI))
	srcClient, err := ConnectMongo(ctx, srcURI)
	if err != nil {
		logger.Error("Failed to connect to source", "err", err)
		return err
	}
	defer func() {
		_ = srcClient.Disconnect(ctx)
	}()

	logger.Info("Connecting to destination MongoDB", "uri", sanitizeMongoURI(dstURI))
	dstClient, err := ConnectMongo(ctx, dstURI)
	if err != nil {
		logger.Error("Failed to connect to destination", "err", err)
		return err
	}
	defer func() {
		_ = dstClient.Disconnect(ctx)
	}()

	logger.Info("Listing databases on source")
	dbNames, err := ListDatabaseNames(ctx, srcClient)
	if err != nil {
		logger.Error("Failed to list databases", "err", err)
		return err
	}

	dbs := FilterDatabases(dbNames, includeDbs, excludeSystem, excludeDbs)
	if len(dbs) == 0 {
		logger.Warn("No databases to copy after filtering")
		return nil
	}
	for _, db := range dbs {
		logger.Info("Database selected for copy", "db", db.Name)
	}

	logger.Info("Starting copy", "db_count", len(dbs), "worker_count", workerCount, "dry_run", dryRun, "progress", progressEnabled)
	err = CopyCollections(ctx, srcClient, dstClient, dbs, workerCount, dryRun, progressEnabled, logger)
	if err != nil {
		logger.Error("Copy failed", "err", err)
		return err
	}
	logger.Info("Copy completed successfully")
	return nil
}

func parseListFlag(raw string) []string {
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	patterns := make([]string, 0, len(parts))
	for _, part := range parts {
		pattern := strings.TrimSpace(part)
		if pattern == "" {
			continue
		}
		patterns = append(patterns, pattern)
	}
	if len(patterns) == 0 {
		return nil
	}

	return patterns
}

func effectiveLogLevel(logLevel string, progressEnabled bool) string {
	if progressEnabled {
		return "fatal"
	}

	return logLevel
}

func isFlagProvided(name string) bool {
	provided := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			provided = true
		}
	})
	return provided
}

func sanitizeMongoURI(raw string) string {
	parsed, err := url.Parse(raw)
	if err == nil && parsed.User != nil {
		if _, hasPassword := parsed.User.Password(); hasPassword {
			parsed.User = url.UserPassword(parsed.User.Username(), "xxxxx")
		}
		return parsed.String()
	}

	return redactMongoURICredentials(raw)
}

func redactMongoURICredentials(raw string) string {
	schemeSeparator := strings.Index(raw, "://")
	if schemeSeparator == -1 {
		return raw
	}

	credentialsStart := schemeSeparator + 3
	atIndex := strings.Index(raw[credentialsStart:], "@")
	if atIndex == -1 {
		return raw
	}
	atIndex += credentialsStart
	colonIndex := strings.Index(raw[credentialsStart:atIndex], ":")
	if colonIndex == -1 {
		return raw
	}
	colonIndex += credentialsStart

	return raw[:colonIndex+1] + "xxxxx" + raw[atIndex:]
}
