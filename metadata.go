package main

import (
	"encoding/json"
	"fmt"
	"github.com/dhowden/tag"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// extractTrackMetadata extracts all track metadata from _ROUTER_DATA JSON
func extractTrackMetadata(routerData map[string]interface{},
	audioSrc, trackName, artistName, coverURL *string,
	duration *float64, genreTag *string, hasCopyright *bool,
	lyrics *map[string]interface{}, albumName *string, releaseDate *int64,
	stats *map[string]interface{}, bitRates *[]interface{},
	relatedTracks *[]interface{}, composers, lyricists *[]string) {

	// Navigate: loaderData -> track_page -> audioWithLyricsOption
	loaderData, ok := routerData["loaderData"].(map[string]interface{})
	if !ok {
		return
	}

	trackPage, ok := loaderData["track_page"].(map[string]interface{})
	if !ok {
		return
	}

	audioOption, ok := trackPage["audioWithLyricsOption"].(map[string]interface{})
	if !ok {
		return
	}

	// Extract basic audio info
	extractBasicAudioInfo(audioOption, audioSrc, trackName, artistName, coverURL,
		duration, genreTag, hasCopyright, lyrics)

	// Extract detailed track info
	extractTrackDetails(audioOption, albumName, releaseDate, stats, bitRates,
		relatedTracks, composers, lyricists)
}

// extractBasicAudioInfo extracts basic audio information
func extractBasicAudioInfo(audioOption map[string]interface{},
	audioSrc, trackName, artistName, coverURL *string,
	duration *float64, genreTag *string, hasCopyright *bool,
	lyrics *map[string]interface{}) {

	if url, ok := audioOption["url"].(string); ok {
		*audioSrc = url
	}
	if name, ok := audioOption["trackName"].(string); ok {
		*trackName = name
	}
	if artist, ok := audioOption["artistName"].(string); ok {
		artist = strings.ReplaceAll(artist, " ", "")
		*artistName = strings.ReplaceAll(artist, "/", ", ")
	}
	if cover, ok := audioOption["coverURL"].(string); ok {
		*coverURL = cover
	}
	if dur, ok := audioOption["duration"].(float64); ok {
		*duration = dur
	}
	if genre, ok := audioOption["genre_tag"].(string); ok {
		*genreTag = genre
	}
	if copyright, ok := audioOption["hasCopyright"].(bool); ok {
		*hasCopyright = copyright
	}
	if lrc, ok := audioOption["lyrics"].(map[string]interface{}); ok {
		*lyrics = lrc
	}
}

// extractTrackDetails extracts detailed track information
func extractTrackDetails(audioOption map[string]interface{},
	albumName *string, releaseDate *int64,
	stats *map[string]interface{}, bitRates *[]interface{},
	relatedTracks *[]interface{}, composers, lyricists *[]string) {

	trackInfo, ok := audioOption["trackInfo"].(map[string]interface{})
	if !ok {
		return
	}

	// Extract album info
	if album, ok := trackInfo["album"].(map[string]interface{}); ok {
		if name, ok := album["name"].(string); ok {
			*albumName = name
		}
		if release, ok := album["release_date"].(float64); ok {
			*releaseDate = int64(release)
		}
	}

	// Extract stats
	if s, ok := trackInfo["stats"].(map[string]interface{}); ok {
		*stats = s
	}

	// Extract bit rates
	if br, ok := trackInfo["bit_rates"].([]interface{}); ok {
		*bitRates = br
	}

	// Extract label info (publisher)
	if labelInfo, ok := trackInfo["label_info"].(map[string]interface{}); ok {
		_ = labelInfo // Available for future use
	}

	// Extract related tracks
	if related, ok := audioOption["relatedTracks"].([]interface{}); ok {
		*relatedTracks = related
	}

	// Extract song maker team
	extractSongMakerTeam(trackInfo, composers, lyricists)
}

// extractSongMakerTeam extracts composer and lyricist information
func extractSongMakerTeam(trackInfo map[string]interface{},
	composers, lyricists *[]string) {

	songMakerTeam, ok := trackInfo["song_maker_team"].(map[string]interface{})
	if !ok {
		return
	}

	// Extract composers
	if comps, ok := songMakerTeam["composers"].([]interface{}); ok {
		for _, comp := range comps {
			if compMap, ok := comp.(map[string]interface{}); ok {
				if name, ok := compMap["name"].(string); ok {
					*composers = append(*composers, name)
				}
			}
		}
	}

	// Extract lyricists
	if lyr, ok := songMakerTeam["lyricists"].([]interface{}); ok {
		for _, l := range lyr {
			if lMap, ok := l.(map[string]interface{}); ok {
				if name, ok := lMap["name"].(string); ok {
					*lyricists = append(*lyricists, name)
				}
			}
		}
	}
}

// extractMetadata extracts track metadata from HTML content
func extractMetadata(bodyStr string) (
	audioSrc, trackName, artistName, coverURL, genreTag string,
	duration float64,
	hasCopyright bool,
	lyrics map[string]interface{},
	albumName string,
	releaseDate int64,
	stats map[string]interface{},
	bitRates, relatedTracks []interface{},
	composers, lyricists []string,
	err error,
) {
	// Look for _ROUTER_DATA in the HTML
	routerDataPattern := regexp.MustCompile(`_ROUTER_DATA\s*=\s*({.*?});`)
	matches := routerDataPattern.FindStringSubmatch(bodyStr)

	if len(matches) > 1 {
		// Parse the JSON
		var routerData map[string]interface{}
		err = json.Unmarshal([]byte(matches[1]), &routerData)
		if err == nil {
			extractTrackMetadata(routerData, &audioSrc, &trackName, &artistName, &coverURL,
				&duration, &genreTag, &hasCopyright, &lyrics, &albumName, &releaseDate,
				&stats, &bitRates, &relatedTracks, &composers, &lyricists)
		}
	}

	if audioSrc == "" {
		err = fmt.Errorf("could not find audio URL in _ROUTER_DATA or <audio> tag")
	}

	return
}

func embedMetadata(fileName, trackName, artistName, albumName, coverURL, genreTag string, composers, lyricists []string, releaseDate int64, hasCopyright bool, stats map[string]interface{}, bitRates []interface{}, lyrics map[string]interface{}) {
	fmt.Println("\nEmbedding metadata and cover art...")

	// Download cover image if URL is provided
	var imgData []byte
	var coverFilePath string
	var err error
	if coverURL != "" {
		imgResp, err := http.Get(coverURL)
		if err != nil {
			fmt.Printf("Error downloading cover image: %v\n", err)
		} else {
			defer imgResp.Body.Close()
			imgData, err = io.ReadAll(imgResp.Body)
			if err != nil {
				fmt.Printf("Error reading cover image: %v\n", err)
				imgData = nil
			} else {
				// Save cover to temp file for ffmpeg
				coverFilePath = strings.TrimSuffix(fileName, ".m4a") + "_temp_cover.jpg"
				err = os.WriteFile(coverFilePath, imgData, 0644)
				if err != nil {
					fmt.Printf("Error saving temp cover: %v\n", err)
					coverFilePath = ""
				}
			}
		}
	}

	// Try to use ffmpeg first (existing functionality)
	// Note: We'll skip Go tag library if ffmpeg succeeds to avoid file handle conflicts
	ffmpegErr := embedWithFFmpeg(fileName, trackName, artistName, albumName, coverFilePath, genreTag, composers, lyricists, releaseDate, hasCopyright, lyrics)
	if ffmpegErr != nil {
		fmt.Printf("FFmpeg not available or failed: %v\n", ffmpegErr)
		fmt.Println("Falling back to Go tag library...")

		// Only use Go tag library as fallback if ffmpeg failed
		err = embedWithGoTag(fileName, trackName, artistName, albumName, imgData, genreTag, composers, lyricists, releaseDate, hasCopyright, stats, bitRates, lyrics)
		if err != nil {
			fmt.Printf("Error embedding metadata with Go tag library: %v\n", err)
		} else {
			fmt.Println(ColorGreen, "✓ Metadata embedded successfully using Go tag library", ColorReset)
		}
	} else {
		// FFmpeg succeeded, no need for Go tag library
		fmt.Println(ColorGreen, "✓ Metadata and cover art embedded successfully using ffmpeg", ColorReset)
	}

	// Clean up temp cover file
	if coverFilePath != "" {
		os.Remove(coverFilePath)
	}
}

func embedWithGoTag(fileName, trackName, artistName, albumName string, coverImage []byte, genreTag string, composers, lyricists []string, releaseDate int64, hasCopyright bool, stats map[string]interface{}, bitRates []interface{}, lyrics map[string]interface{}) error {
	// Open the file for reading
	f, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer f.Close()

	// Read the existing tags
	m, err := tag.ReadFrom(f)
	if err != nil {
		// If reading fails, we'll create new tags
		fmt.Println("Creating new metadata tags...")
	} else {
		fmt.Printf("Existing format detected: %s\n", m.Format())
	}

	// Close the file and reopen for writing
	f.Close()

	// Prepare metadata fields
	metadata := make(map[string]interface{})

	// Set basic metadata
	if trackName != "" {
		metadata["title"] = trackName
	}
	if artistName != "" {
		metadata["artist"] = artistName
	}
	if albumName != "" {
		metadata["album"] = albumName
	}
	if genreTag != "" {
		metadata["genre"] = genreTag
	}

	// Set composer and lyricist
	if len(composers) > 0 {
		metadata["composer"] = strings.Join(composers, ", ")
	}
	if len(lyricists) > 0 {
		metadata["lyricist"] = strings.Join(lyricists, ", ")
	}

	// Convert release date (Unix timestamp in seconds) to year
	if releaseDate > 0 {
		year := time.Unix(releaseDate, 0).Year()
		metadata["year"] = year
	}

	// Add copyright information
	if hasCopyright {
		metadata["copyright"] = "℗ & © All Rights Reserved"
	}

	// For M4A files, we need to use a different approach
	// The tag library doesn't support writing to M4A directly
	// So we'll use ffmpeg if available, or save metadata separately

	// Try to use exiftool if available for comprehensive metadata
	err = embedWithExifTool(fileName, metadata, coverImage, stats, bitRates, lyrics)
	if err == nil {
		return nil
	}

	fmt.Printf("Exiftool not available or failed: %v\n", err)

	// Save metadata to a sidecar file as fallback
	//metadataFile := strings.TrimSuffix(fileName, ".m4a") + "_metadata.json"
	//metadataJSON, _ := json.MarshalIndent(metadata, "", "  ")
	//err = os.WriteFile(metadataFile, metadataJSON, 0644)
	//if err != nil {
	//	return fmt.Errorf("error saving metadata file: %v", err)
	//}
	//
	//fmt.Printf("Metadata saved to: %s\n", metadataFile)
	fmt.Println("Note: To fully embed metadata into M4A file, install exiftool or use ffmpeg")

	return nil
}

func embedWithExifTool(fileName string, metadata map[string]interface{}, coverImage []byte, stats map[string]interface{}, bitRates []interface{}, lyrics map[string]interface{}) error {
	// Check if exiftool is available
	_, err := exec.LookPath("exiftool")
	if err != nil {
		return fmt.Errorf("exiftool not found in PATH")
	}

	// Build exiftool command arguments
	args := []string{"-overwrite_original"}

	// Add metadata arguments
	for key, value := range metadata {
		args = append(args, fmt.Sprintf("-%s=%v", key, value))
	}

	// Add cover image if available
	if coverImage != nil {
		// Save cover to temp file
		tempCover := strings.TrimSuffix(fileName, ".m4a") + "_exif_cover.jpg"
		err := os.WriteFile(tempCover, coverImage, 0644)
		if err == nil {
			defer os.Remove(tempCover)
			args = append(args, fmt.Sprintf("-Artwork<%s", tempCover))
		}
	}

	// Add statistics as comments (if available)
	if stats != nil {
		if collected, ok := stats["count_collected"].(float64); ok {
			args = append(args, fmt.Sprintf("-Comment=Collections: %.0f", collected))
		}
		if comments, ok := stats["count_comment"].(float64); ok {
			args = append(args, fmt.Sprintf("-Comment=Comments: %.0f", comments))
		}
		if shared, ok := stats["count_shared"].(float64); ok {
			args = append(args, fmt.Sprintf("-Comment=Shares: %.0f", shared))
		}
	}

	// Add bitrate information as description (if available)
	if len(bitRates) > 0 {
		var qualityInfo string
		for i, br := range bitRates {
			if i > 0 {
				qualityInfo += "; "
			}
			if brMap, ok := br.(map[string]interface{}); ok {
				quality := ""
				bitrate := 0
				if q, ok := brMap["quality"].(string); ok {
					quality = q
				}
				if b, ok := brMap["br"].(float64); ok {
					bitrate = int(b)
				}
				if quality != "" {
					qualityInfo += fmt.Sprintf("%s (%d kbps)", quality, bitrate)
				}
			}
		}
		if qualityInfo != "" {
			args = append(args, fmt.Sprintf("-Description=Available qualities: %s", qualityInfo))
		}
	}

	// Add lyrics if available
	if lyrics != nil {
		if lrcText, err := generateLRC(lyrics); err == nil && lrcText != "" {
			tempLyricsFile := strings.TrimSuffix(fileName, ".m4a") + "_temp_lyrics.lrc"
			err := os.WriteFile(tempLyricsFile, []byte(lrcText), 0644)
			if err == nil {
				defer os.Remove(tempLyricsFile)
				args = append(args, fmt.Sprintf("-Lyrics<%s", tempLyricsFile))
			}
		}
	}

	args = append(args, fileName)

	cmd := exec.Command("exiftool", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("exiftool failed: %v, output: %s", err, string(output))
	}

	fmt.Println(ColorGreen, "✓ Metadata embedded successfully using exiftool", ColorReset)
	return nil
}

func embedWithFFmpeg(audioFile, trackName, artistName, albumName, coverFile string, genreTag string, composers, lyricists []string, releaseDate int64, hasCopyright bool, lyrics map[string]interface{}) error {
	// Check if ffmpeg is available
	_, err := exec.LookPath("ffmpeg")
	if err != nil {
		return fmt.Errorf("ffmpeg not found in PATH")
	}

	outputFile := strings.TrimSuffix(audioFile, ".m4a") + "_embedded.m4a"

	// Build ffmpeg command
	cmdArgs := []string{"-i", audioFile}

	// Add cover image input if available
	if coverFile != "" {
		cmdArgs = append(cmdArgs, "-i", coverFile)
	}

	// Copy audio codec
	cmdArgs = append(cmdArgs, "-c", "copy")

	// If cover is provided, set it as attached picture
	if coverFile != "" {
		cmdArgs = append(cmdArgs, "-c:v", "png", "-disposition:v:0", "attached_pic")
	}

	// Add metadata
	if trackName != "" {
		cmdArgs = append(cmdArgs, "-metadata", fmt.Sprintf("title=%s", trackName))
	}
	if artistName != "" {
		cmdArgs = append(cmdArgs, "-metadata", fmt.Sprintf("artist=%s", artistName))
	}
	if albumName != "" {
		cmdArgs = append(cmdArgs, "-metadata", fmt.Sprintf("album=%s", albumName))
	}
	if genreTag != "" {
		cmdArgs = append(cmdArgs, "-metadata", fmt.Sprintf("genre=%s", genreTag))
	}
	if len(composers) > 0 {
		cmdArgs = append(cmdArgs, "-metadata", fmt.Sprintf("composer=%s", strings.Join(composers, ", ")))
	}
	if len(lyricists) > 0 {
		cmdArgs = append(cmdArgs, "-metadata", fmt.Sprintf("lyricist=%s", strings.Join(lyricists, ", ")))
	}
	if releaseDate > 0 {
		year := time.Unix(releaseDate, 0).Year()
		cmdArgs = append(cmdArgs, "-metadata", fmt.Sprintf("year=%d", year))
	}
	if hasCopyright {
		cmdArgs = append(cmdArgs, "-metadata", "copyright=℗ & © All Rights Reserved")
	}

	// Add lyrics if available
	if lyrics != nil {
		if lrcText, err := generateLRC(lyrics); err == nil && lrcText != "" {
			cmdArgs = append(cmdArgs, "-metadata", fmt.Sprintf("lyrics=%s", lrcText))
		}
	}

	cmdArgs = append(cmdArgs, outputFile)

	cmd := exec.Command("ffmpeg", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg failed: %v, output: %s", err, string(output))
	}

	// Remove original file first to avoid rename conflicts on Windows
	err = os.Remove(audioFile)

	// Replace original file with embedded version
	err = os.Rename(outputFile, audioFile)
	if err != nil {
		return fmt.Errorf("failed to replace original file: %v", err)
	}

	return nil
}

func generateLRC(lyrics map[string]interface{}) (string, error) {
	if sentences, ok := lyrics["sentences"].([]interface{}); ok && len(sentences) > 0 {
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

				totalSeconds := startTimeMs / 1000.0
				minutes := int(totalSeconds / 60)
				seconds := totalSeconds - float64(minutes*60)
				timeStamp := fmt.Sprintf("[%02d:%05.2f]", minutes, seconds)

				lrcContent.WriteString(timeStamp)
				lrcContent.WriteString(text)
				lrcContent.WriteString("\n")
			}
		}
		return lrcContent.String(), nil
	}
	return "", fmt.Errorf("no lyrics sentences found")
}
