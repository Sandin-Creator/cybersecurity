# Digital Detective: How It Works

This project is a modular Open Source Intelligence (OSINT) tool designed to gather public information about entities (People, Usernames, IPs) using safe, passive, and active reconnaissance techniques.

## Core Modules

### 1. Username Search (`-un`)
**Goal:** Determine if a specific username exists across various platforms (Instagram, Facebook, Twitter, etc.).

**Are we using APIs?**
*   **Official Developer APIs:** **No.** We do not use the official/paid APIs (e.g., Twitter API v2, Instagram Graph API) because they require personal API keys, authentication, and often cost money. This tool is designed to be "plug-and-play" without setup.
*   **Internal/Frontend APIs:** **Yes, for some.** We sometimes use the *hidden* APIs that the websites themselves use to load data into your browser.
    *   *Example:* For **Instagram**, we call `web_profile_info` (an internal endpoint). This gives us precise JSON data instead of messy HTML, but it still counts as "scraping" to Instagram, so they rate-limit it.
*   **Web Scraping:** **Yes, for most.** For sites like **Facebook** or **Twitter**, we download the public profile page (HTML) and look for clues.

**How it works (The Technical Part):**
*   **Direct Scraping:** The tool constructs a URL for the target username (e.g., `instagram.com/username`) and sends a direct HTTP request.
*   **Status Codes:** It checks the HTTP response code.
    *   `404 Not Found` usually means the user doesn't exist.
    *   `200 OK` usually means they do.
*   **Advanced Validation:** Simple status codes are often misleading (False Positives). Many modern sites (like Instagram or Twitter) return `200 OK` even for non-existent users (sending a generic "Search" page).
    *   **Content Check:** The tool scans the page HTML for specific keywords (e.g., "Followers", "id": 12345) to *prove* the profile is real.
    *   **Internal APIs:** For difficult sites (like Instagram), the tool mimics the browser's internal API calls (e.g., `web_profile_info`) using specific headers (`X-IG-App-ID`) to get a definitive JSON response instead of a vague HTML page.
    *   **Headers:** It pretends to be a real Chrome browser (`User-Agent`) to avoid being blocked by anti-bot firewalls.

### 2. Name Search (`-n`)
**Goal:** Find traces of a real person's full name on the internet.

**How it works:**
*   **Google Dorks:** This module does **not** scrape sites directly (which would require solving CAPTCHAs).
*   **Query Generation:** Instead, it generates advanced Google Search operators ("Dorks") customized for the input name.
    *   *Example:* `site:linkedin.com/in/ "John Doe"`
    *   *Example:* `filetype:pdf "John Doe"` (checks for leaked documents)
*   **User Action:** It presents these pre-built, clickable links to the user. The user clicks them to view live results from Google safely.

### 3. Dating App Investigator (`-dt` or automatic)
**Goal:** Assess the likelihood of a person being on dating apps without logging into those apps (which is often impossible via script).

**How it works (Inference Engine):**
*   **Secondary Signals:** It scans the *found* profiles from the Username Search (e.g., Pinterest, Spotify, Twitter) for behavioral patterns.
*   **Keyword Scoring:** It looks for tell-tale keywords in bios and posts:
    *   "swipe right", "tinder", "dating", "single", "snap:"
*   **Profile Pictures:** (If implemented) It checks if the same profile picture hash appears on dating-adjacent sites.
*   **Score:** It calculates a "Likelihood Score" (1-10) based on these signals. It is a probabilistic guess, not a confirmed database lookup.

### 4. IP Investigation (`-ip`)
**Goal:** Geolocation and ISP information for an IP address.

**How it works:**
*   **API Lookup:** It queries public geolocation APIs (like `ip-api.com` or `ipinfo.io`).
*   **Data Retrieval:** It pulls the Country, City, ISP, and coordinates associated with that IP address.

## File Structure & Logic
*   `cmd/digger/main.go`: The entry point. Handles arguments (`-un`, `-n`) and orchestrates the search.
*   `sources/usernames.json`: The "Brain" of the username search. Contains the list of known sites, their URL patterns, logic for detecting errors, and headers needed to bypass obstacles.
*   `internal/usernamelookup/`: logic for making HTTP requests, handling redirects, and parsing responses.

## Limitations
*   **Login Walls:** Sites like Facebook or LinkedIn often hide profile data behind a login screen. The tool does not use valid account credentials (for safety), so it can only see what is *publicly* visible. If a profile is "Friends Only", this tool will likely report "NOT FOUND" or provide a general link.
*   **Rate Limits:** Making too many requests too fast (especially to Instagram) can cause temporary IP blocks, leading to "False Negatives".
