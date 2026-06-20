package web

import (
	"io/fs"
	"strings"
	"testing"
)

func TestEmbeddedStaticAssets(t *testing.T) {
	root, err := fs.Sub(staticFS, "static")
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range []string{"index.html", "css/app.css", "js/app.js", "js/edu-content.js"} {
		data, err := fs.ReadFile(root, file)
		if err != nil {
			t.Fatalf("missing embedded file %q: %v", file, err)
		}
		if len(data) == 0 {
			t.Fatalf("embedded file %q is empty", file)
		}
	}

	css, _ := fs.ReadFile(root, "css/app.css")
	if !strings.Contains(string(css), "--bg:") {
		t.Fatal("css/app.css does not look like the dashboard stylesheet")
	}

	js, _ := fs.ReadFile(root, "js/app.js")
	if !strings.Contains(string(js), "renderDashboard") {
		t.Fatal("js/app.js does not look like the dashboard script")
	}

	edu, _ := fs.ReadFile(root, "js/edu-content.js")
	if !strings.Contains(string(edu), "window.EDU") {
		t.Fatal("js/edu-content.js does not contain educational content")
	}
}
