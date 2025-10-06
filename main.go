package main

import (
	"flag"
	"fmt"
	"os"
	"txtreader/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	fileFlag := flag.String("file", "", "Text file to open")
	flag.Parse()

	m, err := ui.InitialModel(*fileFlag)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Printf("fatal: %v\n", err)
		os.Exit(1)
	}
}
