package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/atotto/clipboard"
)

// parseCommandLine parses command-line arguments and returns extracted values
func parseCommandLine() (string, bool, string) {
	downloadLyricsFlag := flag.Bool("lyrics", false, "Download lyrics file (.lrc) if available")
	outputDirFlag := flag.String("output", ".", "Output directory for downloaded files")
	flag.Parse()

	var input string

	// Check if command-line argument is provided
	if flag.NArg() >= 1 {
		input = flag.Arg(0)
	} else {
		// Try to read from clipboard if no command-line argument
		clipboardContent, err := clipboard.ReadAll()
		if err != nil {
			fmt.Println("Unable to access clipboard:", err)
			fmt.Println("Usage: go run main.go [options] <string_with_url>")
			fmt.Println("Options:")
			fmt.Println("  -lyrics    Download lyrics file (.lrc) if available")
			fmt.Println("  -output    Output directory for downloaded files (default: .)")
			return "", false, ""
		}

		if strings.TrimSpace(clipboardContent) == "" {
			fmt.Println("No input provided via command-line or clipboard.")
			fmt.Println("Usage: go run main.go [options] <string_with_url>")
			fmt.Println("Options:")
			fmt.Println("  -lyrics    Download lyrics file (.lrc) if available")
			fmt.Println("  -output    Output directory for downloaded files (default: .)")
			return "", false, ""
		}

		input = clipboardContent
		fmt.Println("Using URL from clipboard:", input)
	}

	// Extract http or https link
	re := regexp.MustCompile(`https?://[^\s"\']+`)
	pageUrl := re.FindString(input)
	if pageUrl == "" {
		fmt.Println("No http/https URL found in the input.")
		return "", false, ""
	}

	return pageUrl, *downloadLyricsFlag, *outputDirFlag
}

// prepareOutput prepares the output file path
func prepareOutput(trackName, artistName, outputDir string) string {
	// Extract filename from URL or use track name
	fileName := "downloaded_audio.m4a"

	// Use track name if available
	if trackName != "" {
		// Clean names for use as filename
		trackName = sanitizeFilename(trackName)
		artistName = sanitizeFilename(artistName)

		// Remove trailing dots to avoid double extensions
		trackName = strings.TrimRight(trackName, ".")
		artistName = strings.TrimRight(artistName, ".")

		fileName = trackName + " - " + artistName + ".m4a"
	}

	// Add output directory to file path
	fileName = filepath.Join(outputDir, fileName)

	return fileName
}

// downloadFile downloads the audio file with progress tracking
func downloadFile(audioSrc, fileName string) (int64, error) {
	// Download the file with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	fileReq, err := http.NewRequestWithContext(ctx, "GET", audioSrc, nil)
	if err != nil {
		return 0, fmt.Errorf("error creating download request: %v", err)
	}

	fileResp, err := http.DefaultClient.Do(fileReq)
	if err != nil {
		return 0, fmt.Errorf("error downloading audio file: %v", err)
	}
	defer fileResp.Body.Close()

	if fileResp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("error: received non-200 status code %d for audio file", fileResp.StatusCode)
	}

	// Create file
	out, err := os.Create(fileName)
	if err != nil {
		return 0, fmt.Errorf("error creating file: %v", err)
	}

	// Get content length for progress tracking
	totalSize := fileResp.ContentLength

	// Create progress writer
	pw := &progressWriter{
		total:     totalSize,
		startTime: time.Now(),
		out:       out,
	}

	// Save the file with progress tracking
	written, err := io.Copy(pw, fileResp.Body)
	if err != nil {
		out.Close()
		return 0, fmt.Errorf("error saving file: %v", err)
	}

	// Explicitly close the file to release the lock before embedding metadata
	err = out.Close()
	if err != nil {
		return written, fmt.Errorf("error closing file: %v", err)
	}

	// Print final download complete message
	fmt.Println(ColorGreen, "\n✓ Download complete (", formatSize(written), "): ", fileName, ColorReset)

	return written, nil
}

func downloadLyricsFile(lyrics map[string]interface{}, baseName string, outputDir string) {
	if sentences, ok := lyrics["sentences"].([]interface{}); ok && len(sentences) > 0 {
		// Create LRC file name
		lrcFileName := baseName + ".lrc"
		// Add output directory to file path
		lrcFileName = filepath.Join(outputDir, lrcFileName)

		fmt.Println("Downloading lyrics to:", lrcFileName)

		// Build LRC content
		var lrcContent strings.Builder
		for _, sentence := range sentences {
			if sentMap, ok := sentence.(map[string]interface{}); ok {
				startTimeMs := 0.0
				text := ""

				if startMs, ok := sentMap["startMs"].(float64); ok {
					startTimeMs = startMs
				}
				if t, ok := sentMap["text"].(string); ok {
					text = t
				}

				// Convert milliseconds to MM:SS.xx format
				totalSeconds := startTimeMs / 1000.0
				minutes := int(totalSeconds / 60)
				seconds := totalSeconds - float64(minutes*60)
				timeStamp := fmt.Sprintf("[%02d:%05.2f]", minutes, seconds)

				lrcContent.WriteString(timeStamp)
				lrcContent.WriteString(text)
				lrcContent.WriteString("\n")
			}
		}

		// Write LRC file
		err := os.WriteFile(lrcFileName, []byte(lrcContent.String()), 0644)
		if err != nil {
			fmt.Printf("Error saving lyrics file: %v\n", err)
		} else {
			fmt.Printf("Lyrics saved to: %s\n", lrcFileName)
		}
	}
}

func main() {
	// Parse command line arguments
	pageUrl, downloadLyricsFlag, outputDir := parseCommandLine()
	if pageUrl == "" {
		return
	}

	// Fetch webpage content
	bodyStr, err := fetchPage(pageUrl)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Extract metadata
	audioSrc, trackName, artistName, coverURL, genreTag, duration, hasCopyright, lyrics, albumName, releaseDate, stats, bitRates, relatedTracks, composers, lyricists, err := extractMetadata(bodyStr)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println(ColorBold, "Found audio file:", ColorReset, audioSrc)

	// Display track metadata
	displayTrackInfo(trackName, artistName, albumName, releaseDate,
		duration, genreTag, composers, lyricists, coverURL, hasCopyright)

	// Display stats if available
	displayStats(stats)

	// Display available bitrates
	displayBitRates(bitRates)

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		return
	}

	// Prepare output file path
	fileName := prepareOutput(trackName, artistName, outputDir)

	// Display related tracks
	displayRelatedTracks(relatedTracks)
	fmt.Println()

	fmt.Printf("Downloading to: %s\n", fileName)

	// Download lyrics if available and flag is set
	if downloadLyricsFlag && lyrics != nil && trackName != "" {
		// Use fileName without extension for lyrics
		lyricsBaseName := strings.TrimSuffix(filepath.Base(fileName), ".m4a")
		downloadLyricsFile(lyrics, lyricsBaseName, outputDir)
	}

	// Download the audio file
	_, err = downloadFile(audioSrc, fileName)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Embed cover art and metadata if available
	if coverURL != "" || trackName != "" || artistName != "" || albumName != "" || genreTag != "" || len(composers) > 0 || len(lyricists) > 0 || releaseDate > 0 || hasCopyright {
		embedMetadata(fileName, trackName, artistName, albumName, coverURL, genreTag, composers, lyricists, releaseDate, hasCopyright, stats, bitRates)
	}
}
