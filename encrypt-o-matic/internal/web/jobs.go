package web

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// Job tracks an in-progress encrypt operation for the UI.
type Job struct {
	ID        string    `json:"id"`
	Step      int       `json:"step"`
	Total     int       `json:"total"`
	Label     string    `json:"label"`
	Status    string    `json:"status"`
	Error     string    `json:"error,omitempty"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type jobManager struct {
	mu   sync.RWMutex
	jobs map[string]*Job
}

var jobs = &jobManager{jobs: make(map[string]*Job)}

func newJobID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (m *jobManager) create() *Job {
	j := &Job{
		ID:        newJobID(),
		Status:    "running",
		Total:     5,
		UpdatedAt: time.Now().UTC(),
	}
	m.mu.Lock()
	m.jobs[j.ID] = j
	m.mu.Unlock()
	return j
}

func (m *jobManager) update(id string, step int, total int, label string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if j, ok := m.jobs[id]; ok {
		j.Step = step
		j.Total = total
		j.Label = label
		j.UpdatedAt = time.Now().UTC()
	}
}

func (m *jobManager) complete(id string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	j, ok := m.jobs[id]
	if !ok {
		return
	}
	if err != nil {
		j.Status = "error"
		j.Error = err.Error()
	} else {
		j.Status = "complete"
		j.Step = j.Total
		j.Label = "Encryption complete"
	}
	j.UpdatedAt = time.Now().UTC()
}

func (m *jobManager) get(id string) (*Job, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	j, ok := m.jobs[id]
	if !ok {
		return nil, false
	}
	copy := *j
	return &copy, true
}
