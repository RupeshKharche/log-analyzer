package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"
)

func main() {
	n := flag.Int("n", 100000, "number of log lines to generate")
	out := flag.String("out", "sample.log", "output file path")
	flag.Parse()

	f, err := os.Create(*out)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()

	levels := []string{"DEBUG", "INFO", "INFO", "INFO", "WARN", "ERROR", "FATAL"}
	comps := []string{"auth", "db", "api", "worker", "cache"}
	msgs := []string{
		"User login successful", "Connection timeout", "Database unavailable",
		"Invalid request", "Query completed", "Rate limited", "Token refreshed",
		"Job started", "Job finished", "Slow response",
	}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	base := time.Date(2026, 10, 20, 14, 32, 10, 0, time.UTC)

	for i := 0; i < *n; i++ {
		if i%500 == 0 && i > 0 {
			// inject a malformed line occasionally
			fmt.Fprintln(f, "this is a malformed line")
			continue
		}
		ts := base.Add(time.Duration(i) * time.Second)
		lvl := levels[rng.Intn(len(levels))]
		comp := comps[rng.Intn(len(comps))]
		msg := msgs[rng.Intn(len(msgs))]
		fmt.Fprintf(f, "%s %s [%s] %s\n", ts.Format("2006-01-02 15:04:05"), lvl, comp, msg)
	}
	fmt.Printf("wrote %d lines to %s\n", *n, *out)
}
