package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Video represents a single video from the channel
type Video struct {
	ID             string  `json:"id"`
	Title          string  `json:"title"`
	Description    string  `json:"description"`
	Duration       float64 `json:"duration"`
	Thumbnail      string  `json:"thumbnail"`
	ThumbnailSmall string  `json:"thumbnailSmall"`
	Position       int64   `json:"position"`
	Section        string  `json:"section"`
	Index          int     `json:"index"`
	AspectRatio    float64 `json:"aspectRatio"`
}

// ChannelData represents the decoded channel data
type ChannelData struct {
	HashedID  string `json:"hashedId"`
	NumericID int    `json:"numericId"`
	Series    []struct {
		Sections []struct {
			Name     string `json:"name"`
			Episodes []struct {
				HashedID           string  `json:"hashedId"`
				Name               string  `json:"name"`
				EpisodeTitle       string  `json:"episodeTitle"`
				EpisodeDescription string  `json:"episodeDescription"`
				DurationInSeconds  float64 `json:"durationInSeconds"`
				StillURL           string  `json:"stillUrl"`
				ThumbnailURL       string  `json:"thumbnailUrl"`
				Position           int64   `json:"position"`
				Index              int     `json:"index"`
				AspectRatio        float64 `json:"aspectRatio"`
			} `json:"episodes"`
		} `json:"sections"`
	} `json:"series"`
}

func main() {
	videoID := flag.String("id", "", "Wistia video ID (e.g. j4n8x2m7vw)")
	pageURL := flag.String("url", "", "Main Wistia page URL (e.g. https://example.wistia.com/medias/h3b2k9f5xp)")
	clipboardHTML := flag.String("clipboard", "", "HTML snippet from 'Copy link' (contains wvideo parameter)")
	output := flag.String("o", "video.mp4", "Output filename (ignored for channel downloads)")
	flag.Parse()

	var id string

	if *videoID != "" {
		id = *videoID
	} else if *clipboardHTML != "" {
		id = extractVideoIDFromHTML(*clipboardHTML)
		if id == "" {
			fmt.Println("Could not find Wistia video ID in HTML snippet.")
			os.Exit(1)
		}
		fmt.Println("Found video ID from HTML snippet:", id)
	} else if *pageURL != "" {
		// Check if this is a channel page with a specific video ID
		if isChannelURL(*pageURL) {
			// Check if URL also contains wmediaid (specific video in channel)
			videoIDFromURL := extractVideoIDFromURL(*pageURL)
			if videoIDFromURL != "" {
				fmt.Println("Detected Wistia channel page with specific video!")
				fmt.Printf("Found channel and video ID: %s\n", videoIDFromURL)

				// Ask user what they want to download
				choice := askUserChoice(videoIDFromURL)
				if choice == "video" {
					fmt.Println("Downloading single video...")
					id = videoIDFromURL
				} else {
					fmt.Println("Downloading entire channel...")
					handleChannelDownload(*pageURL, *output)
					return
				}
			} else {
				fmt.Println("Detected Wistia channel page!")
				handleChannelDownload(*pageURL, *output)
				return
			}
		} else {
			// First check if the URL contains wmediaid parameter
			id = extractVideoIDFromURL(*pageURL)
			if id != "" {
				fmt.Println("Found video ID from URL parameter:", id)
			} else {
				// If no wmediaid found, try extracting from page content
				id = extractVideoIDFromPage(*pageURL)
				if id == "" {
					fmt.Println("Could not find Wistia video ID in page.")
					os.Exit(1)
				}
				fmt.Println("Found video ID:", id)
			}
		}
	} else {
		fmt.Println("Usage: wistia-downloader -id <videoID> OR -url <WistiaPageURL> OR -clipboard <HTMLSnippet> [-o <output.mp4>]")
		fmt.Println("  For channel pages: -url <WistiaChannelURL> (will download all videos)")
		os.Exit(1)
	}

	// Download single video
	finalOutput := *output
	if *output == "video.mp4" { // Check if using default filename
		// Get video title and create a better filename
		videoTitle := getVideoTitle(id)
		if videoTitle != "" {
			finalOutput = createSafeFilename(videoTitle) + ".mp4"
			fmt.Printf("Using video title as filename: %s\n", finalOutput)
		} else {
			finalOutput = id + ".mp4"
			fmt.Printf("Using video ID as filename: %s\n", finalOutput)
		}
	}
	downloadSingleVideo(id, finalOutput)
}

