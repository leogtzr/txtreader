package model

type ProgressEntry struct {
	FileName       string   `json:"file_name"`
	Line           int      `json:"line"`
	Vocabulary     []string `json:"vocabulary"`
	Notes          []string `json:"notes"`
	ReadingSeconds float64  `json:"reading_seconds"`
	ReadWords      int      `json:"read_words"`
}

type ProgressMap map[string]ProgressEntry
