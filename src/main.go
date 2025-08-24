package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func main() {
	videoID := flag.String("id", "", "Wistia video ID (e.g. m8k4z7x9pq)")
	pageURL := flag.String("url", "", "Main Wistia page URL (e.g. https://site.wistia.com/a/x7k2m9n5qp3w)")
	clipboardHTML := flag.String("clipboard", "", "HTML snippet from 'Copy link' (contains wvideo parameter)")
	output := flag.String("o", "video.mp4", "Output filename")
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
		id = extractVideoIDFromPage(*pageURL)
		if id == "" {
			fmt.Println("Could not find Wistia video ID in page.")
			os.Exit(1)
		}
		fmt.Println("Found video ID:", id)
	} else {
		fmt.Println("Usage: wistia-downloader -id <videoID> OR -url <WistiaPageURL> OR -clipboard <HTMLSnippet> [-o <output.mp4>]")
		os.Exit(1)
	}

	embedURL := "http://fast.wistia.net/embed/iframe/" + id
	resp, err := http.Get(embedURL)
	if err != nil {
		fmt.Println("Error fetching embed page:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading embed page:", err)
		os.Exit(1)
	}

	videoURL := extractVideoURL(string(body))
	if videoURL == "" {
		fmt.Println("Could not find video URL in embed page.")
		os.Exit(1)
	}
	videoURL = strings.Replace(videoURL, ".bin", ".mp4", 1)
	fmt.Println("Downloading:", videoURL)

	out, err := os.Create(*output)
	if err != nil {
		fmt.Println("Error creating output file:", err)
		os.Exit(1)
	}
	defer out.Close()

	videoResp, err := http.Get(videoURL)
	if err != nil {
		fmt.Println("Error downloading video:", err)
		os.Exit(1)
	}
	defer videoResp.Body.Close()

	_, err = io.Copy(out, videoResp.Body)
	if err != nil {
		fmt.Println("Error saving video:", err)
		os.Exit(1)
	}

	fmt.Println("Download complete:", *output)
}

func extractVideoURL(body string) string {
	// Try "type":"original" first
	re := regexp.MustCompile(`"type":"original".*?"url":"(http[^"]+)"`)
	match := re.FindStringSubmatch(body)
	if len(match) > 1 {
		return match[1]
	}
	// Try "type":"hd_mp4_video"
	re = regexp.MustCompile(`"type":"hd_mp4_video".*?"url":"(http[^"]+)"`)
	match = re.FindStringSubmatch(body)
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
		fmt.Printf("Extracted link hashed ID from URL structure: %s\n", linkHashedId)

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

	fmt.Printf("Making GraphQL request to: %s\n", graphqlURL)

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
	fmt.Printf("GraphQL response: %s\n", bodyStr)

	// Extract the video hashedId from the response
	// Look for "hashedId":"4c0kmodatj" in the media object
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
