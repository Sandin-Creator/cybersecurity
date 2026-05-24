package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"cybersecurity-go/digital-detective/internal/iplookup"
	"cybersecurity-go/digital-detective/internal/namelookup"
	"cybersecurity-go/digital-detective/internal/report"
	"cybersecurity-go/digital-detective/internal/usernamelookup"
)

// SearchRequest defines the JSON payload for API
type SearchRequest struct {
	Type  string `json:"type"`  // "username", "name", "ip"
	Query string `json:"query"` // The input value
}

// SearchResponse defines the JSON response
type SearchResponse struct {
	Output string            `json:"output"`
	Images map[string]string `json:"images"` // URL -> ImageURL
}

func main() {
	serverPtr := flag.Bool("server", false, "Start the web interface server")
	namePtr := flag.String("n", "", "Performs a full-name search")
	ipPtr := flag.String("ip", "", "IP Address to investigate")
	// CHANGED: -u to -un to match requirements
	usernamePtr := flag.String("un", "", "Username to search")
	emailPtr := flag.String("em", "", "Email to search")
	datingPtr := flag.String("dt", "", "Username for dating app investigation")

	additionalInfoPtr := flag.String("i", "", "Additional Info (City, Job, etc.)")
	countryPtr := flag.String("c", "", "Country")

	flag.Usage = func() {
		fmt.Println("Usage: data_digger [options] <input>")
		fmt.Println("OPTIONS:")
		fmt.Println("    -server Start the web interface server (http://localhost:8080)")
		fmt.Println("    -n      Performs a full-name search.")
		fmt.Println("    -ip     Performs a IP search.")
		fmt.Println("    -un     Performs a username search.")
		fmt.Println("    -em     Performs an email search.")
		fmt.Println("    -dt     Performs a dating app investigation.")
		fmt.Println("\nAdditional:")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *serverPtr {
		startServer()
		return
	}

	// 1. Validate Input
	if *namePtr == "" && *ipPtr == "" && *usernamePtr == "" && *emailPtr == "" && *datingPtr == "" {
		fmt.Println("No search parameters provided. Use -help for usage.")
		return // Exit if no args
	}

	// 2. Execute Search
	finalResult, filenameParts, reportContent, _ := executeSearch(*namePtr, *ipPtr, *usernamePtr, *emailPtr, *datingPtr, *additionalInfoPtr, *countryPtr)

	// 3. Display Output
	fmt.Println(finalResult)

	// 4. Save to File (Safe Write)
	saveReportToFile(finalResult, filenameParts, reportContent)
}

// startServer launches the HTTP server
func startServer() {
	webDir := getWebDir()
	// Serve static files from "web" directory
	fs := http.FileServer(http.Dir(webDir))
	http.Handle("/", fs)

	// API Endpoint
	http.HandleFunc("/api/search", handleSearch)

	fmt.Println("============================================================")
	fmt.Println("  DIGITAL DETECTIVE // WEB SERVER ACTIVE")
	fmt.Printf("  SERVING UI FROM: %s\n", webDir)
	fmt.Println("  ACCESS UI AT: http://localhost:8080")
	fmt.Println("============================================================")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}
}

func getWebDir() string {
	if info, err := os.Stat("web"); err == nil && info.IsDir() {
		return "./web"
	}
	if info, err := os.Stat("../../web"); err == nil && info.IsDir() {
		return "../../web"
	}
	return "./web"
}

// handleSearch processes API requests
func handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Map API request to CLI args
	var name, ip, username, email, dating string
	switch req.Type {
	case "username":
		username = req.Query
	case "name":
		name = req.Query
	case "ip":
		ip = req.Query
	case "email":
		email = req.Query
	case "dating":
		dating = req.Query
	}

	// Execute
	finalResult, filenameParts, reportContent, imageMap := executeSearch(name, ip, username, email, dating, "", "")

	// Save Report
	saveReportToFile(finalResult, filenameParts, reportContent)

	// Respond
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SearchResponse{Output: finalResult, Images: imageMap})
}

