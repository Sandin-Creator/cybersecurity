package iplookup

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type IPResponse struct {
	Status      string  `json:"status"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Zip         string  `json:"zip"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	ISP         string  `json:"isp"`
	Org         string  `json:"org"`
	AS          string  `json:"as"`
	Query       string  `json:"query"`
}

// SearchByIP performs a geolocation lookup for the given IP using multiple sources
func SearchByIP(ip string) (string, map[string]string) {
	// 0. Enhanced Private IP Handling: Map to Public Identity
	// If the user searches for a local/private IP, they usually want to know "Where is this network?"
	if strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "10.") || strings.HasPrefix(ip, "127.") {
		// Fetch Public IP
		pubIP, err := getPublicIP()
		if err == nil && pubIP != "" {
			return SearchByIP(pubIP) // Recursively search the Public IP
		}
		// Fallback if public fetch fails
		return fmt.Sprintf("IP Search Results for %s:\n\n[!] PRIVATE/LOCAL IP ADDRESS DETECTED\nThis IP is on a local network (LAN).\nCould not auto-resolve public IP: %v", ip, err), nil
	}

	imageMap := make(map[string]string)

	// Source 1: ip-api.com
	url1 := fmt.Sprintf("http://ip-api.com/json/%s", ip)

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	var result1 string
	resp1, err := client.Get(url1)
	if err != nil {
		result1 = fmt.Sprintf("Source 1 Error: %v", err)
	} else {
		defer resp1.Body.Close()
		if resp1.StatusCode != http.StatusOK {
			result1 = fmt.Sprintf("Source 1 API failed: %d", resp1.StatusCode)
		} else {
			var data1 IPResponse
			if err := json.NewDecoder(resp1.Body).Decode(&data1); err != nil {
				result1 = fmt.Sprintf("Source 1 Decode Error: %v", err)
			} else if data1.Status == "fail" {
				result1 = fmt.Sprintf("Source 1 Query Failed for %s", ip)
			} else {
				result1 = fmt.Sprintf("Source 1 (ip-api.com):\nISP: %s\nLocation: %s, %s (%f, %f)\nOrganization: %s",
					data1.ISP, data1.City, data1.RegionName, data1.Lat, data1.Lon, data1.Org)

				// Generate Map Link and Flag
				if data1.Lat != 0 && data1.Lon != 0 {
					mapURL := fmt.Sprintf("https://www.google.com/maps?q=%f,%f", data1.Lat, data1.Lon)
					result1 += fmt.Sprintf("\nMaps: %s", mapURL)

					if data1.CountryCode != "" {
						flagURL := fmt.Sprintf("https://flagcdn.com/w320/%s.png", strings.ToLower(data1.CountryCode))
						imageMap[mapURL] = flagURL
					}
				}
			}
		}
	}

	// Source 2: ipapi.co
	url2 := fmt.Sprintf("https://ipapi.co/%s/json/", ip)
	var result2 string

	req2, err := http.NewRequest("GET", url2, nil)
	if err == nil {
		req2.Header.Set("User-Agent", "data-digger-cli")
		resp2, err := client.Do(req2)
		if err != nil {
			result2 = fmt.Sprintf("Source 2 Error: %v", err)
		} else {
			defer resp2.Body.Close()
			if resp2.StatusCode != http.StatusOK {
				result2 = fmt.Sprintf("Source 2 API failed: %d", resp2.StatusCode)
			} else {
				var data2 map[string]interface{}
				if err := json.NewDecoder(resp2.Body).Decode(&data2); err != nil {
					result2 = fmt.Sprintf("Source 2 Decode Error: %v", err)
				} else {
					city, _ := data2["city"].(string)
					region, _ := data2["region"].(string)
					org, _ := data2["org"].(string)

					result2 = fmt.Sprintf("Source 2 (ipapi.co):\nLocation: %s, %s\nOrganization: %s",
						city, region, org)
				}
			}
		}
	} else {
		result2 = fmt.Sprintf("Source 2 Request Init Error: %v", err)
	}

	// Source 3: ipwho.is (Free, no key)
	url3 := fmt.Sprintf("https://ipwho.is/%s", ip)
	var result3 string
	resp3, err := client.Get(url3)
	if err == nil {
		defer resp3.Body.Close()
		if resp3.StatusCode == 200 {
			var data3 map[string]interface{}
			if err := json.NewDecoder(resp3.Body).Decode(&data3); err == nil {
				if success, ok := data3["success"].(bool); ok && success {
					// Safe Type Assertion
					var isp string
					if conn, ok := data3["connection"].(map[string]interface{}); ok {
						if v, ok := conn["isp"].(string); ok {
							isp = v
						}
					}
					result3 = fmt.Sprintf("Source 3 (ipwho.is):\nLocation: %s, %s, %s\nISP: %s",
						data3["city"], data3["region"], data3["country"], isp)
				} else {
					result3 = fmt.Sprintf("Source 3 Failed: %v", data3["message"])
				}
			}
		}
	}

	return fmt.Sprintf("IP Search Results for %s:\n\n%s\n\n%s\n\n%s", ip, result1, result2, result3), imageMap
}

// getPublicIP fetches the current network's external IP address
func getPublicIP() (string, error) {
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://api.ipify.org?format=text")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
