package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func buildBinary(t *testing.T) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "yamlforge")

	// Find project root by walking up from current directory
	projectRoot := ""
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			projectRoot = dir
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find project root (no go.mod found)")
		}
		dir = parent
	}

	cmd := exec.Command("go", "build", "-o", binary, "./cmd/yamlforge")
	cmd.Dir = projectRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	return binary
}

func writeTempFile(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	return path
}

func TestCLI_Version(t *testing.T) {
	binary := buildBinary(t)
	out, err := exec.Command(binary, "version").CombinedOutput()
	if err != nil {
		t.Fatalf("version command failed: %v\n%s", err, out)
	}
	if len(out) == 0 {
		t.Fatal("expected version output")
	}
}

func TestCLI_Lint(t *testing.T) {
	binary := buildBinary(t)
	path := writeTempFile(t, "test.yaml", `name: test
version: 1.0
`)
	out, err := exec.Command(binary, "lint", path).CombinedOutput()
	if err != nil {
		t.Fatalf("lint command failed: %v\n%s", err, out)
	}
}

func TestCLI_Lint_DuplicateKeys(t *testing.T) {
	binary := buildBinary(t)
	path := writeTempFile(t, "test.yaml", `name: first
name: second
`)
	cmd := exec.Command(binary, "lint", path)
	out, _ := cmd.CombinedOutput()
	s := string(out)
	if len(s) == 0 {
		t.Fatal("expected lint output for duplicate keys")
	}
}

func TestCLI_Fmt(t *testing.T) {
	binary := buildBinary(t)
	path := writeTempFile(t, "test.yaml", `name:   test
version:   1.0
`)
	out, err := exec.Command(binary, "fmt", path).CombinedOutput()
	if err != nil {
		t.Fatalf("fmt command failed: %v\n%s", err, out)
	}
	if len(out) == 0 {
		t.Fatal("expected format output")
	}
}

func TestCLI_Query(t *testing.T) {
	binary := buildBinary(t)
	path := writeTempFile(t, "test.yaml", `name: test
version: 1.0
`)
	out, err := exec.Command(binary, "query", path, "name").CombinedOutput()
	if err != nil {
		t.Fatalf("query command failed: %v\n%s", err, out)
	}
	if string(out) != "test\n" {
		t.Fatalf("expected 'test\\n', got %q", string(out))
	}
}

func TestCLI_Diff(t *testing.T) {
	binary := buildBinary(t)
	dir := t.TempDir()
	path1 := filepath.Join(dir, "a.yaml")
	path2 := filepath.Join(dir, "b.yaml")
	os.WriteFile(path1, []byte(`name: old`), 0644)
	os.WriteFile(path2, []byte(`name: new`), 0644)
	out, err := exec.Command(binary, "diff", path1, path2).CombinedOutput()
	if err != nil {
		t.Fatalf("diff command failed: %v\n%s", err, out)
	}
	if len(out) == 0 {
		t.Fatal("expected diff output")
	}
}

func TestCLI_Convert_JSON(t *testing.T) {
	binary := buildBinary(t)
	path := writeTempFile(t, "test.yaml", `name: test
version: 1.0
`)
	out, err := exec.Command(binary, "convert", path, "--to", "json").CombinedOutput()
	if err != nil {
		t.Fatalf("convert command failed: %v\n%s", err, out)
	}
	if len(out) == 0 {
		t.Fatal("expected JSON output")
	}
}

func TestCLI_Stats(t *testing.T) {
	binary := buildBinary(t)
	path := writeTempFile(t, "test.yaml", `name: test
version: 1.0
`)
	out, err := exec.Command(binary, "stats", path).CombinedOutput()
	if err != nil {
		t.Fatalf("stats command failed: %v\n%s", err, out)
	}
	if len(out) == 0 {
		t.Fatal("expected stats output")
	}
}

func TestCLI_Count(t *testing.T) {
	binary := buildBinary(t)
	path := writeTempFile(t, "test.yaml", `a: 1
---
b: 2
`)
	out, err := exec.Command(binary, "count", path).CombinedOutput()
	if err != nil {
		t.Fatalf("count command failed: %v\n%s", err, out)
	}
	if string(out) != "2 document(s)\n" {
		t.Fatalf("expected '2 document(s)\\n', got %q", string(out))
	}
}

func TestCLI_Sort(t *testing.T) {
	binary := buildBinary(t)
	path := writeTempFile(t, "test.yaml", `zebra: 1
alpha: 2
`)
	out, err := exec.Command(binary, "sort", path).CombinedOutput()
	if err != nil {
		t.Fatalf("sort command failed: %v\n%s", err, out)
	}
	if len(out) == 0 {
		t.Fatal("expected sort output")
	}
}

func TestCLI_Print(t *testing.T) {
	binary := buildBinary(t)
	path := writeTempFile(t, "test.yaml", `name: test
`)
	out, err := exec.Command(binary, "print", path).CombinedOutput()
	if err != nil {
		t.Fatalf("print command failed: %v\n%s", err, out)
	}
	if len(out) == 0 {
		t.Fatal("expected print output")
	}
}

func TestCLI_Keys(t *testing.T) {
	binary := buildBinary(t)
	path := writeTempFile(t, "test.yaml", `name: test
version: 1.0
`)
	out, err := exec.Command(binary, "keys", path).CombinedOutput()
	if err != nil {
		t.Fatalf("keys command failed: %v\n%s", err, out)
	}
	if len(out) == 0 {
		t.Fatal("expected keys output")
	}
}

func TestCLI_Schema(t *testing.T) {
	binary := buildBinary(t)
	path := writeTempFile(t, "test.yaml", `name: test
version: 1.0
`)
	out, err := exec.Command(binary, "schema", path).CombinedOutput()
	if err != nil {
		t.Fatalf("schema command failed: %v\n%s", err, out)
	}
	if len(out) == 0 {
		t.Fatal("expected schema output")
	}
}

func TestCLI_Merge(t *testing.T) {
	binary := buildBinary(t)
	dir := t.TempDir()
	path1 := filepath.Join(dir, "a.yaml")
	path2 := filepath.Join(dir, "b.yaml")
	os.WriteFile(path1, []byte("name: original"), 0644)
	os.WriteFile(path2, []byte("version: 1.0"), 0644)
	out, err := exec.Command(binary, "merge", path1, path2).CombinedOutput()
	if err != nil {
		t.Fatalf("merge command failed: %v\n%s", err, out)
	}
	if len(out) == 0 {
		t.Fatal("expected merge output")
	}
}

func TestCLI_Flatten(t *testing.T) {
	binary := buildBinary(t)
	path := writeTempFile(t, "test.yaml", `server:
  host: localhost
  port: 8080
`)
	out, err := exec.Command(binary, "flatten", path).CombinedOutput()
	if err != nil {
		t.Fatalf("flatten command failed: %v\n%s", err, out)
	}
	if len(out) == 0 {
		t.Fatal("expected flatten output")
	}
}
