package usernamelookup

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// PlatformSource defines a target site structure from JSON
type PlatformSource struct {
	Name          string            `json:"name"`
	URL           string            `json:"url"`
	CheckURL      string            `json:"check_url,omitempty"`
	Category      string            `json:"category"`
	ErrorMsg      string            `json:"error_msg"`
	ValidationMsg string            `json:"validation_msg,omitempty"`
	Headers       map[string]string `json:"headers,omitempty"`
	CheckUsername bool              `json:"check_username,omitempty"`
}

// SearchByUsername checks specific social media sites for the username
func SearchByUsername(input string, includeDating bool) (string, map[string]string) {
	var sb strings.Builder
	var handle string
	// Map to store discovered profile images [url] -> image_url
	imageMap := make(map[string]string)

	isEmail := strings.Contains(input, "@")

	if isEmail {
		parts := strings.Split(input, "@")
		if len(parts) >= 2 {
			handle = parts[0]
			sb.WriteString(fmt.Sprintf("Detected Email: %s\n", input))
			sb.WriteString(fmt.Sprintf("Extracted Username/Handle: %s\n", handle))
			sb.WriteString("--------------------------------------------------\n")
		} else {
			handle = input
		}
	} else {
		// CHANGED: Sanitize input by removing spaces
		handle = strings.ReplaceAll(input, " ", "")
	}

	sb.WriteString(fmt.Sprintf("Username Scan for: '%s'\n", handle))
	sb.WriteString("--------------------------------------------------\n")

	// 1. Load Sources
	sources, err := loadUsernameSources()
	if err != nil {
		sb.WriteString(fmt.Sprintf("[-] Error loading sources: %v\n", err))
		return sb.String(), nil
	}

	var foundResults []string
	var notFoundResults []string

	// Map to store content for Dating Investigator
	foundContent := make(map[string]string)

	var wg sync.WaitGroup
	var mu sync.Mutex

	client := &http.Client{
		Timeout: 10 * time.Second,
		// CHANGED: Allow redirects to detect login pages
		// Default CheckRedirect is nice, but we might want to manually inspect the final URL
	}

	// RATE LIMITING: Semaphore to limit concurrency
	// Only allow 3 concurrent requests to avoid triggering anti-bot protections
	maxConcurrency := 3
	sem := make(chan struct{}, maxConcurrency)

	// 2. Scan Concurrently
	for _, source := range sources {
		wg.Add(1)
		go func(s PlatformSource) {
			// Acquire semaphore
			sem <- struct{}{}
			defer func() {
				// Release semaphore
				<-sem
				wg.Done()
			}()

			// Add Random Jitter (0.5s to 2.5s)
			// This makes the requests look less like a burst from a script
			sleepTime := time.Duration(rand.Intn(2000)+500) * time.Millisecond
			time.Sleep(sleepTime)

			checkURLStr := s.URL
			if s.CheckURL != "" {
				checkURLStr = s.CheckURL
			}
			targetCheckURL := strings.Replace(checkURLStr, "{account}", handle, 1)
			targetDisplayURL := strings.Replace(s.URL, "{account}", handle, 1)

			found, content := checkURL(client, targetCheckURL, s.ErrorMsg, s.ValidationMsg, s.Headers, s.CheckUsername, handle)

			mu.Lock()
			if found {
				foundResults = append(foundResults, fmt.Sprintf("[+] FOUND:     %-12s %s", s.Name, targetDisplayURL))
				foundContent[s.Name] = content

				// Extract Image
				imgURL := extractImageURL(content)
				if imgURL != "" {
					imageMap[targetDisplayURL] = imgURL
				}
			} else {
				notFoundResults = append(notFoundResults, fmt.Sprintf("[-] NOT FOUND: %-12s %s", s.Name, targetDisplayURL))
			}
			mu.Unlock()
		}(source)
	}
	wg.Wait()

	// 3. Output
	// List FOUND first
	if len(foundResults) > 0 {
		for _, res := range foundResults {
			sb.WriteString(res + "\n")
		}
	} else {
		sb.WriteString("[-] No accounts found on configured platforms.\n")
	}

	// List NOT FOUND
	if len(notFoundResults) > 0 {
		sb.WriteString("--------------------------------------------------\n")
		for _, res := range notFoundResults {
			sb.WriteString(res + "\n")
		}
	}

	sb.WriteString("--------------------------------------------------\n")
	sb.WriteString(fmt.Sprintf("CONFIDENCE: %d detections\n", len(foundResults)))

	// 2. Email Dorks (if input was email)
	if isEmail {
		// New: Gravatar Check
		wg.Add(1)
		go func() {
			defer wg.Done()
			if gravatarURL, exists := checkGravatar(input); exists {
				mu.Lock()
				foundResults = append(foundResults, fmt.Sprintf("[+] FOUND: %-12s %s", "Gravatar", gravatarURL))
				imageMap[gravatarURL] = gravatarURL // Gravatar URL is the image itself
				mu.Unlock()
			}
		}()
		wg.Wait() // Wait for gravatar check too

		sb.WriteString("\n" + generateEmailDorks(input))
	}

	// 3. Dating App Investigator (Only if requested)
	if includeDating {
		sb.WriteString("\n")
		sb.WriteString(AnalyzeDatingSignals(foundContent))
	}

	return sb.String(), imageMap
}

