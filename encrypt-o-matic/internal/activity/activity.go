package activity

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"encrypt-o-matic/internal/config"
)

// Type describes an operation category.
const (
	TypeEncrypt = "encrypt"
	TypeDecrypt = "decrypt"
)

// Entry records a completed operation for dashboard history.
type Entry struct {
	Type      string    `json:"type"`
	Path      string    `json:"path"`
	Algorithm string    `json:"algorithm,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
}

type logFile struct {
	Entries []Entry `json:"entries"`
}

var mu sync.Mutex

func logPath() (string, error) {
	root, err := config.RootDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "activity.json"), nil
}

// Record appends an activity entry (keeps last 200 entries).
func Record(entryType, path, algorithm, source string) {
	mu.Lock()
	defer mu.Unlock()

	pathFile, err := logPath()
	if err != nil {
		return
	}

	var data logFile
	if raw, err := os.ReadFile(pathFile); err == nil {
		_ = json.Unmarshal(raw, &data)
	}

	data.Entries = append(data.Entries, Entry{
		Type:      entryType,
		Path:      path,
		Algorithm: algorithm,
		Timestamp: time.Now().UTC(),
		Source:    source,
	})

	if len(data.Entries) > 200 {
		data.Entries = data.Entries[len(data.Entries)-200:]
	}

	_ = os.MkdirAll(filepath.Dir(pathFile), 0o700)
	encoded, _ := json.MarshalIndent(data, "", "  ")
	_ = os.WriteFile(pathFile, encoded, 0o600)
}

// List returns activity entries newest first.
func List(limit int) ([]Entry, error) {
	pathFile, err := logPath()
	if err != nil {
		return nil, err
	}

	raw, err := os.ReadFile(pathFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var data logFile
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}

	entries := data.Entries
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}

	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}
	return entries, nil
}

// Stats returns aggregate counts from the activity log.
func Stats() (encrypted, decrypted int) {
	pathFile, err := logPath()
	if err != nil {
		return 0, 0
	}
	raw, err := os.ReadFile(pathFile)
	if err != nil {
		return 0, 0
	}
	var data logFile
	if err := json.Unmarshal(raw, &data); err != nil {
		return 0, 0
	}
	for _, e := range data.Entries {
		switch e.Type {
		case TypeEncrypt:
			encrypted++
		case TypeDecrypt:
			decrypted++
		}
	}
	return encrypted, decrypted
}
