package core

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestInsertScore_SortsByScoreDescending(t *testing.T) {
	entries := []ScoreEntry{
		{Name: "A", Score: 100, When: "2025-01-01 10:00"},
		{Name: "B", Score: 300, When: "2025-01-01 11:00"},
		{Name: "C", Score: 200, When: "2025-01-01 12:00"},
	}

	result := InsertScore(entries[:0], entries[0])
	result = InsertScore(result, entries[1])
	result = InsertScore(result, entries[2])

	if len(result) != 3 {
		t.Errorf("expected 3 entries, got %d", len(result))
	}
	if result[0].Score != 300 {
		t.Errorf("first entry should have score 300, got %d", result[0].Score)
	}
	if result[1].Score != 200 {
		t.Errorf("second entry should have score 200, got %d", result[1].Score)
	}
	if result[2].Score != 100 {
		t.Errorf("third entry should have score 100, got %d", result[2].Score)
	}
}

func TestInsertScore_SortsByTimeDescendingForTies(t *testing.T) {
	entries := []ScoreEntry{
		{Name: "A", Score: 100, When: "2025-01-01 10:00"},
		{Name: "B", Score: 100, When: "2025-01-01 12:00"},
		{Name: "C", Score: 100, When: "2025-01-01 11:00"},
	}

	result := InsertScore(entries[:0], entries[0])
	result = InsertScore(result, entries[1])
	result = InsertScore(result, entries[2])

	if result[0].Name != "B" {
		t.Errorf("first entry should be B (newest), got %s", result[0].Name)
	}
	if result[1].Name != "C" {
		t.Errorf("second entry should be C, got %s", result[1].Name)
	}
	if result[2].Name != "A" {
		t.Errorf("third entry should be A (oldest), got %s", result[2].Name)
	}
}

func TestInsertScore_CapsAt50Entries(t *testing.T) {
	var entries []ScoreEntry
	for i := 0; i < 60; i++ {
		entry := ScoreEntry{
			Name:  "AAA",
			Score: i,
			When:  "2025-01-01 10:00",
		}
		entries = InsertScore(entries, entry)
	}

	if len(entries) != 50 {
		t.Errorf("expected 50 entries, got %d", len(entries))
	}
	if entries[0].Score != 59 {
		t.Errorf("first entry should have highest score (59), got %d", entries[0].Score)
	}
}

func TestInsertScore_HandlesEmptySlice(t *testing.T) {
	entry := ScoreEntry{Name: "A", Score: 100, When: "2025-01-01 10:00"}
	result := InsertScore([]ScoreEntry{}, entry)

	if len(result) != 1 {
		t.Errorf("expected 1 entry, got %d", len(result))
	}
	if result[0].Name != "A" {
		t.Errorf("expected entry A, got %s", result[0].Name)
	}
}

func TestFormatTime_ReturnsExpectedFormat(t *testing.T) {
	testTime := time.Date(2025, 1, 15, 14, 30, 0, 0, time.Local)
	result := FormatTime(testTime)

	expected := testTime.Local().Format("2006-01-02 15:04")
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestFormatTime_ZeroPadding(t *testing.T) {
	testTime := time.Date(2025, 1, 2, 9, 5, 0, 0, time.Local)
	result := FormatTime(testTime)

	if len(result) != 16 {
		t.Errorf("expected format length 16, got %d", len(result))
	}
	if result[0:4] != "2025" {
		t.Errorf("expected year 2025, got %s", result[0:4])
	}
}

func TestSaveConfig_ThenLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")

	saveCfg := Config{Ghost: true, StartLevel: 5}
	err := saveConfigToPath(saveCfg, configFile)
	if err != nil {
		t.Fatalf("saveConfigToPath failed: %v", err)
	}

	loaded, err := loadConfigFromPath(configFile)
	if err != nil {
		t.Fatalf("loadConfigFromPath failed: %v", err)
	}
	if loaded.Ghost != saveCfg.Ghost {
		t.Errorf("Ghost: expected %v, got %v", saveCfg.Ghost, loaded.Ghost)
	}
	if loaded.StartLevel != saveCfg.StartLevel {
		t.Errorf("StartLevel: expected %d, got %d", saveCfg.StartLevel, loaded.StartLevel)
	}
}

func TestLoadConfig_MissingFile_ReturnsDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "nonexistent.json")

	loaded, err := loadConfigFromPath(configFile)
	if err != nil {
		t.Fatalf("loadConfigFromPath failed: %v", err)
	}
	if loaded.Ghost != false {
		t.Errorf("expected Ghost=false, got %v", loaded.Ghost)
	}
	if loaded.StartLevel != 0 {
		t.Errorf("expected StartLevel=0, got %d", loaded.StartLevel)
	}
}