// isChannelURL checks if the URL is a Wistia channel page
func isChannelURL(pageURL string) bool {
	return strings.Contains(pageURL, "/embed/channel/") || strings.Contains(pageURL, "wchannelid=")
}

// askUserChoice prompts user to choose between downloading single video or entire channel
func askUserChoice(videoID string) string {
	// Fetch video title for better user experience
	videoTitle := getVideoTitle(videoID)
	if videoTitle == "" {
		videoTitle = videoID // Fallback to ID if title fetch fails
	}

	fmt.Println("\nWhat would you like to download?")
	fmt.Printf("1) Just this video: \"%s\" (ID: %s)\n", videoTitle, videoID)
	fmt.Println("2) Entire channel (all videos)")
	fmt.Print("Enter your choice (1 or 2): ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		os.Exit(1)
	}

	choice := strings.TrimSpace(input)
	if choice == "1" {
		return "video"
	} else if choice == "2" {
		return "channel"
	} else {
		fmt.Println("Invalid choice. Defaulting to single video download.")
		return "video"
	}
}

// getVideoTitle fetches the video title from Wistia API
func getVideoTitle(videoID string) string {
	assetResp, err := http.Get(fmt.Sprintf("https://fast.wistia.com/embed/medias/%s.json", videoID))
	if err != nil {
		return ""
	}
	defer assetResp.Body.Close()

	var assetInfo map[string]interface{}
	if err := json.NewDecoder(assetResp.Body).Decode(&assetInfo); err != nil {
		return ""
	}

	media, ok := assetInfo["media"].(map[string]interface{})
	if !ok {
		return ""
	}

	name, ok := media["name"].(string)
	if !ok {
		return ""
	}

	return name
}

// createSafeFilename creates a safe filename from a video title
func createSafeFilename(title string) string {
	// Replace unsafe characters
	filename := regexp.MustCompile(`[<>:"/\\|?*]`).ReplaceAllString(title, "_")
	filename = regexp.MustCompile(`\s+`).ReplaceAllString(filename, " ")
	filename = strings.TrimSpace(filename)

	// Truncate if too long (leave room for .mp4 extension)
	if len(filename) > 200 {
		filename = filename[:200]
	}

	return filename
}

