# Digital Detective 🕵️‍♂️

Digital Detective is an Open Source Intelligence (OSINT) CLI tool designed to gather and analyze information from publicly available sources. It helps cybersecurity professionals and enthusiasts understand digital footprints by aggregating data related to names, IP addresses, usernames, and emails.

## Features ✨

*   **Multi-Parameter Search**: Search by Name, IP, Username, or Email individually or simultaneously.
*   **IP Intelligence**: Geolocation, ISP, and Organization details using 3 different APIs (ip-api, ipapi.co, ipwho.is).
*   **Social Media Scan**: Checks for username presence across major platforms (GitHub, Reddit, Instagram, etc.).
*   **Google Dork Generation**: Generates advanced search queries to uncover deep web info (documents, leaks, government records) for names and emails.
*   **Email Verification**: **Bonus Feature!** Checks for Gravatar profiles to validate email existence.
*   **Smart Reporting**:
    *   Saves reports to a dedicated `reports/` directory.
    *   Prevents file overwriting (auto-increments filenames).
    *   Visualized ASCII output in the terminal.

## Installation 🛠️

### Prerequisites
*   [Go](https://go.dev/dl/) (1.21 or higher recommended)
*   *Optional*: Docker

### Build from Source
```bash
git clone https://github.com/yourusername/digital-detective.git
cd digital-detective
go build -o data_digger cmd/digger/main.go
```

### Run via Docker
```bash
docker build -t data_digger .
docker run --rm data_digger --help
```

## Usage 🚀

Run the tool using the built binary or directly with `go run`.

```bash
# View Help
./data_digger --help

# 0. Web Interface (NEW!)
./data_digger -server
# Then open http://localhost:8080 in your browser

# 1. Name Search (Generates OSINT Dorks)
./data_digger -n "Christian Sandin"

# 2. IP Lookup (Geolocation & ISP)
./data_digger -ip 8.8.8.8

# 3. Username Search (Social Media Check)
./data_digger -un "testuser"

# 4. Email Search (Username Check + Gravatar + Dorks)
./data_digger -em "test@example.com"

# 5. Dating App Investigation (Username Reuse + Signals)
./data_digger -dt "potential_match"

# 6. Combined Search (All at once!)
./data_digger -n "John Doe" -ip 1.1.1.1 -un "jdoe"
```

## Output 📁

Reports are automatically saved to the `reports/` folder.
*   Format: `reports/{TYPE}_{VALUE}.txt`
*   Example: `reports/n_un_JohnDoe.txt`

## Architecture 🏗️

*   **cmd/digger**: Entry point and CLI handling.
*   **internal/iplookup**: Logic for querying IP APIs.
*   **internal/usernamelookup**: Logic for platform scraping and Gravatar checks.
*   **internal/namelookup**: Logic for generating search query links.
*   **internal/report**: Output formatting and file management.

## Disclaimer ⚠️

This tool is for educational and cybersecurity research purposes only. Always respect privacy laws and Terms of Service of the platforms you investigate.
