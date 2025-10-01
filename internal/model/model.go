package model

type ProgressEntry struct {
	FileName   string   `json:"file_name"`
	Line       int      `json:"line"`
	Vocabulary []string `json:"vocabulary"`
	Notes      []string `json:"notes"`
}

type ProgressMap map[string]ProgressEntry
