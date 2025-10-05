// ./internal/progress/progress.go
package progress

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"txtreader/internal/model"
	"txtreader/internal/utils"
)

func Save(filePath string, line int, vocabulary, notes []string, readingSeconds float64, readWords int) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting home directory: %v", err)
	}
	progressDir := filepath.Join(homeDir, "ltbr")
	progressPath := filepath.Join(progressDir, "progress.json")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(progressDir, 0755); err != nil {
		return fmt.Errorf("error creating progress directory: %v", err)
	}

	// Read existing progress or create new
	var progress model.ProgressMap
	data, err := os.ReadFile(progressPath)
	if err == nil {
		if err := json.Unmarshal(data, &progress); err != nil {
			return fmt.Errorf("error parsing progress file: %v", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("error reading progress file: %v", err)
	}
	if progress == nil {
		progress = make(model.ProgressMap)
	}

	// Update progress entry
	hash := utils.HashPath(filePath)
	progress[hash] = model.ProgressEntry{
		FileName:       filePath,
		Line:           line,
		Vocabulary:     vocabulary,
		Notes:          notes,
		ReadingSeconds: readingSeconds,
		ReadWords:      readWords,
	}

	// Write back to file
	data, err = json.MarshalIndent(progress, "", "    ")
	if err != nil {
		return fmt.Errorf("error marshaling progress data: %v", err)
	}
	if err := os.WriteFile(progressPath, data, 0644); err != nil {
		return fmt.Errorf("error writing progress file: %v", err)
	}
	return nil
}

func Load(filePath string) (int, []string, []string, float64, int, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return 0, nil, nil, 0, 0, fmt.Errorf("error getting home directory: %v", err)
	}
	progressPath := filepath.Join(homeDir, "ltbr", "progress.json")

	data, err := os.ReadFile(progressPath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil, nil, 0, 0, nil // No progress file exists, return default values
		}
		return 0, nil, nil, 0, 0, fmt.Errorf("error reading progress file: %v", err)
	}

	var textProgress model.ProgressMap
	if err := json.Unmarshal(data, &textProgress); err != nil {
		return 0, nil, nil, 0, 0, fmt.Errorf("error parsing progress file: %v", err)
	}

	hash := utils.HashPath(filePath)
	if entry, exists := textProgress[hash]; exists {
		return entry.Line, entry.Vocabulary, entry.Notes, entry.ReadingSeconds, entry.ReadWords, nil
	}
	return 0, nil, nil, 0, 0, nil // No entry for this file
}