// handleChannelDownload processes a channel page and downloads all videos
func handleChannelDownload(pageURL, outputFlag string) {
	fmt.Println("Fetching channel page...")

	// Fetch the channel page HTML
	resp, err := http.Get(pageURL)
	if err != nil {
		fmt.Printf("Error fetching channel page: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading channel page: %v\n", err)
		os.Exit(1)
	}

	htmlContent := string(body)

	// Extract and decode the channel data
	channelData, err := extractChannelData(htmlContent)
	if err != nil {
		fmt.Printf("Error extracting channel data: %v\n", err)
		os.Exit(1)
	}

	// Convert to videos list
	videos := extractVideosFromChannelData(channelData)

	if len(videos) == 0 {
		fmt.Println("No videos found in channel!")
		os.Exit(1)
	}

	// Display channel information
	fmt.Printf("\n📺 Channel Information:\n")
	fmt.Printf("Channel ID: %s\n", channelData.HashedID)
	fmt.Printf("Total videos found: %d\n", len(videos))

	// Group videos by section
	sectionCounts := make(map[string]int)
	for _, video := range videos {
		sectionCounts[video.Section]++
	}

	fmt.Println("\nVideos by section:")
	for section, count := range sectionCounts {
		fmt.Printf("  %s: %d videos\n", section, count)
	}

	fmt.Printf("\nNote: -o flag will be ignored. Files will be named based on video titles.\n")

	// Ask user for confirmation
	fmt.Print("\nDo you want to download all videos? (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		os.Exit(1)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		fmt.Println("Download cancelled.")
		os.Exit(0)
	}

	// Create downloads directory
	downloadsDir := "wistia_downloads"
	if err := os.MkdirAll(downloadsDir, 0755); err != nil {
		fmt.Printf("Error creating downloads directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nStarting download of %d videos to %s/\n", len(videos), downloadsDir)
	fmt.Println(strings.Repeat("=", 60))

	// Download all videos
	successful := 0
	failed := 0
	skipped := 0

	for i, video := range videos {
		fmt.Printf("\n[%d/%d] %s\n", i+1, len(videos), video.Title)

		filename := generateVideoFilename(video)
		outputPath := filepath.Join(downloadsDir, filename+".mp4")

		// Check if file already exists
		if _, err := os.Stat(outputPath); err == nil {
			fmt.Printf("⏭️  Skipping - file already exists: %s\n", filename+".mp4")
			skipped++
			continue
		}

		if downloadSingleVideoToPath(video.ID, outputPath) {
			successful++
		} else {
			failed++
		}

		// Small delay between downloads
		time.Sleep(1 * time.Second)
	}

	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("📊 Download Summary:\n")
	fmt.Printf("✅ Successful: %d\n", successful)
	fmt.Printf("⏭️  Skipped: %d\n", skipped)
	fmt.Printf("❌ Failed: %d\n", failed)
	fmt.Printf("📁 Files saved to: %s/\n", downloadsDir)
}

func extractVideoIDFromURL(url string) string {
	// Extract video ID from URL parameters, specifically wmediaid
	// Example: https://fast.wistia.com/embed/channel/m9k8d7f2jq?wchannelid=m9k8d7f2jq&wmediaid=p5v8q3n7rb
	re := regexp.MustCompile(`[?&]wmediaid=([a-zA-Z0-9]+)`)
	match := re.FindStringSubmatch(url)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}

func extractVideoIDFromPage(pageURL string) string {
	// Extract the link hash from URL structure
	urlRe := regexp.MustCompile(`wistia\.com/[^/]+/([a-zA-Z0-9]+)`)
	match := urlRe.FindStringSubmatch(pageURL)
	if len(match) > 1 {
		linkHashedId := match[1]
		// fmt.Printf("Extracted link hashed ID from URL structure: %s\n", linkHashedId) // Debug message hidden

		// Try to resolve this link hash to get the actual video ID using GraphQL API
		if videoID := resolveWistiaLinkGraphQL(pageURL, linkHashedId); videoID != "" {
			return videoID
		}
	}

	return ""
}

func resolveWistiaLinkGraphQL(pageURL, linkHashedId string) string {
	// Extract the domain from the page URL to use the correct GraphQL endpoint
	domainRe := regexp.MustCompile(`https?://([^/]+)`)
	domainMatch := domainRe.FindStringSubmatch(pageURL)
	if len(domainMatch) < 2 {
		fmt.Println("Could not extract domain from URL")
		return ""
	}
	domain := domainMatch[1]

	// Use the GraphQL endpoint from the HAR analysis
	graphqlURL := fmt.Sprintf("https://%s/graphql?op=AudienceLink", domain)

	// Construct the complete GraphQL query based on the HAR file analysis
	query := `{
		"operationName": "AudienceLink",
		"variables": {"hashedId": "` + linkHashedId + `"},
		"query": "query AudienceLink($hashedId: HashedId!) {\n  audienceLink(hashedId: $hashedId) {\n    id\n    status\n    validFrom\n    media {\n      id\n      ...anonymousMedia\n      __typename\n    }\n    __typename\n  }\n}\n\nfragment anonymousMedia on AnonymousMedia {\n  __typename\n  id\n  hashedId\n  aspectRatio\n  name\n  displayDescription\n  publicCommentsSelection\n  playerColor\n  mediaType\n  createdByRecord\n  hasMediaPage\n  hasReadyTimeCodedTranscript\n  hasSpeakers\n  mediaPage {\n    id\n    hasCustomizations\n    customizations\n    __typename\n  }\n  imageUrl\n  permissionsForCurrentContact {\n    canDownload\n    __typename\n  }\n  topLevelAudienceComments {\n    pageInfo {\n      endCursor\n      hasNextPage\n      __typename\n    }\n    edges {\n      node {\n        id\n        ...anonymousAudienceCommentFields\n        replies {\n          id\n          ...anonymousAudienceCommentFields\n          __typename\n        }\n        __typename\n      }\n      __typename\n    }\n    __typename\n  }\n  topLevelTeamComments {\n    pageInfo {\n      endCursor\n      hasNextPage\n      __typename\n    }\n    edges {\n      node {\n        id\n        ...anonymousTeamCommentFields\n        replies {\n          id\n          ...anonymousTeamCommentFields\n          __typename\n        }\n        __typename\n      }\n      __typename\n    }\n    __typename\n  }\n  account {\n    id\n    numericId\n    wistiaBrandingOptional\n    __typename\n  }\n}\n\nfragment anonymousAudienceCommentFields on AnonymousComment {\n  id\n  displayName\n  initials\n  body\n  createdAt\n  updatedAt\n  editedAt\n  mediaTimestamp\n  canEdit\n  canDelete\n  __typename\n}\n\nfragment anonymousTeamCommentFields on AnonymousTeamComment {\n  id\n  displayName\n  initials\n  body\n  bodyHtml\n  createdAt\n  updatedAt\n  editedAt\n  mediaTimestamp\n  canEdit\n  canDelete\n  __typename\n}"
	}`

	// fmt.Printf("Making GraphQL request to: %s\n", graphqlURL) // Debug message hidden

	// Create the HTTP request with proper headers
	req, err := http.NewRequest("POST", graphqlURL, strings.NewReader(query))
	if err != nil {
		fmt.Printf("Error creating GraphQL request: %v\n", err)
		return ""
	}

	// Add the required headers from the HAR analysis
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("x-wistia-gql-schema", "AnonymousSchema")
	req.Header.Set("Origin", fmt.Sprintf("https://%s", domain))
	req.Header.Set("Referer", pageURL)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.6 Safari/605.1.15")

	// Make the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making GraphQL request: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("GraphQL request failed with status: %d\n", resp.StatusCode)

		// Try to read the error response for debugging
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error response: %s\n", string(body))
		return ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading GraphQL response: %v\n", err)
		return ""
	}

	bodyStr := string(body)
	// fmt.Printf("GraphQL response: %s\n", bodyStr) // Debug message hidden

	// Extract the video hashedId from the response
	// Look for "hashedId":"x9f2k7m8vq" in the media object
	re := regexp.MustCompile(`"media":\s*{[^}]*"hashedId"\s*:\s*"([a-zA-Z0-9]+)"`)
	match := re.FindStringSubmatch(bodyStr)
	if len(match) > 1 {
		videoID := match[1]
		fmt.Printf("Found video ID via GraphQL: %s\n", videoID)
		return videoID
	}

	fmt.Println("Could not find video ID in GraphQL response")
	return ""
}

