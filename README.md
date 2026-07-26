# mongocp

[![CI](https://github.com/kstieger/mongocp/workflows/CI/badge.svg)](https://github.com/kstieger/mongocp/actions)
[![golangci-lint](https://img.shields.io/badge/linted%20with-golangci--lint-brightgreen.svg)](https://golangci-lint.run/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**mongocp** is a fast CLI tool for copying MongoDB databases and collections from one server to another, built for migration and replication workflows where you want filtering, dry-run safety, and clear per-worker progress visibility.

## ✨ Key Features

- 🗄️ **Database & Collection Copy**: Copies databases and collections, including secondary indexes, from a source MongoDB server to a destination
- 🎯 **Wildcard Filtering**: Include/exclude databases with shell-style glob patterns (e.g. `dev*`, `prod_*`)
- 🛡️ **System DB Protection**: Excludes `admin`, `local`, and `config` by default
- ⚡ **Parallel Workers**: Configurable worker pool for concurrent collection copies
- 📊 **Live Progress UI**: Stable one-line-per-worker progress bars by default, with percent and document counts
- 🧪 **Dry-Run Mode**: Preview exactly what would be copied without writing any data

## 🚀 Installation

### Prerequisites

- Go 1.26 or higher

### Install from Source

```bash
go install github.com/kstieger/mongocp/cmd/mongocp@latest
```

Or pin a specific version:

```bash
go install github.com/kstieger/mongocp/cmd/mongocp@v0.1.0
```

This will install the `mongocp` command in your `GOPATH/bin` folder — make sure it's on your `PATH`.

## 📖 Usage

### Basic Usage

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

### Flags

| Flag | Description | Default |
| --- | --- | --- |
| `-src` | Source MongoDB URI (required) | — |
| `-dst` | Destination MongoDB URI (required) | — |
| `-include-dbs` | Comma-separated include wildcard patterns | *(all)* |
| `-exclude-dbs` | Comma-separated exclude wildcard patterns | *(none)* |
| `-exclude-system_dbs` | Exclude `admin`, `local`, `config` | `true` |
| `-worker` | Number of parallel workers | `10` |
| `-dry-run` | Preview copy operations without writing data | `false` |
| `-log-level` | Logging level (`info`, `debug`, `warn`, `error`); disables progress mode when set | `info` |
| `-loglevel` | Alias for `-log-level` | `info` |
| `-version` | Print version information and exit | `false` |

## 🎯 Examples

### Progress UI

Progress mode is on by default: regular logs are suppressed and each worker gets a live, color-filled bar showing percent complete and document counts:

```
mydb.orders             ▕████████████████████░░░░░░░░░░░░░░░░░░▏  52% ( 52000/100000)
mydb.users              ▕████████████████████████████████████▏  100% ( 12000/ 12000)
mydb.sessions           ▕██████░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░▏   16% (  1600/ 10000)
```

Set `-log-level` (or `-loglevel`) to disable progress mode and get normal structured logs instead — useful for CI or piping to a log aggregator.

### Dry Run Output

With `-dry-run`, `mongocp` lists exactly which databases and collections it would copy — including index counts — without connecting to the destination for writes.

## 📦 Releases

Pushing a `v*` tag (via `task release`) triggers the `Release` GitHub Actions workflow, which cross-compiles and publishes prebuilt binaries as GitHub Release assets for:

- Linux: `amd64`, `arm64`
- macOS (Darwin): `amd64`, `arm64`
- Windows: `amd64`, `arm64`

## 🤝 Contributing

Contributions are welcome! To get started:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

### Local Tasks

This project uses [Task](https://taskfile.dev/) to run common developer commands:

```bash
task build            # cross-compile for linux/darwin/windows (amd64+arm64)
task format           # gofmt + go mod tidy
task lint             # golangci-lint
task vulncheck        # govulncheck
task secretleakcheck  # gitleaks secret scan
task test             # go test -race with coverage
task pre-checkin      # format, lint, vulncheck, secretleakcheck, and test
task release          # tag and push a release; CI builds the binaries and publishes the GitHub Release
```

### Security Before Push

Before committing and pushing:

1. Run `task pre-checkin` (formatting, linting, vulnerability scanning, secret scanning, and tests).
2. Verify docs/examples do not include real credentials, tokens, or private keys.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
