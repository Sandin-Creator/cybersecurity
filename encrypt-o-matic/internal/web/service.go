package web

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"encrypt-o-matic/internal/activity"
	"encrypt-o-matic/internal/config"
	"encrypt-o-matic/internal/debug"
	"encrypt-o-matic/internal/fileops"
	"encrypt-o-matic/internal/metadata"
	"encrypt-o-matic/internal/timer"
)

// DashboardData aggregates homepage statistics.
type DashboardData struct {
	MasterPasswordConfigured bool     `json:"masterPasswordConfigured"`
	EncryptedFileCount       int      `json:"encryptedFileCount"`
	BackupCount              int      `json:"backupCount"`
	MetadataCount            int      `json:"metadataCount"`
	Algorithms               []string `json:"algorithms"`
	RecentActivity           []activity.Entry `json:"recentActivity"`
	Stats                    DashboardStats   `json:"stats"`
	LastOperation            *time.Time       `json:"lastOperation,omitempty"`
}

type DashboardStats struct {
	TotalEncrypted int `json:"totalEncrypted"`
	TotalDecrypted int `json:"totalDecrypted"`
	TotalBackups   int `json:"totalBackups"`
	DataProcessed  int64 `json:"dataProcessedBytes"`
}

// FileRow is a summary row for the encrypted files table.
type FileRow struct {
	ID            string    `json:"id"`
	FileName      string    `json:"fileName"`
	Path          string    `json:"path"`
	Algorithm     string    `json:"algorithm"`
	OriginalSize  int64     `json:"originalSize"`
	CurrentSize   int64     `json:"currentSize"`
	PaddingSize   int64     `json:"paddingSize"`
	EncryptedAt   time.Time `json:"encryptedAt"`
	UnlockTime    time.Time `json:"unlockTime"`
	Remaining     string    `json:"remaining"`
	Status        string    `json:"status"`
	StatusClass   string    `json:"statusClass"`
}

// FileDetail is full detail for a single encrypted file.
type FileDetail struct {
	FileRow
	OriginalHash string   `json:"originalHash"`
	IntegrityOK  bool     `json:"integrityOk"`
	NonceHex     string   `json:"nonceHex"`
	SaltHex      string   `json:"saltHex"`
	Compressed   bool     `json:"compressed"`
	Backups      []string `json:"backups"`
}

func BuildDashboard() (DashboardData, error) {
	var dash DashboardData
	dash.Algorithms = []string{"AES", "ChaCha20", "Twofish"}

	hashPath, err := config.MasterHashPath()
	if err != nil {
		return dash, err
	}
	if _, err := os.Stat(hashPath); err == nil {
		dash.MasterPasswordConfigured = true
	}

	metaList, err := metadata.ListAll()
	if err != nil {
		return dash, err
	}
	dash.EncryptedFileCount = len(metaList)
	dash.MetadataCount = len(metaList)

	backups, err := debug.ListBackupFiles()
	if err != nil {
		return dash, err
	}
	dash.BackupCount = len(backups)
	dash.Stats.TotalBackups = len(backups)

	enc, dec := activity.Stats()
	dash.Stats.TotalEncrypted = enc
	dash.Stats.TotalDecrypted = dec

	var processed int64
	for _, m := range metaList {
		if m.OriginalSize > 0 {
			processed += m.OriginalSize
		}
	}
	dash.Stats.DataProcessed = processed

	recent, err := activity.List(10)
	if err != nil {
		return dash, err
	}
	if recent == nil {
		recent = []activity.Entry{}
	}
	dash.RecentActivity = recent
	if len(recent) > 0 {
		t := recent[0].Timestamp
		dash.LastOperation = &t
	}

	return dash, nil
}

func ListEncryptedFiles() ([]FileRow, error) {
	metaList, err := metadata.ListAll()
	if err != nil {
		return nil, err
	}

	var rows []FileRow
	for _, m := range metaList {
		row, err := fileRowFromMeta(m)
		if err != nil {
			continue
		}
		rows = append(rows, row)
	}
	if rows == nil {
		rows = []FileRow{}
	}
	return rows, nil
}

