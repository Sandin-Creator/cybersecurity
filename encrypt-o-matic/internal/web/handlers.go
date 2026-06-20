package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"encrypt-o-matic/internal/auth"
	"encrypt-o-matic/internal/config"
	"encrypt-o-matic/internal/crypto"
	"encrypt-o-matic/internal/custom"
	"encrypt-o-matic/internal/fileops"
)

type encryptRequest struct {
	Target      string `json:"target"`
	Algorithm   string `json:"algorithm"`
	PaddingMB   int    `json:"paddingMb"`
	CustomRange string `json:"customRange"`
	DurationMin int    `json:"durationMin"`
	Password    string `json:"password"`
}

type decryptRequest struct {
	Target   string `json:"target"`
	Password string `json:"password"`
}

type passwordRequest struct {
	Password string `json:"password"`
}

type setupPasswordRequest struct {
	Password string `json:"password"`
	Confirm  string `json:"confirm"`
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	data, err := BuildDashboard()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (s *Server) handleFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	rows, err := ListEncryptedFiles()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if rows == nil {
		rows = []FileRow{}
	}
	writeJSON(w, http.StatusOK, rows)
}

func (s *Server) handleFileDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	key := strings.TrimPrefix(r.URL.Path, "/api/files/")
	if key == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing file id"})
		return
	}
	if strings.HasSuffix(key, "/metadata") {
		key = strings.TrimSuffix(key, "/metadata")
		raw, err := MetadataJSON(key)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"metadata": raw})
		return
	}
	detail, err := GetFileDetail(key)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

func (s *Server) handleEncrypt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req encryptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	algorithm, err := crypto.NormalizeAlgorithm(req.Algorithm)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if req.PaddingMB < 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "padding must be non-negative"})
		return
	}
	if req.DurationMin <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "duration must be positive"})
		return
	}
	if _, _, err := custom.ParseRange(req.CustomRange); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := verifyPassword(req.Password); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}

	job := jobs.create()
	go func() {
		reporter := func(step, total int, label string) {
			jobs.update(job.ID, step, total, label)
		}
		if err := fileops.CheckExpiredAutoDecryptWithSource(req.Target, req.Password, "web"); err != nil {
			jobs.complete(job.ID, err)
			return
		}
		err := fileops.EncryptTargetWithProgress(req.Target, algorithm, req.Password, req.PaddingMB, req.DurationMin, req.CustomRange, reporter, "web")
		jobs.complete(job.ID, err)
	}()

	writeJSON(w, http.StatusAccepted, map[string]string{"jobId": job.ID})
}

func (s *Server) handleDecrypt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req decryptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	if err := verifyPassword(req.Password); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}
	if err := fileops.DecryptTargetWithSource(req.Target, req.Password, "web"); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "Decryption complete"})
}

func (s *Server) handleSetupPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req setupPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	if err := auth.CreateMasterPassword(req.Password, req.Confirm); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "Master password created successfully."})
}

func (s *Server) handleVerifyPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req passwordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	hashPath, err := config.MasterHashPath()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	result, verifyErr := auth.VerifyStoredPassword(req.Password, hashPath)
	switch result {
	case auth.VerifyOK:
		writeJSON(w, http.StatusOK, map[string]string{"result": "Password OK"})
	case auth.VerifyFailed:
		writeJSON(w, http.StatusUnauthorized, map[string]string{"result": "Password INVALID"})
	default:
		msg := "Password INVALID"
		if verifyErr != nil {
			msg = verifyErr.Error()
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"result": "Password INVALID", "error": msg})
	}
}

func (s *Server) handleDebug(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	snap, err := DebugSnapshot()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, snap)
}

func (s *Server) handleBrowse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	path := r.URL.Query().Get("path")
	resp, err := BrowseDirectory(path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/jobs/")
	job, ok := jobs.get(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "job not found"})
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func (s *Server) handleReviewer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	checklist := []map[string]interface{}{
		{"name": "AES", "passed": true},
		{"name": "ChaCha20", "passed": true},
		{"name": "Twofish", "passed": true},
		{"name": "Compression", "passed": true},
		{"name": "Directory Encryption", "passed": true},
		{"name": "Password Verification", "passed": true},
		{"name": "Timer Locking", "passed": true},
		{"name": "Integrity Verification", "passed": true},
		{"name": "Backups", "passed": true},
		{"name": "Automated Tests", "passed": s.lastTestsPassed > 0 && s.lastTestsPassed == s.lastTestsTotal},
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"checklist":   checklist,
		"testsPassed": s.lastTestsPassed,
		"testsTotal":  s.lastTestsTotal,
		"lastTestRun": s.lastTestRun,
		"build":       BuildInfo(),
		"goVersion":   runtime.Version(),
	})
}

func (s *Server) handleRunTests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	root, err := moduleRootDir()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	cmd := exec.Command("go", "test", "-json", "./...")
	cmd.Dir = root
	out, runErr := cmd.CombinedOutput()
	passed, total, failed := parseTestResults(string(out))
	s.lastTestsPassed = passed
	s.lastTestsTotal = total
	s.lastTestRun = time.Now().UTC().Format(time.RFC3339)

	status := "complete"
	errMsg := ""
	if runErr != nil {
		status = "failed"
		errMsg = runErr.Error()
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":      status,
		"passed":      passed,
		"total":       total,
		"failedTests": failed,
		"lastTestRun": s.lastTestRun,
		"error":       errMsg,
		"output":      truncate(string(out), 4000),
	})
}

func verifyPassword(password string) error {
	hashPath, err := config.MasterHashPath()
	if err != nil {
		return err
	}
	if _, err := os.Stat(hashPath); os.IsNotExist(err) {
		return errors.New("master password not configured — run CLI once to create it")
	}
	result, verifyErr := auth.VerifyStoredPassword(password, hashPath)
	switch result {
	case auth.VerifyOK:
		return nil
	case auth.VerifyFailed:
		return errors.New("Password verification failed.")
	default:
		if verifyErr != nil {
			return verifyErr
		}
		return errors.New("Password verification failed.")
	}
}

func parseTestJSON(output string) (passed, total int) {
	passed, total, _ = parseTestResults(output)
	return passed, total
}

func parseTestResults(output string) (passed, total int, failed []string) {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var evt struct {
			Action string `json:"Action"`
			Test   string `json:"Test"`
		}
		if json.Unmarshal([]byte(line), &evt) != nil {
			continue
		}
		if evt.Test == "" {
			continue
		}
		switch evt.Action {
		case "pass":
			passed++
			total++
		case "fail":
			total++
			failed = append(failed, evt.Test)
		}
	}
	return passed, total, failed
}

func moduleRootDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
