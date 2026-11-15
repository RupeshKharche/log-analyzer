package reader

import (
	"bufio"
	"context"
	"log"
	"loganalyzer/v2/internal/parser"
	"os"
)

func readOneFile(ctx context.Context, path string, jobs chan<- parser.RawLine) {
	// Try to open the file
	file, fileErr := os.Open(path)
	if fileErr != nil {
		log.Printf("failed to read file %s: %s\n", path, fileErr)
		return
	}

	// Close file on exiting function
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Printf("failed to close file %s", path)
		}
	}(file)

	lineNumber := 0

	// Create scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	// Increase the max size of the scanner buffer to read very long log lines
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024) // 1 MB buffer

	// Iterate on each line of the file
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		rawLine := parser.NewRawLine(path, lineNumber, line)

		select {
		case jobs <- *rawLine:
		case <-ctx.Done():
			return
		}
	}

	// Check errors occurred during the scan
	if scanErr := scanner.Err(); scanErr != nil {
		log.Printf("error reading file[%s]: %s\n", path, scanErr)
	}
}

// ReadFiles reads the given files line by line and pushes a RawLine to the given channel
// for further processing
func ReadFiles(ctx context.Context, paths []string, c chan<- parser.RawLine) {
	defer close(c)

	for _, path := range paths {
		readOneFile(ctx, path, c)
	}
}