func GetFileDetail(key string) (*FileDetail, error) {
	meta, err := metadata.LoadByKey(key)
	if err != nil {
		return nil, err
	}

	row, err := fileRowFromMeta(*meta)
	if err != nil {
		return nil, err
	}

	backups, _ := fileops.ListBackupsForFile(meta.OriginalPath)
	if backups == nil {
		backups = []string{}
	}

	detail := &FileDetail{
		FileRow:      row,
		OriginalHash: meta.OriginalHash,
		IntegrityOK:  metadata.IsEncrypted(meta.OriginalPath),
		NonceHex:     hex.EncodeToString(meta.Nonce),
		SaltHex:      hex.EncodeToString(meta.Salt),
		Compressed:   meta.Compressed,
		Backups:      backups,
	}
	return detail, nil
}

func fileRowFromMeta(m metadata.FileMetadata) (FileRow, error) {
	id, err := metadata.PathKey(m.OriginalPath)
	if err != nil {
		return FileRow{}, err
	}

	var currentSize int64
	if info, err := os.Stat(m.OriginalPath); err == nil {
		currentSize = info.Size()
	}

	originalSize := m.OriginalSize
	if originalSize == 0 {
		originalSize = currentSize - m.PaddingSize - int64(len(config.EncryptedMagic))
		if originalSize < 0 {
			originalSize = 0
		}
	}

	status, class := fileStatus(m)

	remaining := timer.FormatUnlockStatus(m.UnlockTime)
	if timer.IsUnlockExpired(m.UnlockTime) {
		remaining = "Unlocked (timer expired)"
	}

	return FileRow{
		ID:           id,
		FileName:     filepath.Base(m.OriginalPath),
		Path:         m.OriginalPath,
		Algorithm:    m.Algorithm,
		OriginalSize: originalSize,
		CurrentSize:  currentSize,
		PaddingSize:  m.PaddingSize,
		EncryptedAt:  m.EncryptedAt,
		UnlockTime:   m.UnlockTime,
		Remaining:    remaining,
		Status:       status,
		StatusClass:  class,
	}, nil
}

func fileStatus(m metadata.FileMetadata) (label, class string) {
	if _, err := metadata.Load(m.OriginalPath); err != nil {
		return "Integrity issue", "red"
	}
	if timer.IsUnlockExpired(m.UnlockTime) {
		return "Unlocked", "green"
	}
	return "Encrypted", "orange"
}

func DebugSnapshot() (map[string]interface{}, error) {
	root, err := config.RootDir()
	if err != nil {
		return nil, err
	}
	hashPath, err := config.MasterHashPath()
	if err != nil {
		return nil, err
	}

	hashExists := false
	if _, err := os.Stat(hashPath); err == nil {
		hashExists = true
	}

	metaFiles, err := debug.ListMetadataFiles()
	if err != nil {
		return nil, err
	}
	backupFiles, err := debug.ListBackupFiles()
	if err != nil {
		return nil, err
	}
	if metaFiles == nil {
		metaFiles = []string{}
	}
	if backupFiles == nil {
		backupFiles = []string{}
	}

	return map[string]interface{}{
		"configDir":   root,
		"hashPath":    hashPath,
		"hashExists":  hashExists,
		"metadata":    metaFiles,
		"backups":     backupFiles,
	}, nil
}

func BuildInfo() map[string]string {
	return map[string]string{
		"goVersion": runtime.Version(),
		"os":        runtime.GOOS + "/" + runtime.GOARCH,
		"buildTime": time.Now().UTC().Format(time.RFC3339),
	}
}

func MetadataJSON(key string) (string, error) {
	meta, err := metadata.LoadByKey(key)
	if err != nil {
		return "", err
	}
	path, err := metadataPathForDisplay(meta.OriginalPath)
	if err != nil {
		return "", err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func metadataPathForDisplay(targetPath string) (string, error) {
	abs, err := filepath.Abs(targetPath)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(abs))
	name := hex.EncodeToString(sum[:]) + ".json"
	metaDir, err := config.MetadataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(metaDir, name), nil
}

func FormatBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for n/div >= unit && exp < 4 {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGT"[exp])
}
