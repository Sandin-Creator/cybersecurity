package usernamelookup

import (
	"fmt"
	"strings"
)

// DatingSignal represents a detected potential indicator
type DatingSignal struct {
	Level       string // "Unlikely", "Possible", "Likely"
	Description string
	Points      int
}

// AnalyzeDatingSignals correlates findings from username search to infer dating app usage
// AnalyzeDatingSignals correlates findings from username search to infer dating app usage
func AnalyzeDatingSignals(results map[string]string) string {
	var sb strings.Builder
	var signals []DatingSignal
	totalScore := 0

	sb.WriteString("============================================================\n")
	sb.WriteString("| DATING APP INVESTIGATOR (OSINT Inference)                |\n")
	sb.WriteString("============================================================\n")
	sb.WriteString("Note: This tool checks for *public* signals (bios, links) \n")
	sb.WriteString("that suggest dating app usage. It does NOT query dating sites API.\n\n")

	// 1. Platform Presence Analysis
	hasSpotify := false
	hasInstagram := false
	hasSnapchat := false
	hasTinder := false
	hasBadoo := false
	hasOkCupid := false

	// Check what we found
	for platform := range results {
		switch platform {
		case "Spotify":
			hasSpotify = true
		case "Instagram":
			hasInstagram = true
		case "Snapchat":
			hasSnapchat = true
		case "Tinder":
			hasTinder = true
		case "Badoo":
			hasBadoo = true
		case "OkCupid":
			hasOkCupid = true
		}
	}

	// DIRECT DATING APPS (High Confidence)
	if hasTinder {
		signals = append(signals, DatingSignal{Level: "Likely", Description: "Public Tinder profile found", Points: 10})
		totalScore += 10
	}
	if hasBadoo {
		signals = append(signals, DatingSignal{Level: "Likely", Description: "Public Badoo profile found", Points: 10})
		totalScore += 10
	}
	if hasOkCupid {
		signals = append(signals, DatingSignal{Level: "Likely", Description: "Public OkCupid profile found", Points: 10})
		totalScore += 10
	}

	// INDIRECT SIGNALS (Lower Confidence)
	if hasSpotify {
		// Downgraded from Possible to Weak
		signals = append(signals, DatingSignal{
			Level:       "Weak",
			Description: "Spotify account found (common on dating bios, but generic)",
			Points:      1,
		})
		totalScore += 1
	}

	if hasSnapchat {
		signals = append(signals, DatingSignal{
			Level:       "Possible",
			Description: "Snapchat account found (often used for private messaging)",
			Points:      2,
		})
		totalScore += 2
	}

	// COMBINATIONS
	if hasInstagram && hasTinder {
		signals = append(signals, DatingSignal{
			Level:       "Likely",
			Description: "Potential Account Linkage: Instagram matches Tinder username",
			Points:      5,
		})
		totalScore += 5
	}

	if hasInstagram && hasSpotify {
		// Only worth point if we didn't already find tinder, it's a weak signal
		signals = append(signals, DatingSignal{
			Level:       "Weak",
			Description: "Instagram + Spotify combination (common pattern, but weak)",
			Points:      1,
		})
		totalScore += 1
	}

	// 2. Keyword/Content Analysis
	for platform, content := range results {
		// Fix: Pinterest HTML contains CSS variables like "--swipe-easing" and "--container-snap"
		// which trigger false positives. We skip keyword analysis for Pinterest.
		if platform == "Pinterest" {
			continue
		}

		contentLower := strings.ToLower(content)

		// Keywords
		keywords := map[string]int{
			"tinder":      5,
			"bumble":      5,
			"hinge":       5,
			"grindr":      5,
			"single":      4, // Increased weight
			"looking for": 4, // "looking for friends/dates"
			"swipe":       3,
			"dm me":       3,
			"snap:":       2,
			"sc:":         2,
		}

		// Negative Keywords
		if strings.Contains(contentLower, "married") || strings.Contains(contentLower, "husband") || strings.Contains(contentLower, "wife") {
			signals = append(signals, DatingSignal{Level: "Contra-indicator", Description: "Bio suggests relationship ('married/husband/wife')", Points: -5})
			totalScore -= 5
		}

		for word, points := range keywords {
			if strings.Contains(contentLower, word) {
				signals = append(signals, DatingSignal{
					Level:       "Likely",
					Description: fmt.Sprintf("Found keyword '%s' on %s", word, platform),
					Points:      points,
				})
				totalScore += points
			}
		}

		// Link Detection
		if strings.Contains(contentLower, "linktr.ee") || strings.Contains(contentLower, "carrd.co") {
			signals = append(signals, DatingSignal{Level: "Possible", Description: fmt.Sprintf("Link-in-bio tool found on %s (check for dating links)", platform), Points: 2})
			totalScore += 2
		}
	}

	// 3. Calculate Confidence
	var confidence string
	if totalScore >= 8 {
		confidence = "LIKELY"
	} else if totalScore >= 4 {
		confidence = "POSSIBLE"
	} else {
		confidence = "UNLIKELY"
	}

	// 4. Output Generation
	// Visualization: ASCII Bar
	scoreVis := totalScore
	if scoreVis > 10 {
		scoreVis = 10
	}
	if scoreVis < 0 {
		scoreVis = 0
	}
	barLength := 10
	filled := scoreVis
	empty := barLength - filled
	visualBar := "[" + strings.Repeat("█", filled) + strings.Repeat("░", empty) + "]"

	sb.WriteString(fmt.Sprintf("Composite Score: %d/10 %s\n", totalScore, visualBar))
	sb.WriteString(fmt.Sprintf("Inference:       %s\n", confidence))
	sb.WriteString("--------------------------------------------------\n")

	if len(signals) == 0 {
		sb.WriteString("[-] No public signals detected.\n")
	} else {
		for _, s := range signals {
			sb.WriteString(fmt.Sprintf("[!] %s (%+d): %s\n", s.Level, s.Points, s.Description))
		}
	}

	return sb.String()
}
