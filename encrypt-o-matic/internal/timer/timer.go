package timer

import (
	"fmt"
	"time"
)

// ComputeUnlockTime returns the UTC unlock time based on duration in minutes.
func ComputeUnlockTime(durationMinutes int) time.Time {
	return time.Now().UTC().Add(time.Duration(durationMinutes) * time.Minute)
}

// IsUnlockExpired reports whether the stored unlock time has passed.
func IsUnlockExpired(unlockTime time.Time) bool {
	return time.Now().UTC().After(unlockTime) || time.Now().UTC().Equal(unlockTime)
}

// FormatUnlockStatus returns a human-readable timer status string.
func FormatUnlockStatus(unlockTime time.Time) string {
	now := time.Now().UTC()
	if IsUnlockExpired(unlockTime) {
		return fmt.Sprintf("Timer expired at %s — automatic decryption is allowed", unlockTime.Format(time.RFC3339))
	}
	remaining := unlockTime.Sub(now).Round(time.Second)
	return fmt.Sprintf("Locked until %s (%s remaining)", unlockTime.Format(time.RFC3339), remaining)
}
