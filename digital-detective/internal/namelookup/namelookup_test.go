package namelookup

import (
	"strings"
	"testing"
)

func TestSearchByName_ExactMatch(t *testing.T) {
	name := "Dennis Laiho"
	_, imageMap := SearchByName(name, "", "")

	// We only need to check one URL to verify the behavior, as they all share the same logic
	found := false
	for url := range imageMap {
		found = true

		// Decode URL for easier checking is tricky without importing net/url everywhere
		// but checking encoded strings is safer for what we send to Google.

		// Expected: full name in quotes, URL encoded.

		// NOT Expected: double quotes

		if !strings.Contains(url, "%22Dennis+Laiho%22") {
			t.Errorf("URL does not contain correctly quoted exact match: %s", url)
		}

		if strings.Contains(url, "%22%22Dennis+Laiho%22%22") {
			t.Errorf("URL contains DOUBLE quoted match (bad): %s", url)
		}

		// Break after one check to avoid spamming errors if they all fail
		break
	}

	if !found {
		t.Error("No dork links were generated, cannot verify quoting.")
	}
}
