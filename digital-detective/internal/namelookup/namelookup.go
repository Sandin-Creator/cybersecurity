package namelookup

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
}

func getRandomUserAgent() string {
	return userAgents[rand.Intn(len(userAgents))]
}

// ValidationStatus represents the result of a search check
type ValidationStatus int

const (
	StatusFound ValidationStatus = iota
	StatusEmpty
	StatusBlocked
	StatusError
)

// DorkEntry represents a single search tool
type DorkEntry struct {
	Name        string `json:"name"`
	Dork        string `json:"dork"`
	Description string `json:"description"`
}

// DorkConfig holds all categories
type DorkConfig struct {
	General    []DorkEntry `json:"general"`
	Files      []DorkEntry `json:"files"`
	Social     []DorkEntry `json:"social"`
	Government []DorkEntry `json:"government"`
	Leaks      []DorkEntry `json:"leaks"`
}

// SearchByName generates Google Dorks for a full name, loaded from Dynamic Configuration
func SearchByName(fullname, additionalInfo, country string) (string, map[string]string) {
	var sb strings.Builder
	imageMap := make(map[string]string)

	// Google Logo - Using a reliable Wikimedia/stable URL or base64.
	// For simplicity using a CDN link to a Google "G" icon
	googleIcon := "https://upload.wikimedia.org/wikipedia/commons/5/53/Google_%22G%22_Logo.svg"

	sb.WriteString(fmt.Sprintf("OSINT Report for: %s\n", fullname))

	// Context for Dork Templates
	ctx := struct {
		Name      string
		Info      string
		Country   string
		FullQuery string
	}{
		Name:    fullname,
		Info:    additionalInfo,
		Country: country,
	}

	// Build a base query for easy fallback
	baseQuery := fmt.Sprintf(`"%s"`, fullname)
	if additionalInfo != "" {
		baseQuery += fmt.Sprintf(` "%s"`, additionalInfo)
	}
	if country != "" {
		baseQuery += fmt.Sprintf(` "%s"`, country)
	}
	ctx.FullQuery = baseQuery

	if additionalInfo != "" || country != "" {
		sb.WriteString(fmt.Sprintf("Criteria: %s %s\n", additionalInfo, country))
	}
	sb.WriteString("--------------------------------------------------\n")
	sb.WriteString("REAL DATA LOOKUP TOOLS (Active Verification)\n")
	sb.WriteString("Checking links for results (this may take a moment)...\n\n")

	// Load Dorks
	config, err := loadDorks()
	if err != nil {
		sb.WriteString(fmt.Sprintf("[-] Error loading dorks.json: %v\n", err))
		return sb.String(), nil
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Helper to process a category
	printCategory := func(title string, entries []DorkEntry) {
		if len(entries) == 0 {
			return
		}
		sb.WriteString(fmt.Sprintf(">>> %s\n", title))
		for _, entry := range entries {
			query := entry.Dork
			query = strings.ReplaceAll(query, "{{.Name}}", ctx.FullQuery)

			encoded := url.QueryEscape(query)
			link := fmt.Sprintf("https://www.google.com/search?q=%s", encoded)

			// ACTIVE VALIDATION
			status := validateDork(client, link)
			switch status {
			case StatusFound:
				sb.WriteString(fmt.Sprintf("[+] %s\n    %s\n", entry.Name, link))
				imageMap[link] = googleIcon
			case StatusEmpty:
				sb.WriteString(fmt.Sprintf("[-] %s: No results found.\n", entry.Name))
			case StatusBlocked:
				sb.WriteString(fmt.Sprintf("[?] %s: Verification Blocked (Manual Check)\n    %s\n", entry.Name, link))
				imageMap[link] = googleIcon
			case StatusError:
				sb.WriteString(fmt.Sprintf("[!] %s: Validation Failed (Network Error)\n    %s\n", entry.Name, link))
				imageMap[link] = googleIcon
			}

			// Sleep to avoid IP ban (4s - 8s jitter)
			time.Sleep(time.Duration(rand.Intn(4000)+4000) * time.Millisecond)
		}
		sb.WriteString("\n")
	}

	// Print Categories
	printCategory("GENERAL SEARCH", config.General)
	printCategory("SOCIAL MEDIA INVESTIGATION", config.Social)
	// User requested to remove Files, Government, and Leaks categories
	// printCategory("FILE FORENSICS (PDF, Excel, Docs)", config.Files)
	// printCategory("GOVERNMENT RECORDS", config.Government)
	// printCategory("LEAKS & EXPOSURES", config.Leaks)

	sb.WriteString("--------------------------------------------------\n")

	// SOURCE 2: DuckDuckGo API (Real OSINT Cross-Reference)
	sb.WriteString("CROSS-REFERENCE SOURCE 2: DuckDuckGo Instant Answer API\n")
	duckResult, duckIcon := searchDuckDuckGo(fullname, client)
	sb.WriteString(duckResult)
	if duckIcon != "" {
		imageMap["DuckDuckGo"] = duckIcon
	}

	sb.WriteString("--------------------------------------------------\n")
	sb.WriteString(" PRIVACY TIP: These results are live from Google and DuckDuckGo. \n")
	sb.WriteString(" To remove them, you must contact the hosting site or search engine.\n")

	return sb.String(), imageMap
}

// validateDork checks if the Google search result is empty or blocked
func validateDork(client *http.Client, url string) ValidationStatus {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return StatusError
	}

	// Mimic browser to avoid immediate block
	req.Header.Set("User-Agent", getRandomUserAgent())
	// Add Accept-Language to prefer English, which helps with matching "did not match" logic
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Referer", "https://www.google.com/")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	// Consent Cookie to bypass "Before you continue" wall in EU
	req.Header.Set("Cookie", "CONSENT=YES+; SOCS=CAESHAgBEhJnd3NfMjAyMzA4MTAtMF9SQzEaAmVuIAEaBgiAo_WmBg")

	resp, err := client.Do(req)
	if err != nil {
		return StatusError
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		return StatusBlocked
	}

	if resp.StatusCode != 200 {
		return StatusError
	}

	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024*500)) // Read first 500KB
	body := string(bodyBytes)

	// Check for block signatures first
	if strings.Contains(body, "httpservice/retry") ||
		strings.Contains(body, "If you're having trouble accessing Google Search") ||
		strings.Contains(body, "google.com/sorry") ||
		strings.Contains(body, "document.getElementById('captcha')") {
		return StatusBlocked
	}

	// Check for "no results" indicators
	// Relaxed check to "did not match" to catch cases with formatting tags like <b> or <em>
	if strings.Contains(body, "did not match") ||
		strings.Contains(body, "Esittämiäsi hakusanoja") { // Finnish Google
		return StatusEmpty
	}

	// Default to Found if we have a 200 OK and no negative indicators
	// Default to Found if we have a 200 OK and no negative indicators
	return StatusFound
}

func loadDorks() (DorkConfig, error) {
	var config DorkConfig

	// Try local path first (relative to binary if run from root)
	path := "cmd/digger/sources/dorks.json"
	file, err := os.Open(path)
	if err != nil {
		// Fallback for running inside cmd/digger or tests
		path = "../../cmd/digger/sources/dorks.json"
		file, err = os.Open(path)
		if err != nil {
			// Container/Root fallback
			path = "sources/dorks.json"
			file, err = os.Open(path)
			if err != nil {
				return config, err
			}
		}
	}
	defer file.Close()

	bytes, _ := io.ReadAll(file)
	err = json.Unmarshal(bytes, &config)
	return config, err
}