// extractImageURL attempts to find og:image or twitter:image
func extractImageURL(htmlBody string) string {
	// Simple regex for og:image
	// <meta property="og:image" content="URL" />
	// This is not a full HTML parser but sufficient for most cases
	if strings.Contains(htmlBody, "og:image") {
		start := strings.Index(htmlBody, "og:image")
		contentStart := strings.Index(htmlBody[start:], "content=\"")
		if contentStart != -1 {
			realStart := start + contentStart + 9
			end := strings.Index(htmlBody[realStart:], "\"")
			if end != -1 {
				return htmlBody[realStart : realStart+end]
			}
		}
	}
	return ""
}

func checkGravatar(email string) (string, bool) {
	hash := md5.Sum([]byte(strings.ToLower(strings.TrimSpace(email))))
	hashStr := hex.EncodeToString(hash[:])
	targetURL := fmt.Sprintf("https://en.gravatar.com/%s.json", hashStr)
	profileURL := fmt.Sprintf("https://en.gravatar.com/%s", hashStr)

	resp, err := http.Get(targetURL)
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return profileURL, true
	}
	return "", false
}

func loadUsernameSources() ([]PlatformSource, error) {
	// Try local path first
	path := "cmd/digger/sources/usernames.json"
	file, err := os.Open(path)
	if err != nil {
		// Fallback for tests/different dirs
		path = "../../cmd/digger/sources/usernames.json"
		file, err = os.Open(path)
		if err != nil {
			// Container/Root fallback
			path = "sources/usernames.json"
			file, err = os.Open(path)
			if err != nil {
				return nil, err
			}
		}
	}
	defer file.Close()

	bytes, _ := io.ReadAll(file)
	var sources []PlatformSource
	err = json.Unmarshal(bytes, &sources)
	return sources, err
}

func checkURL(client *http.Client, url string, errorMsg string, validationMsg string, headers map[string]string, checkUsername bool, handle string) (bool, string) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, ""
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, ""
	}
	defer resp.Body.Close()

	// Basic Status Check
	if resp.StatusCode != 200 {
		return false, ""
	}

	// CHANGED: Redirect Detection
	// If the final URL is significantly different (e.g., /login), it's likely a redirect
	finalURL := resp.Request.URL.String()
	if strings.Contains(finalURL, "/login") || strings.Contains(finalURL, "signin") {
		return false, ""
	}

	// Double check content
	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024*500))
	body := string(bodyBytes)

	// Positive Validation (Strongest Check)
	if validationMsg != "" {
		if !strings.Contains(body, validationMsg) {
			return false, ""
		}
	}

	if errorMsg != "" {
		if strings.Contains(body, errorMsg) {
			return false, ""
		}
	}

	// Generic "Log in" check for false positive 200s
	// Many sites return 200 OK for their login page when you try to visit a profile that doesn't exist but redirects to login.
	// This is a heuristic and might need tuning.
	if strings.Contains(strings.ToLower(body), "<title>login") || strings.Contains(strings.ToLower(body), "log in to") {
		return false, ""
	}

	// Logic for CheckUsername
	if checkUsername && handle != "" {
		if !strings.Contains(strings.ToLower(body), strings.ToLower(handle)) {
			return false, ""
		}
	}

	return true, body
}

func generateEmailDorks(email string) string {
	var sb strings.Builder
	sb.WriteString("--------------------------------------------------\n")
	sb.WriteString("REAL DATA LOOKUP TOOLS (Email Dorks)\n")
	sb.WriteString("Click these links to find where this email is mentioned publically:\n\n")

	escapedEmail := url.QueryEscape(fmt.Sprintf("\"%s\"", email))

	dorks := []struct {
		Title string
		Query string
	}{
		{"General Search (Exact Match)", escapedEmail},
		{"Social Media Mentions", escapedEmail + "+site:twitter.com+OR+site:facebook.com+OR+site:linkedin.com+OR+site:instagram.com"},
		{"Leaked Databases/Pastebin", escapedEmail + "+site:pastebin.com+OR+site:throwbin.io+OR+site:justpaste.it"},
		{"PDF Documents", escapedEmail + "+filetype:pdf"},
		{"Text Files (Breach Dumps)", escapedEmail + "+filetype:txt"},
	}

	for _, d := range dorks {
		fullQuery := "https://www.google.com/search?q=" + d.Query
		sb.WriteString(fmt.Sprintf("[+] %s:\n    %s\n\n", d.Title, fullQuery))
	}

	return sb.String()
}