// downloadSingleVideo downloads a video with the given ID to the specified output file
func downloadSingleVideo(videoID, output string) {
	downloadSingleVideoToPath(videoID, output)
}

// downloadSingleVideoToPath downloads a video and returns success/failure
func downloadSingleVideoToPath(videoID, outputPath string) bool {
	// Get asset info for the video
	assetResp, err := http.Get(fmt.Sprintf("https://fast.wistia.com/embed/medias/%s.json", videoID))
	if err != nil {
		fmt.Printf("❌ Error fetching video info: %v\n", err)
		return false
	}
	defer assetResp.Body.Close()

	var assetInfo map[string]interface{}
	if err := json.NewDecoder(assetResp.Body).Decode(&assetInfo); err != nil {
		fmt.Printf("❌ Error parsing video info: %v\n", err)
		return false
	}

	// Find the best quality video URL
	videoURL := findBestVideoURL(assetInfo)
	if videoURL == "" {
		fmt.Printf("❌ No video download URL found\n")
		return false
	}

	// Download the video
	fmt.Printf("🎬 Downloading: %s\n", filepath.Base(outputPath))
	return downloadFile(videoURL, outputPath)
}

// extractChannelData extracts and decodes channel data from HTML content
func extractChannelData(htmlContent string) (*ChannelData, error) {
	// Find the encoded channel data in the HTML
	// Look for pattern like: window['wchanneljsonp-xxx'] = JSON.parse(decodeURIComponent(atob("encoded_data")));
	re := regexp.MustCompile(`window\['wchanneljsonp-[^']+'\]\s*=\s*JSON\.parse\(decodeURIComponent\(atob\("([^"]+)"\)\)\);`)
	matches := re.FindStringSubmatch(htmlContent)

	if len(matches) < 2 {
		return nil, fmt.Errorf("channel data not found in HTML")
	}

	encodedData := matches[1]

	// First base64 decode
	decodedOnce, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return nil, fmt.Errorf("base64 decode failed: %v", err)
	}

	// Then URL decode
	decodedTwice, err := url.QueryUnescape(string(decodedOnce))
	if err != nil {
		return nil, fmt.Errorf("URL decode failed: %v", err)
	}

	// Parse JSON
	var channelData ChannelData
	if err := json.Unmarshal([]byte(decodedTwice), &channelData); err != nil {
		return nil, fmt.Errorf("JSON decode failed: %v", err)
	}

	return &channelData, nil
}

