package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	lines          []string
	currentLine    int
	currentWordIdx int
	selectedWord   string
	currentTab     int
	tabs           []string
	width, height  int
	filePath       string
	showDialog     bool
	lineInput      string
}

func initialModel() model {
	m := model{
		tabs:           []string{"Texto", "Vocabulario", "Referencias"},
		currentTab:     0,
		currentLine:    0,
		currentWordIdx: 0,
		selectedWord:   "",
		showDialog:     false,
		lineInput:      "",
	}

	if len(os.Args) > 1 {
		file, err := os.Open(os.Args[1])
		if err != nil {
			fmt.Printf("Error opening file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			m.lines = append(m.lines, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Sample lines for testing
		sampleLines := []string{
			"Line 1: This is the first line with some words to navigate.",
			"Line 2: Another line containing various words like hello world test.",
			"Line 3: Third line for demonstration purposes only.",
			"Line 4: More content to simulate a longer file.",
			"Line 5: Fifth line with additional words.",
			// Add more lines to simulate hundreds...
		}
		for i := 6; i <= 200; i++ {
			sampleLines = append(sampleLines, fmt.Sprintf("Line %d: This is a sample line.", i))
		}
		m.lines = sampleLines
	}

	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.showDialog {
			switch msg.String() {
			case "esc":
				m.showDialog = false
				m.lineInput = ""
			case "enter":
				if m.lineInput != "" {
					if lineNum, err := strconv.Atoi(m.lineInput); err == nil && lineNum > 0 && lineNum <= len(m.lines) {
						m.currentLine = lineNum - 1 // Convert to 0-based index
						m.currentWordIdx = 0        // Reset word index
					}
				}
				m.showDialog = false
				m.lineInput = ""
			case "backspace":
				if len(m.lineInput) > 0 {
					m.lineInput = m.lineInput[:len(m.lineInput)-1]
				}
			default:
				if len(msg.String()) == 1 && msg.String() >= "0" && msg.String() <= "9" {
					m.lineInput += msg.String()
				}
			}
			return m, nil
		}
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "1":
			m.currentTab = 0
		case "2":
			m.currentTab = 1
		case "3":
			m.currentTab = 2
		case "g":
			if m.currentTab == 0 {
				m.showDialog = true
				m.lineInput = ""
			}
		default:
			if m.currentTab != 0 {
				break
			}
			switch msg.String() {
			case "j":
				if m.currentLine < len(m.lines)-1 {
					m.currentLine++
					m.currentWordIdx = 0 // Reset word index on line change
				}
			case "k":
				if m.currentLine > 0 {
					m.currentLine--
					m.currentWordIdx = 0 // Reset word index on line change
				}
			case "left":
				words := strings.Fields(m.lines[m.currentLine])
				if len(words) > 0 {
					m.currentWordIdx = (m.currentWordIdx - 1 + len(words)) % len(words)
				}
			case "right":
				words := strings.Fields(m.lines[m.currentLine])
				if len(words) > 0 {
					m.currentWordIdx = (m.currentWordIdx + 1) % len(words)
				}
			case "w":
				words := strings.Fields(m.lines[m.currentLine])
				if len(words) > 0 && m.currentWordIdx < len(words) {
					m.selectedWord = words[m.currentWordIdx]
				}
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m model) View() string {
	var b strings.Builder

	// Tab bar
	var tabViews []string
	for i, tab := range m.tabs {
		if i == m.currentTab {
			tabViews = append(tabViews, lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15")). // Bright white
				Background(lipgloss.Color("27")). // Blue background
				Border(lipgloss.RoundedBorder(), true).
				BorderForeground(lipgloss.Color("51")). // Cyan border
				Padding(0, 2).
				Margin(0, 1, 0, 0).
				Render(tab))
		} else {
			tabViews = append(tabViews, lipgloss.NewStyle().
				Italic(true).
				Foreground(lipgloss.Color("250")). // Light gray
				Background(lipgloss.Color("237")). // Dark gray background
				Border(lipgloss.NormalBorder(), true).
				BorderForeground(lipgloss.Color("244")). // Medium gray border
				Padding(0, 2).
				Margin(0, 1, 0, 0).
				Render(tab))
		}
	}
	tabBar := lipgloss.JoinHorizontal(lipgloss.Left, tabViews...)
	tabBar = lipgloss.NewStyle().
		Padding(0, 1).
		Height(3). // Increased height to accommodate borders
		Border(lipgloss.NormalBorder(), false, false, true, false).
		Foreground(lipgloss.Color("244")). // Medium gray for border
		Render(tabBar)
	b.WriteString(tabBar + "\n")

	// Content area
	contentHeight := m.height - 4 // Adjusted for increased tab bar height (3) + status (1)
	if contentHeight < 1 {
		contentHeight = 1
	}

	if m.currentTab == 0 {
		// Texto tab: show lines around current
		viewStart := max(0, m.currentLine-contentHeight/2)
		viewEnd := min(len(m.lines), viewStart+contentHeight)
		for i := viewStart; i < viewEnd; i++ {
			if i == m.currentLine {
				// Highlight current line and word
				line := m.lines[i]
				words := strings.Fields(line)
				var highlightedWords []string
				for j, word := range words {
					if j == m.currentWordIdx && len(words) > 0 {
						highlightedWords = append(highlightedWords, lipgloss.NewStyle().
							Bold(true).
							Background(lipgloss.Color("226")). // Bright yellow
							Foreground(lipgloss.Color("232")). // Dark gray
							Padding(0, 1).
							Render(word))
					} else {
						highlightedWords = append(highlightedWords, word)
					}
				}
				hlLine := strings.Join(highlightedWords, " ")
				// Apply line highlight
				hlLine = lipgloss.NewStyle().
					Background(lipgloss.Color("236")). // Darker gray background
					Foreground(lipgloss.Color("15")). // Bright white text
					Padding(0, 1).
					Render(hlLine)
				b.WriteString(hlLine + "\n")
			} else {
				b.WriteString(lipgloss.NewStyle().
					Foreground(lipgloss.Color("252")). // Light gray for non-current lines
					Render(m.lines[i]) + "\n")
			}
		}
	} else if m.currentTab == 1 {
		// Vocabulario: placeholder
		placeholder := "Contenido de Vocabulario (por determinarse).\n"
		b.WriteString(strings.Repeat(placeholder, contentHeight))
	} else if m.currentTab == 2 {
		// Referencias: placeholder
		placeholder := "Contenido de Referencias (por determinarse).\n"
		b.WriteString(strings.Repeat(placeholder, contentHeight))
	}

	// Dialog for line input
	if m.showDialog {
		dialogWidth := 30
		dialogHeight := 3
		dialog := lipgloss.NewStyle().
			Width(dialogWidth).
			Height(dialogHeight).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")). // Purple border
			Padding(1, 2).
			Align(lipgloss.Center).
			Render(fmt.Sprintf("Go to line: %s", m.lineInput))
		dialog = lipgloss.Place(
			m.width,
			m.height-2, // Account for status bar
			lipgloss.Center,
			lipgloss.Center,
			dialog,
			lipgloss.WithWhitespaceBackground(lipgloss.Color("235")),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("63")),
		)
		b.WriteString("\n" + dialog)
	}

	// Status bar
	total := len(m.lines)
	percent := float64(0)
	if total > 0 {
		percent = float64(m.currentLine) / float64(total-1) * 100
	}
	lineInfo := fmt.Sprintf("Línea: %d/%d (%.0f%%)", m.currentLine+1, total, percent)
	selInfo := ""
	if m.selectedWord != "" {
		selInfo = fmt.Sprintf(" | Seleccionada: %s", m.selectedWord)
	}
	status := lineInfo + selInfo
	statusStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Height(1).
		Bold(true).
		Foreground(lipgloss.Color("15")). // Bright white
		Background(lipgloss.Color("234")). // Very dark gray
		Width(m.width).
		Align(lipgloss.Left)
	b.WriteString(statusStyle.Render(status))

	return lipgloss.NewStyle().Width(m.width).Height(m.height).Render(b.String())
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("fatal: %v\n", err)
		os.Exit(1)
	}
}