// executeSearch contains the core logic, refactored for reuse
func executeSearch(name, ip, username, email, dating, info, country string) (string, []string, string, map[string]string) {
	var resultsBuilder strings.Builder
	var filenameParts []string
	var reportContent string
	imageMap := make(map[string]string)

	resultsBuilder.WriteString(report.GetBanner())

	// Name Search
	if name != "" {
		res, imgs := namelookup.SearchByName(name, info, country)
		resultsBuilder.WriteString(report.FormatSection("NAME LOOKUP", res))
		filenameParts = append(filenameParts, "n")
		if reportContent == "" {
			reportContent = name
		}
		// Merge images
		for k, v := range imgs {
			imageMap[k] = v
		}
	}

	// Username Search
	if username != "" {
		res, imgs := usernamelookup.SearchByUsername(username, false)
		resultsBuilder.WriteString(report.FormatSection("USERNAME SEARCH", res))
		filenameParts = append(filenameParts, "un")
		if reportContent == "" {
			reportContent = username
		}
		// Merge images
		for k, v := range imgs {
			imageMap[k] = v
		}
	}

	// Email Search
	if email != "" {
		res, imgs := usernamelookup.SearchByUsername(email, false)
		resultsBuilder.WriteString(report.FormatSection("EMAIL SEARCH", res))
		filenameParts = append(filenameParts, "em")
		if reportContent == "" {
			reportContent = email
		}
		// Merge images
		for k, v := range imgs {
			imageMap[k] = v
		}
	}

	// Dating Search
	if dating != "" {
		res, imgs := usernamelookup.SearchByUsername(dating, true)
		resultsBuilder.WriteString(report.FormatSection("DATING INVESTIGATION", res))
		filenameParts = append(filenameParts, "dt")
		if reportContent == "" {
			reportContent = dating
		}
		// Merge images
		for k, v := range imgs {
			imageMap[k] = v
		}
	}

	// IP Search
	if ip != "" {
		res, imgs := iplookup.SearchByIP(ip)
		resultsBuilder.WriteString(report.FormatSection("IP INVESTIGATION", res))
		filenameParts = append(filenameParts, "ip")
		if reportContent == "" {
			reportContent = ip
		}
		// Merge images
		for k, v := range imgs {
			imageMap[k] = v
		}
	}

	return resultsBuilder.String(), filenameParts, reportContent, imageMap
}

func saveReportToFile(finalResult string, filenameParts []string, reportContent string) {
	// Filename format: params_{value}.txt. e.g. n_un_ip_JohnDoe.txt
	basePrefix := strings.Join(filenameParts, "_")

	// Sanitize reportContent for filename
	safeValue := strings.ReplaceAll(reportContent, " ", "_")
	safeValue = strings.ReplaceAll(safeValue, "/", "_")
	safeValue = strings.ReplaceAll(safeValue, "\\", "_")

	// Determine subdirectory
	subDir := "combined"
	if len(filenameParts) == 1 {
		switch filenameParts[0] {
		case "n":
			subDir = "fullname"
		case "ip":
			subDir = "ip"
		case "un":
			// Legacy fallback or explicit un
			if strings.Contains(reportContent, "@") {
				subDir = "email"
			} else {
				subDir = "username"
			}
		case "em":
			subDir = "email"
		case "dt":
			subDir = "date"
		}
	}

	outputDir := getOutputDir(subDir)
	baseFilename := fmt.Sprintf("%s/%s_%s.txt", outputDir, basePrefix, safeValue)
	finalFilename := getSafeFilename(baseFilename)

	err := os.WriteFile(finalFilename, []byte(finalResult), 0644)
	if err != nil {
		fmt.Printf("[-] Error writing report: %v\n", err)
	} else {
		// Clean up path for display if it's cleaner
		displayPath := finalFilename
		// Fix relative path display issues for readability
		if strings.HasPrefix(displayPath, "../../") {
			// normalized for display if needed
		}
		fmt.Printf("[+] Result written to file: %s\n", displayPath)
	}
}

// getOutputDir attempts to find the correct reports directory and ensures subfolder exists
func getOutputDir(subDir string) string {
	baseReports := "reports"

	// 1. Check if "reports" is in current directory
	if info, err := os.Stat("reports"); err == nil && info.IsDir() {
		baseReports = "reports"
	} else if info, err := os.Stat("../../reports"); err == nil && info.IsDir() {
		// 2. Check if running from cmd/digger (so reports is ../../reports)
		baseReports = "../../reports"
	} else {
		// 3. Default: Create reports in current directory if not found
		if err := os.Mkdir("reports", 0755); err != nil {
			// Fallback to current dir if mkdir fails
			baseReports = "."
		}
	}

	// Now handle the subfolder
	fullPath := fmt.Sprintf("%s/%s", baseReports, subDir)
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		// If we can't create the subfolder, just return the base reports dir
		return baseReports
	}

	return fullPath
}

// getSafeFilename checks if file exists and appends number if needed
// e.g. report.txt -> report1.txt -> report2.txt
func getSafeFilename(filename string) string {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return filename
	}

	// Split name and extension
	ext := ".txt"
	nameWithoutExt := strings.TrimSuffix(filename, ext)

	for i := 1; ; i++ {
		newFilename := fmt.Sprintf("%s%d%s", nameWithoutExt, i, ext)
		if _, err := os.Stat(newFilename); os.IsNotExist(err) {
			return newFilename
		}
		// Safety break to prevent infinite loops
		if i > 1000 {
			return fmt.Sprintf("%s_%d%s", nameWithoutExt, i, ext)
		}
	}
}