// extractVideosFromChannelData converts channel data to a list of videos
func extractVideosFromChannelData(channelData *ChannelData) []Video {
	var videos []Video

	for _, series := range channelData.Series {
		for _, section := range series.Sections {
			for _, episode := range section.Episodes {
				videos = append(videos, Video{
					ID:          episode.HashedID,
					Title:       episode.Name,
					Description: episode.EpisodeDescription,
					Section:     section.Name,
				})
			}
		}
	}

	return videos
}

// generateVideoFilename creates a safe filename from video information
func generateVideoFilename(video Video) string {
	// Start with section and title
	filename := ""
	if video.Section != "" {
		filename = video.Section + " - " + video.Title
	} else {
		filename = video.Title
	}

	// Replace unsafe characters
	filename = regexp.MustCompile(`[<>:"/\\|?*]`).ReplaceAllString(filename, "_")
	filename = regexp.MustCompile(`\s+`).ReplaceAllString(filename, " ")
	filename = strings.TrimSpace(filename)

	// Truncate if too long
	if len(filename) > 200 {
		filename = filename[:200]
	}

	return filename
}

// findBestVideoURL finds the highest quality video URL from asset info
func findBestVideoURL(assetInfo map[string]interface{}) string {
	media, ok := assetInfo["media"].(map[string]interface{})
	if !ok {
		return ""
	}

	assets, ok := media["assets"].([]interface{})
	if !ok {
		return ""
	}

	var bestURL string
	var bestBitrate float64

	for _, assetInterface := range assets {
		asset, ok := assetInterface.(map[string]interface{})
		if !ok {
			continue
		}

		assetType, ok := asset["type"].(string)
		if !ok || assetType != "original" {
			continue
		}

		url, ok := asset["url"].(string)
		if !ok {
			continue
		}

		// Get bitrate if available
		bitrate, _ := asset["bitrate"].(float64)

		if bestURL == "" || bitrate > bestBitrate {
			bestURL = url
			bestBitrate = bitrate
		}
	}

	return bestURL
}

// downloadFile downloads a file from URL to the specified path with progress tracking
func downloadFile(url, filepath string) bool {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		fmt.Printf("❌ Error creating file: %v\n", err)
		return false
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("❌ Error downloading: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ Bad status: %s\n", resp.Status)
		return false
	}

	// Get the content length for progress tracking
	contentLength := resp.ContentLength
	if contentLength <= 0 {
		// If content length is unknown, download without progress
		fmt.Printf("📥 Downloading (size unknown)...")
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			fmt.Printf("\r❌ Error saving file: %v\n", err)
			return false
		}
		fmt.Printf("\r✅ Downloaded successfully\n")
		return true
	}

	// Create a progress reader
	var downloaded int64
	buffer := make([]byte, 32*1024) // 32KB buffer
	lastPercent := -1

	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			_, writeErr := out.Write(buffer[:n])
			if writeErr != nil {
				fmt.Printf("\r❌ Error writing to file: %v\n", writeErr)
				return false
			}
			downloaded += int64(n)

			// Calculate and display progress
			percent := int(float64(downloaded) / float64(contentLength) * 100)
			if percent != lastPercent {
				// Create progress bar (20 characters wide)
				barWidth := 20
				filled := int(float64(barWidth) * float64(percent) / 100)
				bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

				sizeMB := float64(contentLength) / 1024 / 1024
				downloadedMB := float64(downloaded) / 1024 / 1024

				// Use \r to overwrite the same line
				fmt.Printf("\r📥 [%s] %3d%% (%.2f/%.2f MB)", bar, percent, downloadedMB, sizeMB)
				lastPercent = percent
			}
		}
		if err != nil {
			if err != io.EOF {
				fmt.Printf("\r❌ Error reading: %v\n", err)
				return false
			}
			break
		}
	}

	fmt.Printf("\r✅ Downloaded successfully (%.2f MB)%s\n", float64(contentLength)/1024/1024, strings.Repeat(" ", 10))
	return true
}

// extractVideoIDFromHTML extracts video ID from HTML snippet
func extractVideoIDFromHTML(htmlSnippet string) string {
	// Extract video ID from HTML snippet (like from "Copy link")
	// Look for wvideo=xxx pattern
	re := regexp.MustCompile(`wvideo=([a-zA-Z0-9]+)`)
	match := re.FindStringSubmatch(htmlSnippet)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}
