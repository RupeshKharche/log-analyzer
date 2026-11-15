package main

import (
	"context"
	"flag"
	"fmt"
	LogAnalyzer "loganalyzer/v2"
	"loganalyzer/v2/internal/analyzer"
	"loganalyzer/v2/internal/config"
	"loganalyzer/v2/internal/parser"
	"loganalyzer/v2/internal/reporter"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"time"
)

func printUsage() {
	fmt.Fprintln(os.Stderr, `Usage: loganalyzer <file> [file...] [flags]

Analyze one or more log files and print a summary.

Flags:`)
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr, `
Examples:
  loganalyzer app.log
  loganalyzer  a.log b.log --workers 8 --level ERROR --format json`)
}

// normalizeArgs moves flags before positional arguments so that the standard
// flag package (which stops parsing at the first non-flag) accepts flags in any
// position, e.g. "loganalyzer app.log --workers 8".
func normalizeArgs() {
	args := os.Args[1:]
	var flags, pos []string
	i := 0
	for i < len(args) {
		a := args[i]
		if strings.HasPrefix(a, "-") {
			flags = append(flags, a)
			// "--key value" form: consume the following value unless it is itself a flag.
			if !strings.Contains(a, "=") && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				flags = append(flags, args[i+1])
				i++
			}
		} else {
			pos = append(pos, a)
		}
		i++
	}
	os.Args = append([]string{os.Args[0]}, append(flags, pos...)...)
}

func buildFilters(flagCtx config.Context) []analyzer.Filter {
	var filters []analyzer.Filter
	filters = append(filters, analyzer.KeepAll)

	level := flagCtx.Level
	if len(level) != 0 && level != "ALL" {
		if minLevel, ok := parser.ParseLogLevel(level); ok {
			lf := analyzer.LevelFilter(minLevel)
			filters = append(filters, lf)
		} else {
			fmt.Fprintln(os.Stderr, `Level must be one of DEBUG, INFO, WARN, ERROR, FATAL`)
			os.Exit(1)
		}
	}

	comp := flagCtx.Component
	if comp != "ALL" {
		cf := analyzer.ComponentFilter(comp)
		filters = append(filters, cf)
	}

	since := flagCtx.Since
	if since != "" {
		parsedSince, sinceErr := time.Parse(LogAnalyzer.TimestampFormat, since)
		if sinceErr != nil {
			fmt.Fprintln(os.Stderr, `--since must be of the format YYYY-MM-DD hh:mm:ss`)
			os.Exit(1)
		}
		filters = append(filters, analyzer.SinceFilter(parsedSince))
	}

	until := flagCtx.Until
	if until != "" {
		parsedUntil, err := time.Parse(LogAnalyzer.TimestampFormat, until)
		if err != nil {
			fmt.Fprintln(os.Stderr, `--until must be of the format YYYY-MM-DD hh:mm:ss`)
			os.Exit(1)
		}
		filters = append(filters, analyzer.UntilFilter(parsedUntil))
	}

	contains := flagCtx.Contains
	if len(contains) != 0 {
		conf := analyzer.ContainsFilter(contains)
		filters = append(filters, conf)
	}

	return filters
}

func main() {
	workers := flag.Int("workers", runtime.NumCPU(), "Number of concurrent workers to use. Defaults to number of cpu cores")
	level := flag.String("level", "ALL", "Minimum severity to include in the report. Defaults to ALL")
	component := flag.String("component", "ALL", "Component filter. Defaults to ALL")
	since := flag.String("since", "", "Starting timestamp (YYYY-MM-DD hh:mm:ss)")
	until := flag.String("until", "", "Ending timestamp (YYYY-MM-DD hh:mm:ss)")
	contains := flag.String("contains", "", "Search message text")
	format := flag.String("format", "text", "Output format (text/json/csv). Defaults to text")
	output := flag.String("output", "stdout", "Output file path. Defaults to stdout")
	buffer := flag.Int("buffer", 0, "Channel buffer size. Defaults to 2 * workers")

	// Allow flags to appear after positional file arguments by reordering os.Args.
	normalizeArgs()

	flag.Parse()

	if *workers <= 0 {
		*workers = runtime.NumCPU()
	}

	if *buffer <= 0 {
		*buffer = *workers * 2
	}

	if flag.NArg() == 0 {
		printUsage()
		os.Exit(2)
	}

	files := flag.Args()
	if len(files) == 0 {
		printUsage()
		os.Exit(1)
	}

	flagCtx := config.NewContext(
		*workers,
		*level,
		*component,
		*since,
		*until,
		*contains,
		*format,
		*output,
		*buffer,
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	filters := buildFilters(*flagCtx)

	stats := analyzer.AnalyzeConcurrently(ctx, flagCtx, files, filters)

	// Resolve output destination.
	out := os.Stdout
	if flagCtx.Output != "" && flagCtx.Output != "stdout" {
		f, err := os.Create(flagCtx.Output)
		if err != nil {
			fmt.Fprintln(os.Stderr, "output file:", err)
			os.Exit(1)
		}
		defer f.Close()
		out = f
	}

	switch flagCtx.Format {
	case "text":
		if err := reporter.WriteText(out, files, stats, *workers); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		fmt.Fprintln(os.Stderr, "unknown format:", flagCtx.Format)
		os.Exit(2)
	}
}
