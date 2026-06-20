package web

import "testing"

func TestParseTestJSONIgnoresPackageEvents(t *testing.T) {
	output := `{"Action":"run","Test":"TestA","Package":"pkg"}
{"Action":"pass","Test":"TestA","Package":"pkg"}
{"Action":"pass","Package":"pkg"}
{"Action":"run","Test":"TestB","Package":"pkg"}
{"Action":"fail","Test":"TestB","Package":"pkg"}
{"Action":"fail","Package":"pkg"}`

	passed, total, failed := parseTestResults(output)
	if passed != 1 || total != 2 {
		t.Fatalf("got passed=%d total=%d, want 1/2", passed, total)
	}
	if len(failed) != 1 || failed[0] != "TestB" {
		t.Fatalf("unexpected failed tests: %#v", failed)
	}
}

func TestParseTestJSONSubtests(t *testing.T) {
	output := `{"Action":"pass","Test":"TestParent/sub","Package":"pkg"}
{"Action":"pass","Test":"TestParent","Package":"pkg"}
{"Action":"pass","Package":"pkg"}`

	passed, total, _ := parseTestResults(output)
	if passed != 2 || total != 2 {
		t.Fatalf("got passed=%d total=%d, want 2/2", passed, total)
	}
}
