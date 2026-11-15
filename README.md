# Log Analyzer

A concurrent CLI log analyzer written in Go. It parses application log files, applies filters, and prints summary statistics (level counts, top errors, processing time, throughput).

## Features

- **Streaming reads** — processes files line by line (no full-file load)
- **Concurrent pipeline** — reader → worker pool → aggregator using channels
- **Configurable workers** — defaults to `runtime.NumCPU()`
- **Filtering** — by minimum log level, component, time range, and message text
- **Graceful cancellation** — handles `Ctrl+C` via context
- **Safe parsing** — malformed lines are counted and skipped
- **Text report** — human-readable summary to stdout or a file

## Log format

Each line must match:

```text
YYYY-MM-DD HH:MM:SS LEVEL [component] message
```

Example:

```text
2026-10-20 14:32:10 INFO [auth] User login successful
```

Supported levels: `DEBUG`, `INFO`, `WARN`, `ERROR`, `FATAL`.

## Requirements

- Go (see `go.mod` for the version used by this module)

## Build

```bash
go build -o loganalyzer ./cmd/loganalyzer
```

## Usage

```bash
loganalyzer <file> [file...] [flags]
```

Flags may appear before or after file arguments.

### Flags

| Flag | Description | Default |
| --- | --- | --- |
| `--workers` | Number of concurrent workers | CPU count |
| `--level` | Minimum severity (`DEBUG`, `INFO`, `WARN`, `ERROR`, `FATAL`, `ALL`) | `ALL` |
| `--component` | Exact component filter | `ALL` |
| `--since` | Start timestamp (`YYYY-MM-DD hh:mm:ss`) | none |
| `--until` | End timestamp (`YYYY-MM-DD hh:mm:ss`) | none |
| `--contains` | Substring to search in message (case-insensitive) | none |
| `--format` | Output format (`text`) | `text` |
| `--output` | Output file path | `stdout` |
| `--buffer` | Channel buffer size | `workers * 2` |

### Examples

```bash
# Analyze a single file
loganalyzer sample.log

# Multiple files with 8 workers
loganalyzer a.log b.log --workers 8

# Only errors and above from the auth component
loganalyzer app.log --level ERROR --component auth

# Filter by time range and message text
loganalyzer app.log --since "2026-10-20 14:30:00" --until "2026-10-20 15:00:00" --contains timeout

# Write report to a file
loganalyzer app.log --output report.txt
```

## Sample output

```text
Log Analysis Report
===================

Files analyzed:  1
Total lines:     250000
Parsed lines:    249500
Malformed lines: 500
Filtered lines:  0

Log Levels:
  DEBUG:  35700
  INFO:   107200
  WARN:   35600
  ERROR:  35500
  FATAL:  35500

Top Errors:
  1. Connection timeout  7100
  2. Database unavailable  7050

Duration:    1.2s
Throughput:  208333 lines/sec
Workers:     8
```

*(Values are illustrative.)*

## Generate sample logs

A small helper generates synthetic log files for testing:

```bash
go run ./cmd/genlog -n 250000 -out big.log
```

Flags:
- `-n` — number of lines (default: 100000)
- `-out` — output file path (default: `sample.log`)

## Architecture

```text
CLI/config ──► Reader (goroutine)
                  │  RawLine
                  ▼
            jobs channel (bounded)
                  │
        ┌─────────┼─────────┐
        ▼         ▼         ▼
     Worker     Worker    Worker   (parse + filter)
        └─────────┼─────────┘
                  ▼
           results channel
                  ▼
              Aggregator ──► AnalysisStats ──► Reporter
```

- **Reader** scans files and pushes `RawLine` values to `jobs`, then closes it.
- **Workers** parse/filter lines and emit `ParseResult` to `results`.
- **Aggregator** (single goroutine) updates stats to avoid shared mutable state.

## Project structure

```text
.
├── cmd/
│   ├── genlog/           # Sample log generator
│   └── loganalyzer/      # CLI entry point
├── internal/
│   ├── analyzer/         # Pipeline, workers, aggregator, filters, stats
│   ├── config/           # Runtime configuration
│   ├── parser/           # Log parsing and models
│   ├── reader/           # File reading into jobs channel
│   ├── reporter/         # Text report writer
│   └── utils/            # Helpers (e.g. top-N)
├── constants.go
├── sample.log
└── DESIGN.md
```

## Testing

```bash
go test ./...
```

With race detector:

```bash
go test -race ./...
```

## Notes

- `--level` filters by minimum severity (e.g. `WARN` includes `WARN`, `ERROR`, `FATAL`).
- Time buckets are aggregated per minute (`YYYY-MM-DD HH:MM`).
- Unreadable files are logged and skipped; malformed lines increment the malformed count.
