package agentreliability_test

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

// runExample executes `go run ./examples/<dir>` from the module root and returns
// combined output. It proves each demo really compiles and prints the expected
// decision — the "clone and run in 5 minutes" guarantee.
func runExample(t *testing.T, dir string) string {
	t.Helper()
	cmd := exec.Command("go", "run", "./examples/"+dir)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("go run ./examples/%s failed: %v\n%s", dir, err, out.String())
	}
	return out.String()
}

func TestExampleBasic(t *testing.T) {
	out := runExample(t, "basic")
	if !strings.Contains(out, "DENY") {
		t.Fatalf("basic should print DENY for rm -rf /, got:\n%s", out)
	}
}

func TestExampleFileProtection(t *testing.T) {
	out := runExample(t, "file-protection")
	if !strings.Contains(out, "DENY") || !strings.Contains(out, "untouched") {
		t.Fatalf("file-protection should DENY and leave fs untouched, got:\n%s", out)
	}
}

func TestExampleGitProtection(t *testing.T) {
	out := runExample(t, "git-protection")
	if !strings.Contains(out, "DENY") || !strings.Contains(out, "protected") {
		t.Fatalf("git-protection should DENY force push, got:\n%s", out)
	}
}

func TestExampleDatabaseApproval(t *testing.T) {
	out := runExample(t, "database-approval")
	if !strings.Contains(out, "ASK") {
		t.Fatalf("database-approval should ASK for prod write, got:\n%s", out)
	}
}

func TestExampleModify(t *testing.T) {
	out := runExample(t, "modify")
	if !strings.Contains(out, "MODIFY") || !strings.Contains(out, "instead") {
		t.Fatalf("modify should MODIFY and suggest a safe target, got:\n%s", out)
	}
}
