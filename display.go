package main

import (
	"fmt"
	"strings"
)

// Color codes for terminal output
const (
	ColorReset   = "\033[0m"
	ColorBold    = "\033[1m"
	ColorDim     = "\033[2m"
	ColorRed     = "\033[91m" // Bright red
	ColorCyan    = "\033[96m" // Bright cyan
	ColorGreen   = "\033[92m" // Bright green
	ColorYellow  = "\033[93m" // Bright yellow
	ColorBlue    = "\033[94m" // Bright blue
	ColorMagenta = "\033[95m" // Bright magenta
	ColorWhite   = "\033[97m" // Bright white
	ColorGray    = "\033[90m" // Bright black/gray
)

// displayTrackInfo displays track metadata information
func displayTrackInfo(trackName, artistName, albumName string, releaseDate int64,
	duration float64, genreTag string, composers, lyricists []string,
	coverURL string, hasCopyright bool) {

	fmt.Println("\n" + ColorBold + ColorCyan + "=== Track Information ===" + ColorReset)
	if trackName != "" {
		fmt.Printf(ColorBold+"Track Name:"+ColorReset+" %s\n", trackName)
	}
	if artistName != "" {
		fmt.Printf(ColorBold+"Artist:"+ColorReset+" %s\n", artistName)
	}
	if albumName != "" {
		fmt.Printf(ColorBold+"Album:"+ColorReset+" %s\n", albumName)
	}
	if releaseDate > 0 {
		fmt.Printf(ColorBold+"Release Date:"+ColorReset+" %s\n", formatDate(releaseDate))
	}
	if duration > 0 {
		fmt.Printf(ColorBold+"Duration:"+ColorReset+" %s\n", formatDuration(duration))
	}
	if genreTag != "" {
		fmt.Printf(ColorBold+"Genre:"+ColorReset+" %s\n", genreTag)
	}
	if len(composers) > 0 {
		fmt.Printf(ColorBold+"Composer(s):"+ColorReset+" %s\n", strings.Join(composers, ", "))
	}
	if len(lyricists) > 0 {
		fmt.Printf(ColorBold+"Lyricist(s):"+ColorReset+" %s\n", strings.Join(lyricists, ", "))
	}
	if coverURL != "" {
		fmt.Printf(ColorBold+"Cover URL:"+ColorReset+" %s\n", coverURL)
	}
	fmt.Printf(ColorBold+"Copyright Protected:"+ColorReset+" %v\n", hasCopyright)
}

// displayStats displays track statistics
func displayStats(stats map[string]interface{}) {
	if stats == nil {
		return
	}

	fmt.Println("\n" + ColorBold + ColorGreen + "=== Statistics ===" + ColorReset)
	if collected, ok := stats["count_collected"].(float64); ok {
		fmt.Printf("Collections:"+" %.0f\n", collected)
	}
	if comments, ok := stats["count_comment"].(float64); ok {
		fmt.Printf("Comments:"+" %.0f\n", comments)
	}
	if shared, ok := stats["count_shared"].(float64); ok {
		fmt.Printf("Shares:"+" %.0f\n", shared)
	}
}

// displayBitRates displays available audio bitrates
func displayBitRates(bitRates []interface{}) {
	if len(bitRates) == 0 {
		return
	}

	fmt.Println("\n" + ColorBold + ColorBlue + "=== Available Qualities ===" + ColorReset)
	for _, br := range bitRates {
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
				fmt.Printf("- %s (%d kbps)\n", quality, bitrate)
			}
		}
	}
}

// displayRelatedTracks displays related tracks list
func displayRelatedTracks(relatedTracks []interface{}) {
	if len(relatedTracks) == 0 {
		return
	}

	fmt.Println("\n" + ColorBold + ColorMagenta + "=== Related Tracks ===" + ColorReset)
	for i, track := range relatedTracks {
		if i >= 5 { // Limit to first 5
			break
		}
		if trackMap, ok := track.(map[string]interface{}); ok {
			if trackData, ok := trackMap["track"].(map[string]interface{}); ok {
				if name, ok := trackData["name"].(string); ok {
					artists := []interface{}{}
					if art, ok := trackData["artists"].([]interface{}); ok {
						artists = art
					}
					artistNames := ""
					for j, artist := range artists {
						if j > 0 {
							artistNames += ", "
						}
						if artistMap, ok := artist.(map[string]interface{}); ok {
							if aname, ok := artistMap["name"].(string); ok {
								artistNames += aname
							}
						}
					}
					fmt.Printf(ColorWhite+"%d. %s - %s"+ColorReset+"\n", i+1, name, artistNames)
				}
			}
		}
	}
}
