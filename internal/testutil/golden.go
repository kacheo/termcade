package testutil

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "update golden files")

// CheckGolden compares the given output against a golden file.
// If the -update flag is set, it writes the output to the golden file (creating
// testdata/ directory if needed). Otherwise, it reads the golden file and fails
// with a diff if the contents don't match. If the golden file doesn't exist and
// -update is not set, it fails with a clear message.
//
// To regenerate all golden files run: make update-golden
// (equivalent to: go test -run TestGoldenRender -args -update ./games/...)
func CheckGolden(t *testing.T, name, got string) {
	t.Helper()

	goldenPath := filepath.Join("testdata", name+".golden")

	if *update {
		// Create testdata directory if it doesn't exist
		testdataDir := filepath.Dir(goldenPath)
		if err := os.MkdirAll(testdataDir, 0o750); err != nil {
			t.Fatalf("failed to create testdata directory: %v", err)
		}

		// Write the golden file
		if err := os.WriteFile(goldenPath, []byte(got), 0o600); err != nil {
			t.Fatalf("failed to write golden file: %v", err)
		}
		return
	}

	// Read the golden file
	want, err := os.ReadFile(goldenPath) //nolint:gosec // path is constructed from test names, not user input
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("golden file not found: %s\nrun with -update to create", goldenPath)
		}
		t.Fatalf("failed to read golden file: %v", err)
	}

	// Compare
	if got != string(want) {
		t.Errorf("golden file mismatch: %s\ngot:\n%s\nwant:\n%s", goldenPath, got, string(want))
	}
}
