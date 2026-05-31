package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type Config struct {
	Ghost      bool `json:"ghost"`
	StartLevel int  `json:"startLevel"`
}

type ScoreEntry struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
	Lines int    `json:"lines"`
	Level int    `json:"level"`
	When  string `json:"when"`
}

func defaultConfig() Config {
	return Config{
		Ghost:      false,
		StartLevel: 0,
	}
}

func LoadConfig() (Config, error) {
	config := defaultConfig()
	path, err := configPath()
	if err != nil {
		return config, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return config, nil
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return config, err
	}
	if config.StartLevel < 0 {
		config.StartLevel = 0
	}
	if config.StartLevel > 9 {
		config.StartLevel = 9
	}
	return config, nil
}

func SaveConfig(cfg Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func LoadScores() ([]ScoreEntry, error) {
	path, err := scoresPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return []ScoreEntry{}, nil
	}
	var scores []ScoreEntry
	if err := json.Unmarshal(data, &scores); err != nil {
		return []ScoreEntry{}, err
	}
	return scores, nil
}

func SaveScores(scores []ScoreEntry) error {
	path, err := scoresPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(scores, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func InsertScore(scores []ScoreEntry, entry ScoreEntry) []ScoreEntry {
	scores = append(scores, entry)
	sort.Slice(scores, func(i, j int) bool {
		if scores[i].Score == scores[j].Score {
			return scores[i].When > scores[j].When
		}
		return scores[i].Score > scores[j].Score
	})
	if len(scores) > 50 {
		return scores[:50]
	}
	return scores
}

func configPath() (string, error) {
	root, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(root, "tmvgs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func scoresPath() (string, error) {
	root, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(root, "tmvgs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "scores.json"), nil
}

func FormatTime(t time.Time) string {
	return t.Local().Format("2006-01-02 15:04")
}