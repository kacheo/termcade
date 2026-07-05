package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckGoldenUpdate(t *testing.T) {
	dir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig) //nolint:errcheck

	*update = true
	defer func() { *update = false }()

	CheckGolden(t, "sample", "golden content")

	got, err := os.ReadFile(filepath.Join("testdata", "sample.golden"))
	if err != nil {
		t.Fatalf("golden file not created: %v", err)
	}
	if string(got) != "golden content" {
		t.Errorf("golden file = %q, want \"golden content\"", string(got))
	}
}

func TestCheckGoldenMatch(t *testing.T) {
	dir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig) //nolint:errcheck

	if err := os.MkdirAll("testdata", 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join("testdata", "match.golden"), []byte("expected"), 0o600); err != nil {
		t.Fatal(err)
	}

	*update = false
	CheckGolden(t, "match", "expected")
}

func TestCheckGoldenMismatch(t *testing.T) {
	dir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig) //nolint:errcheck

	if err := os.MkdirAll("testdata", 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join("testdata", "mismatch.golden"), []byte("expected"), 0o600); err != nil {
		t.Fatal(err)
	}

	*update = false
	inner := &testing.T{}
	CheckGolden(inner, "mismatch", "different content")
	if !inner.Failed() {
		t.Error("expected CheckGolden to fail on content mismatch")
	}
}