func TestLoadConfig_InvalidJSON_ReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(configFile, []byte("invalid json"), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err := loadConfigFromPath(configFile)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestLoadConfig_ClampsStartLevel(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		json      string
		wantLevel int
	}{
		{"negative clamps to 0", `{"ghost":false,"startLevel":-1}`, 0},
		{"over 9 clamps to 9", `{"ghost":true,"startLevel":15}`, 9},
		{"valid 0 stays 0", `{"ghost":false,"startLevel":0}`, 0},
		{"valid 9 stays 9", `{"ghost":true,"startLevel":9}`, 9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile := filepath.Join(tmpDir, tt.json+".json")
			if err := os.WriteFile(configFile, []byte(tt.json), 0o644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}
			loaded, err := loadConfigFromPath(configFile)
			if err != nil {
				t.Fatalf("loadConfigFromPath failed: %v", err)
			}
			if loaded.StartLevel != tt.wantLevel {
				t.Errorf("StartLevel: expected %d, got %d", tt.wantLevel, loaded.StartLevel)
			}
		})
	}
}

func TestSaveScores_ThenLoadScores(t *testing.T) {
	tmpDir := t.TempDir()
	scoresFile := filepath.Join(tmpDir, "scores.json")

	scores := []ScoreEntry{
		{Name: "AAA", Score: 100, Lines: 10, Level: 1, When: "2025-01-01 10:00"},
		{Name: "BBB", Score: 200, Lines: 20, Level: 2, When: "2025-01-02 11:00"},
	}
	err := saveScoresToPath(scores, scoresFile)
	if err != nil {
		t.Fatalf("saveScoresToPath failed: %v", err)
	}

	loaded, err := loadScoresFromPath(scoresFile)
	if err != nil {
		t.Fatalf("loadScoresFromPath failed: %v", err)
	}
	if len(loaded) != 2 {
		t.Errorf("expected 2 entries, got %d", len(loaded))
	}
	if loaded[0].Score != 100 || loaded[1].Score != 200 {
		t.Errorf("scores order should be preserved, got %d, %d", loaded[0].Score, loaded[1].Score)
	}
}

func TestLoadScores_MissingFile_ReturnsEmptySlice(t *testing.T) {
	tmpDir := t.TempDir()
	scoresFile := filepath.Join(tmpDir, "nonexistent.json")

	loaded, err := loadScoresFromPath(scoresFile)
	if err != nil {
		t.Fatalf("loadScoresFromPath failed: %v", err)
	}
	if len(loaded) != 0 {
		t.Errorf("expected 0 entries, got %d", len(loaded))
	}
}

func TestLoadScores_InvalidJSON_ReturnsEmptySliceWithError(t *testing.T) {
	tmpDir := t.TempDir()
	scoresFile := filepath.Join(tmpDir, "scores.json")
	if err := os.WriteFile(scoresFile, []byte("invalid json"), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	loaded, err := loadScoresFromPath(scoresFile)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
	if len(loaded) != 0 {
		t.Errorf("expected 0 entries on error, got %d", len(loaded))
	}
}

func TestSaveConfigToPath_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "subdir", "config.json")

	cfg := Config{Ghost: true, StartLevel: 3}
	err := saveConfigToPath(cfg, configFile)
	if err != nil {
		t.Fatalf("saveConfigToPath failed: %v", err)
	}

	loaded, err := loadConfigFromPath(configFile)
	if err != nil {
		t.Fatalf("loadConfigFromPath failed: %v", err)
	}
	if loaded.Ghost != cfg.Ghost || loaded.StartLevel != cfg.StartLevel {
		t.Errorf("loaded config doesn't match saved")
	}
}

func TestSaveScoresToPath_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	scoresFile := filepath.Join(tmpDir, "subdir", "scores.json")

	scores := []ScoreEntry{{Name: "AAA", Score: 500, Lines: 50, Level: 5, When: "2025-01-01 10:00"}}
	err := saveScoresToPath(scores, scoresFile)
	if err != nil {
		t.Fatalf("saveScoresToPath failed: %v", err)
	}

	loaded, err := loadScoresFromPath(scoresFile)
	if err != nil {
		t.Fatalf("loadScoresFromPath failed: %v", err)
	}
	if len(loaded) != 1 || loaded[0].Score != 500 {
		t.Errorf("loaded scores don't match saved")
	}
}

func TestInsertScore_ScoreTieBreakerByTime(t *testing.T) {
	entries := []ScoreEntry{
		{Name: "Old", Score: 100, When: "2025-01-01 10:00"},
		{Name: "New", Score: 100, When: "2025-01-02 10:00"},
	}

	result := InsertScore([]ScoreEntry{}, entries[0])
	result = InsertScore(result, entries[1])

	if result[0].Name != "New" {
		t.Errorf("expected New (newer) first, got %s", result[0].Name)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()
	if cfg.Ghost != false {
		t.Errorf("expected Ghost=false, got %v", cfg.Ghost)
	}
	if cfg.StartLevel != 0 {
		t.Errorf("expected StartLevel=0, got %d", cfg.StartLevel)
	}
}
