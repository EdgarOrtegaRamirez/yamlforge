package lint

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/EdgarOrtegaRamirez/yamlforge/internal/lint"
)

func writeTempYAML(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	return path
}

func TestLint_ValidYAML(t *testing.T) {
	path := writeTempYAML(t, `name: test
version: 1.0
`)
	engine := lint.NewEngine()
	report, err := engine.LintFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.HasErrors() {
		t.Fatalf("expected no errors, got %d", report.ErrorCount())
	}
}

func TestLint_DuplicateKeys(t *testing.T) {
	path := writeTempYAML(t, `name: first
name: second
`)
	engine := lint.NewEngine()
	report, err := engine.LintFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, issue := range report.Issues {
		if issue.Rule == "no-duplicate-keys" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected duplicate key issue")
	}
}

func TestLint_TrailingSpaces(t *testing.T) {
	path := writeTempYAML(t, "name: test   \nversion: 1\n")
	engine := lint.NewEngine()
	report, err := engine.LintFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, issue := range report.Issues {
		if issue.Rule == "no-trailing-spaces" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected trailing spaces issue")
	}
}

func TestLint_TruthyValues(t *testing.T) {
	path := writeTempYAML(t, `enabled: yes
disabled: no
`)
	engine := lint.NewEngine()
	report, err := engine.LintFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, issue := range report.Issues {
		if issue.Rule == "truthy-values" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected truthy values issue")
	}
}

func TestLint_InvalidYAML(t *testing.T) {
	path := writeTempYAML(t, `invalid: yaml: content: [unclosed`)
	engine := lint.NewEngine()
	report, err := engine.LintFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !report.HasErrors() {
		t.Fatal("expected parse error")
	}
}

func TestLint_LintBytes(t *testing.T) {
	engine := lint.NewEngine()
	report, err := engine.LintBytes([]byte(`name: test`), "test.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.HasErrors() {
		t.Fatalf("expected no errors, got %d", report.ErrorCount())
	}
}
