package custom

import (
	"crypto/sha256"
	"fmt"
	"strconv"
	"time"
)

// ParseRange parses a range like "0-100000" into start and end (inclusive).
func ParseRange(input string) (int, int, error) {
	dash := -1
	for i, c := range input {
		if c == '-' {
			dash = i
			break
		}
	}
	if dash <= 0 || dash >= len(input)-1 {
		return 0, 0, fmt.Errorf("invalid custom variable: expected format like 0-100000")
	}

	start, err := strconv.Atoi(input[:dash])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid custom variable: start must be an integer")
	}
	end, err := strconv.Atoi(input[dash+1:])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid custom variable: end must be an integer")
	}
	if start < 0 || end < start {
		return 0, 0, fmt.Errorf("invalid custom variable: range must satisfy 0 <= start <= end")
	}

	return start, end, nil
}

// RunOperation performs a safe, time-consuming SHA-256 loop over the range.
func RunOperation(customRange string) error {
	start, end, err := ParseRange(customRange)
	if err != nil {
		return err
	}

	total := end - start + 1
	if total == 0 {
		return nil
	}

	fmt.Printf("Running custom operation over %d values (%d-%d)...\n", total, start, end)

	const progressEvery = 1000
	const yieldEvery = 5000
	lastReport := time.Now()

	for i := start; i <= end; i++ {
		value := strconv.Itoa(i)
		_ = sha256.Sum256([]byte(value))

		if (i-start+1)%progressEvery == 0 || i == end {
			if time.Since(lastReport) >= 250*time.Millisecond || i == end {
				pct := float64(i-start+1) / float64(total) * 100
				fmt.Printf("  Custom operation progress: %d/%d (%.1f%%)\n", i-start+1, total, pct)
				lastReport = time.Now()
			}
		}

		if (i-start+1)%yieldEvery == 0 {
			time.Sleep(time.Millisecond)
		}
	}

	fmt.Println("Custom operation complete.")
	return nil
}
