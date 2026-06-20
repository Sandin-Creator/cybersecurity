package web

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// browseEntry is a folder or file row in the path picker.
type browseEntry struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// browseResponse is returned by GET /api/browse.
type browseResponse struct {
	CurrentPath string        `json:"currentPath"`
	ParentPath  string        `json:"parentPath"`
	Folders     []browseEntry `json:"folders"`
	Files       []browseEntry `json:"files"`
	Roots       []browseEntry `json:"roots"`
}

// AllowedBrowseRoots returns reviewer-safe starting locations (no whole-disk scan).
func AllowedBrowseRoots() ([]browseEntry, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	roots := []browseEntry{
		{Name: "Current working directory", Path: cwd},
		{Name: "Test data (tests/testdata)", Path: filepath.Join(cwd, "tests", "testdata")},
	}

	if runtime.GOOS == "windows" {
		roots = append(roots, browseEntry{Name: "Demo folder (C:\\encrypt-demo)", Path: `C:\encrypt-demo`})
	} else {
		roots = append(roots, browseEntry{Name: "Demo folder (C:\\encrypt-demo)", Path: `C:\encrypt-demo`})
		if _, err := os.Stat("/mnt/c/encrypt-demo"); err == nil {
			roots = append(roots, browseEntry{Name: "Demo folder (WSL: /mnt/c/encrypt-demo)", Path: "/mnt/c/encrypt-demo"})
		}
	}

	return roots, nil
}

func allowedRootPaths() ([]string, error) {
	entries, err := AllowedBrowseRoots()
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0, len(entries))
	for _, e := range entries {
		if abs, err := filepath.Abs(e.Path); err == nil {
			paths = append(paths, abs)
		}
	}
	return paths, nil
}

func isPathUnderAllowedRoots(target string, roots []string) (string, bool) {
	abs, err := filepath.Abs(filepath.Clean(target))
	if err != nil {
		return "", false
	}
	for _, root := range roots {
		rootAbs, err := filepath.Abs(filepath.Clean(root))
		if err != nil {
			continue
		}
		if pathWithinRoot(rootAbs, abs) {
			return abs, true
		}
	}
	return "", false
}

func pathWithinRoot(root, target string) bool {
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return false
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return false
	}
	return true
}

// BrowseDirectory lists one directory level under allowed roots only.
func BrowseDirectory(requestedPath string) (*browseResponse, error) {
	roots, err := AllowedBrowseRoots()
	if err != nil {
		return nil, err
	}
	rootPaths, err := allowedRootPaths()
	if err != nil {
		return nil, err
	}

	resp := &browseResponse{Roots: roots, Folders: []browseEntry{}, Files: []browseEntry{}}

	startPath := requestedPath
	if strings.TrimSpace(startPath) == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		startPath = cwd
	}

	abs, ok := isPathUnderAllowedRoots(startPath, rootPaths)
	if !ok {
		return nil, fmt.Errorf("path not allowed: browsing is limited to demo locations and the working directory")
	}

	info, err := os.Stat(abs)
	if err != nil {
		return nil, fmt.Errorf("path not accessible: %w", err)
	}
	if !info.IsDir() {
		abs = filepath.Dir(abs)
	}

	resp.CurrentPath = abs

	parent := filepath.Dir(abs)
	if parentAbs, parentOK := isPathUnderAllowedRoots(parent, rootPaths); parentOK && parentAbs != abs {
		resp.ParentPath = parentAbs
	}

	entries, err := os.ReadDir(abs)
	if err != nil {
		return nil, fmt.Errorf("cannot read directory: %w", err)
	}

	var folders, files []browseEntry
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		full := filepath.Join(abs, name)
		entry := browseEntry{Name: name, Path: full}
		if e.IsDir() {
			folders = append(folders, entry)
			continue
		}
		if strings.EqualFold(filepath.Ext(name), ".exe") {
			files = append(files, entry)
		}
	}

	sortBrowseEntries(folders)
	sortBrowseEntries(files)
	resp.Folders = folders
	resp.Files = files
	return resp, nil
}

func sortBrowseEntries(entries []browseEntry) {
	sort.Slice(entries, func(i, j int) bool {
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})
}
