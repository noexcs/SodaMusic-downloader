package main

import (
	"context"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"strings"
	"time"
)

// fetchPage fetches the webpage content
func fetchPage(pageUrl string) (string, error) {
	fmt.Println(ColorBold, "Fetching page:", ColorReset, pageUrl)

	// Open the webpage with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", pageUrl, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error fetching page URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error: received non-200 status code %d", resp.StatusCode)
	}

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	// Parse HTML (for validation)
	_, err = html.Parse(strings.NewReader(string(bodyBytes)))
	if err != nil {
		return "", fmt.Errorf("error parsing HTML: %v", err)
	}

	return string(bodyBytes), nil
}

// formatSize formats bytes to human-readable size
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// formatDate formats Unix timestamp to human-readable date
func formatDate(timestamp int64) string {
	if timestamp <= 0 {
		return ""
	}
	t := time.Unix(timestamp, 0)
	return t.Format("2006-01-02")
}

// formatDuration formats seconds to human-readable duration
func formatDuration(seconds float64) string {
	if seconds <= 0 {
		return ""
	}
	minutes := int(seconds / 60)
	secs := int(seconds) % 60
	return fmt.Sprintf("%d:%02d", minutes, secs)
}

// sanitizeFilename replaces invalid Windows filename characters and handles edge cases
func sanitizeFilename(name string) string {
	// Replace invalid characters
	invalidChars := []struct {
		old, new string
	}{
		{"/", "_"},
		{"\\", "_"},
		{":", "-"},
		{"*", "-"},
		{"?", "-"},
		{"\"", "'"},
		{"<", "-"},
		{">", "-"},
		{"|", "-"},
	}

	for _, char := range invalidChars {
		name = strings.ReplaceAll(name, char.old, char.new)
	}

	// Remove leading/trailing dots and spaces (Windows reserved)
	name = strings.Trim(name, ". ")

	// Check for Windows reserved names
	reservedNames := []string{
		"CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
	}
	upperName := strings.ToUpper(name)
	for _, reserved := range reservedNames {
		if upperName == reserved || strings.HasPrefix(upperName, reserved+".") {
			name = "_" + name
			break
		}
	}

	// Limit filename length (Windows MAX_PATH is 260, leave room for path and extension)
	if len(name) > 200 {
		name = name[:200]
	}

	return name
}
