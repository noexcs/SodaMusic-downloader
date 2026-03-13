package main

import (
	"fmt"
	"io"
	"time"
)

// progressWriter implements io.Writer to track download progress
type progressWriter struct {
	total     int64
	written   int64
	startTime time.Time
	out       io.Writer
}

// Write implements io.Writer
func (pw *progressWriter) Write(p []byte) (n int, err error) {
	n, err = pw.out.Write(p)
	pw.written += int64(n)
	pw.printProgress()
	return
}

// printProgress displays the current download progress
func (pw *progressWriter) printProgress() {
	totalSize := pw.total
	downloaded := pw.written

	// Calculate progress percentage
	var progress float64
	if totalSize > 0 {
		progress = float64(downloaded) / float64(totalSize) * 100
	}

	// Calculate download speed
	eclapsed := time.Since(pw.startTime)
	speed := float64(downloaded) / eclapsed.Seconds()

	// Calculate ETA
	var eta string
	if totalSize > 0 && speed > 0 {
		remaining := float64(totalSize-downloaded) / speed
		eta = fmt.Sprintf("%.0fs", remaining)
	} else {
		eta = "--"
	}

	// Format sizes
	downloadedStr := formatSize(downloaded)
	totalSizeStr := formatSize(totalSize)
	speedStr := formatSize(int64(speed)) + "/s"

	// Clear line and print progress
	fmt.Printf("\r\033[K")
	fmt.Printf("Progress: %.1f%% | %s / %s | %s | ETA: %s",
		progress, downloadedStr, totalSizeStr, speedStr, eta)
}
