# mongocp

`mongocp` is a fast CLI tool for copying MongoDB databases and collections from one server to another.

It is designed for migration and replication workflows where you want filtering, dry-run safety, and clear progress visibility per worker.

## Features

- Copy databases and collections from source to destination MongoDB.
- Include/exclude databases with wildcard patterns (for example: `dev*`, `prod_*`).
- Exclude system databases (`admin`, `local`, `config`) by default.
- Parallel copy with configurable workers.
- Default progress UI with stable one-line-per-worker bars.
- Copies secondary indexes.
- Dry-run mode to preview what would be copied.

## Installation

### Install Latest

```bash
go install github.com/kstieger/mongocp/cmd/mongocp@latest
```

### Install A Specific Version

```bash
go install github.com/kstieger/mongocp/cmd/mongocp@v0.1.0
```

After installation, ensure your Go bin directory is in `PATH`.

## Usage

### Basic

```bash
mongocp -src "mongodb://root:xxxxx@source:27017/" -dst "mongodb://root:xxxxx@dest:27017/"
```

### Filtered Copy

```bash
mongocp \
  -src "mongodb://root:xxxxx@source:27017/" \
  -dst "mongodb://root:xxxxx@dest:27017/" \
  -include-dbs "dev_*" \
  -exclude-dbs "dev_tmp*" \
  -worker 8
```

### Dry Run

```bash
mongocp \
  -src "mongodb://root:xxxxx@source:27017/" \
  -dst "mongodb://root:xxxxx@dest:27017/" \
  -include-dbs "dev_*" \
  -dry-run
```

## Progress And Logging

- Progress mode is enabled by default.
- In default progress mode, regular logs are suppressed and only fatal failures are printed.
- Setting `-log-level` (or `-loglevel`) disables progress mode and enables normal logging.

## Flags

- `-src` (required): source MongoDB URI.
- `-dst` (required): destination MongoDB URI.
- `-include-dbs`: comma-separated include wildcard patterns.
- `-exclude-dbs`: comma-separated exclude wildcard patterns.
- `-exclude-system_dbs`: exclude `admin`, `local`, `config` (default: `true`).
- `-worker`: number of parallel workers (default: `10`).
- `-dry-run`: preview copy operations without writing data.
- `-log-level`: logging level (`info`, `debug`, `warn`, `error`); disables progress mode when set.
- `-loglevel`: alias for `-log-level`.

## Releases

GitHub Releases include prebuilt binaries for:

- Linux: `amd64`, `arm64`
- macOS (Darwin): `amd64`, `arm64`
- Windows: `amd64`, `arm64`

## Development

### Build

```bash
go build -o out/mongocp ./cmd/mongocp
```

### Test

```bash
go test ./...
```

### Lint

```bash
golangci-lint run
```

## Security Before Push

Before committing and pushing:

1. Run tests and lint:
   - `go test ./...`
   - `golangci-lint run`
2. Scan for secrets (for example with `gitleaks detect`).
3. Verify docs/examples do not include real credentials, tokens, or private keys.

## License

MIT. See [LICENSE](LICENSE).
