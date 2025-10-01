package ui

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"txtreader/internal/progress"
	"txtreader/internal/text"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type UiModel struct {
	lines                 []string
	currentLine           int
	currentWordIdx        int
	selectedWord          string
	copiedToClipboardWord string
	currentTab            int
	tabs                  []string
	width, height         int
	filePath              string
	showDialog            bool
	lineInput             string
	tabWidths             []int // Store rendered width of each tab
	vocabulary            []string
	currentVocabIdx       int // Track selected vocabulary word
	notes                 []string
	currentNoteIdx        int // Track selected note
	showNoteDialog        bool
	noteInput             []string // Multiline note input
	showLinksDialog       bool     // Track links widget visibility
	currentLinkIdx        int      // Track selected link
}

func InitialModel(filePath string) UiModel {
	m := UiModel{
		tabs:            []string{"Texto", "Vocabulario", "Notas"},
		currentTab:      0,
		currentLine:     0,
		currentWordIdx:  0,
		currentVocabIdx: 0,
		currentNoteIdx:  0,
		currentLinkIdx:  0,
		selectedWord:    "",
		showDialog:      false,
		lineInput:       "",
		tabWidths:       make([]int, 3), // Initialize for 3 tabs
		vocabulary:      []string{},
		notes:           []string{},
		showNoteDialog:  false,
		noteInput:       []string{""},
		showLinksDialog: false,
	}

	m.filePath = filePath
	file, err := os.Open(m.filePath)
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
	// Load progress for the file
	line, vocab, notes, err := progress.Load(m.filePath)
	if err != nil {
		fmt.Printf("Error loading progress: %v\n", err)
		os.Exit(1)
	}
	if line > 0 && line < len(m.lines) {
		m.currentLine = line
	}
	if vocab == nil {
		vocab = []string{}
	}
	if notes == nil {
		notes = []string{}
	}
	m.vocabulary = vocab
	m.notes = notes

	return m
}

func (m UiModel) Init() tea.Cmd {
	return nil
}

func (m UiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		if m.showNoteDialog {
			switch msg.String() {
			case "esc":
				m.showNoteDialog = false
				m.noteInput = []string{""} // Reset note input
			case "enter":
				// Add new line to note input
				m.noteInput = append(m.noteInput, "")
			case "backspace":
				currentLine := len(m.noteInput) - 1
				if len(m.noteInput[currentLine]) > 0 {
					m.noteInput[currentLine] = m.noteInput[currentLine][:len(m.noteInput[currentLine])-1]
				} else if currentLine > 0 {
					m.noteInput = m.noteInput[:currentLine]
				}
			case "ctrl+s":
				// Save note
				note := strings.Join(m.noteInput, "\n")
				if note != "" {
					m.notes = append(m.notes, note)
				}
				m.showNoteDialog = false
				m.noteInput = []string{""} // Reset note input
			case "ctrl+c":
				// Cancel note
				m.showNoteDialog = false
				m.noteInput = []string{""} // Reset note input
			default:
				if len(msg.String()) == 1 {
					currentLine := len(m.noteInput) - 1
					m.noteInput[currentLine] += msg.String()
				}
			}
			return m, nil
		}
		if m.showLinksDialog {
			switch msg.String() {
			case "esc":
				m.showLinksDialog = false
				m.currentLinkIdx = 0
			case "j":
				if m.currentLinkIdx < 1 { // Only two items: GoodReads, RAE
					m.currentLinkIdx++
				}
			case "k":
				if m.currentLinkIdx > 0 {
					m.currentLinkIdx--
				}
			case "enter":
				// Open the selected link in the default browser
				links := []string{"https://www.goodreads.com/", "https://www.rae.es/"}
				if m.currentLinkIdx >= 0 && m.currentLinkIdx < len(links) {
					url := links[m.currentLinkIdx]
					var cmd *exec.Cmd
					switch runtime.GOOS {
					case "windows":
						cmd = exec.Command("cmd", "/c", "start", url)
					case "darwin":
						cmd = exec.Command("open", url)
					default: // linux, bsd, etc.
						cmd = exec.Command("xdg-open", url)
					}
					if err := cmd.Start(); err != nil {
						fmt.Printf("Error opening browser: %v\n", err)
					}
				}
				m.showLinksDialog = false
				m.currentLinkIdx = 0
			case "ctrl+c":
				// Cancel the dialog
				m.showLinksDialog = false
				m.currentLinkIdx = 0
			}
			return m, nil
		}
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "s":
			if m.filePath != "" {
				if err := progress.Save(m.filePath, m.currentLine, m.vocabulary, m.notes); err != nil {
					fmt.Printf("Error saving progress: %v\n", err)
				}
			}
		case "1":
			m.currentTab = 0
		case "2":
			m.currentTab = 1
			m.currentVocabIdx = 0 // Reset vocabulary index when switching to Vocabulario tab
		case "3":
			m.currentTab = 2
			m.currentNoteIdx = 0 // Reset note index when switching to Notas tab
		case "n":
			m.showNoteDialog = true
			m.noteInput = []string{""} // Initialize note input
		case "g":
			if m.currentTab == 0 {
				m.showDialog = true
				m.lineInput = ""
			}
		case "o":
			m.showLinksDialog = true
			m.currentLinkIdx = 0
		default:
			if m.currentTab == 0 {
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
						// Add to vocabulary if not already present:
						wordAlreadyInVocab := text.Contains(&m.vocabulary, m.selectedWord)
						if !wordAlreadyInVocab {
							m.vocabulary = append(m.vocabulary, m.selectedWord)
						}
					}
				case "c":
					words := strings.Fields(m.lines[m.currentLine])
					if len(words) > 0 && m.currentWordIdx < len(words) {
						m.copiedToClipboardWord = words[m.currentWordIdx]
						err := clipboard.WriteAll(m.copiedToClipboardWord)
						if err != nil {
							fmt.Printf("Error copying word: %v\n", err)
						}
					}
				}
			} else if m.currentTab == 1 {
				switch msg.String() {
				case "j":
					if m.currentVocabIdx < len(m.vocabulary)-1 {
						m.currentVocabIdx++
					}
				case "k":
					if m.currentVocabIdx > 0 {
						m.currentVocabIdx--
					}
				case "d":
					// Delete current vocabulary word
					if len(m.vocabulary) > 0 && m.currentVocabIdx < len(m.vocabulary) {
						// Remove the word at currentVocabIdx
						m.vocabulary = append(m.vocabulary[:m.currentVocabIdx], m.vocabulary[m.currentVocabIdx+1:]...)
						// Adjust currentVocabIdx if necessary
						if m.currentVocabIdx >= len(m.vocabulary) && len(m.vocabulary) > 0 {
							m.currentVocabIdx = len(m.vocabulary) - 1
						}
						if len(m.vocabulary) == 0 {
							m.currentVocabIdx = 0
						}
					}
				}
			} else if m.currentTab == 2 {
				switch msg.String() {
				case "j":
					if m.currentNoteIdx < len(m.notes)-1 {
						m.currentNoteIdx++
					}
				case "k":
					if m.currentNoteIdx > 0 {
						m.currentNoteIdx--
					}
				}
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m UiModel) View() string {
	if m.showDialog {
		return m.renderWithDialog(m.renderGoToLineDialog())
	}
	if m.showNoteDialog {
		return m.renderWithDialog(m.renderNoteDialog())
	}
	if m.showLinksDialog {
		return m.renderWithDialog(m.renderLinksDialog())
	}

	return m.renderMainContent()
}

func (m UiModel) renderMainContent() string {
	var content strings.Builder

	// Tab bar
	var tabViews []string
	m.tabWidths = make([]int, len(m.tabs)) // Reset tab widths
	for i, tab := range m.tabs {
		style := lipgloss.NewStyle().
			Padding(0, 2).
			Margin(0, 1, 0, 0)
		if i == m.currentTab {
			style = style.
				Bold(true).
				Foreground(lipgloss.Color("15")). // Bright white
				Background(lipgloss.Color("27")). // Blue background
				Border(lipgloss.RoundedBorder(), true).
				BorderForeground(lipgloss.Color("51")) // Cyan border
		} else {
			style = style.
				Italic(true).
				Foreground(lipgloss.Color("250")). // Light gray
				Background(lipgloss.Color("237")). // Dark gray background
				Border(lipgloss.NormalBorder(), true).
				BorderForeground(lipgloss.Color("244")) // Medium gray border
		}
		renderedTab := style.Render(tab)
		tabViews = append(tabViews, renderedTab)
		m.tabWidths[i] = lipgloss.Width(renderedTab) // Store width of rendered tab
	}
	tabBar := lipgloss.JoinHorizontal(lipgloss.Left, tabViews...)
	tabBar = lipgloss.NewStyle().
		Padding(0, 1).
		Height(3). // Increased height to accommodate borders
		Border(lipgloss.NormalBorder(), false, false, true, false).
		Foreground(lipgloss.Color("244")). // Medium gray for border
		Render(tabBar)
	content.WriteString(tabBar + "\n")

	// Content area
	contentHeight := m.height - 4 // Adjusted for tab bar height (3) + status (1)
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
				content.WriteString(hlLine + "\n")
			} else {
				content.WriteString(lipgloss.NewStyle().
					Foreground(lipgloss.Color("252")). // Light gray for non-current lines
					Render(m.lines[i]) + "\n")
			}
		}
	} else if m.currentTab == 1 {
		// Vocabulario tab: show vocabulary words with navigation
		if len(m.vocabulary) == 0 {
			content.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("252")).
				Align(lipgloss.Right).
				Render("No Vocab\n"))
		} else {
			viewStart := max(0, m.currentVocabIdx-contentHeight/2)
			viewEnd := min(len(m.vocabulary), viewStart+contentHeight)
			for i := viewStart; i < viewEnd; i++ {
				word := m.vocabulary[i]
				if i == m.currentVocabIdx {
					content.WriteString(lipgloss.NewStyle().
						Background(lipgloss.Color("236")). // Darker gray background
						Foreground(lipgloss.Color("15")). // Bright white text
						Padding(0, 1).
						Render(word) + "\n")
				} else {
					content.WriteString(lipgloss.NewStyle().
						Foreground(lipgloss.Color("252")). // Light gray
						Render(word) + "\n")
				}
			}
		}
	} else if m.currentTab == 2 {
		// Notas tab: show notes with navigation and borders
		if len(m.notes) == 0 {
			content.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("252")).
				Render("No Notes\n"))
		} else {
			viewStart := max(0, m.currentNoteIdx-contentHeight/2)
			viewEnd := min(len(m.notes), viewStart+contentHeight)
			renderedNotes := []string{}
			usedHeight := 0
			for i := viewStart; i < viewEnd && usedHeight < contentHeight; i++ {
				// Split note into lines
				lines := strings.Split(m.notes[i], "\n")
				// Calculate note height (including padding and borders)
				noteHeight := len(lines)
				if noteHeight+2 > contentHeight-usedHeight {
					// Limit lines to fit remaining height (2 for borders)
					noteHeight = contentHeight - usedHeight - 2
					if noteHeight < 1 {
						break // No room for another note
					}
					lines = lines[:noteHeight]
				}
				noteText := strings.Join(lines, "\n")
				// Determine max width for the note
				maxWidth := 0
				for _, line := range lines {
					lineWidth := lipgloss.Width(line)
					if lineWidth > maxWidth {
						maxWidth = lineWidth
					}
				}
				if maxWidth > m.width-6 { // Account for padding (2) and borders (4)
					maxWidth = m.width - 6
				}
				// Apply styling with border
				noteStyle := lipgloss.NewStyle().
					Width(maxWidth).
					Border(lipgloss.NormalBorder(), true).
					BorderForeground(lipgloss.Color("244")). // Medium gray border
					Padding(0, 1)
				if i == m.currentNoteIdx {
					noteStyle = noteStyle.
						Background(lipgloss.Color("236")). // Darker gray background
						Foreground(lipgloss.Color("15")) // Bright white text
				} else {
					noteStyle = noteStyle.
						Foreground(lipgloss.Color("252")) // Light gray
				}
				renderedNote := noteStyle.Render(noteText)
				renderedNotes = append(renderedNotes, renderedNote)
				usedHeight += noteHeight + 2 // Add border height
			}
			// Join notes vertically with a newline separator
			content.WriteString(lipgloss.JoinVertical(lipgloss.Left, renderedNotes...) + "\n")
		}
	}

	// Status bar
	total := len(m.lines)
	percent := float64(0)
	if total > 0 {
		percent = float64(m.currentLine) / float64(total-1) * 100
	}
	lineInfo := fmt.Sprintf("Línea: %d/%d (%.4f%%)", m.currentLine+1, total, percent)
	selInfo := ""
	if m.currentTab == 0 && m.selectedWord != "" {
		selInfo = fmt.Sprintf(" | Seleccionada: %s", m.selectedWord)
	} else if m.currentTab == 0 && m.copiedToClipboardWord != "" {
		selInfo += fmt.Sprintf(" | Copiada al portapapeles: %s", m.copiedToClipboardWord)
	} else if m.currentTab == 1 && len(m.vocabulary) > 0 {
		selInfo = fmt.Sprintf(" | Palabra: %s", m.vocabulary[m.currentVocabIdx])
	} else if m.currentTab == 2 && len(m.notes) > 0 {
		selInfo = fmt.Sprintf(" | Nota: %d/%d", m.currentNoteIdx+1, len(m.notes))
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
	content.WriteString(statusStyle.Render(status))

	return content.String()
}

func (m UiModel) renderGoToLineDialog() string {
	dialogContent := fmt.Sprintf("Go to line: %s", m.lineInput)

	dialog := lipgloss.NewStyle().
		Width(30).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2).
		Align(lipgloss.Center).
		Background(lipgloss.Color("235")).
		Render(dialogContent)

	return dialog
}

func (m UiModel) renderNoteDialog() string {
	dialogWidth := min(m.width*3/4, m.width-4)
	if dialogWidth > 80 {
		dialogWidth = 80
	}

	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Align(lipgloss.Center).
		Padding(0, 0).
		Width(dialogWidth - 4).
		Render("Agregar nota")

	noteText := strings.Join(m.noteInput, "\n")
	inputBox := lipgloss.NewStyle().
		Width(dialogWidth - 4).
		Height(10).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1).
		Render(noteText)

	saveButton := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("28")).
		Padding(0, 2).
		Margin(0, 1).
		Render("Guardar (Ctrl+S)")

	cancelButton := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("160")).
		Padding(0, 2).
		Margin(0, 1).
		Render("Cancelar (Ctrl+C)")

	buttons := lipgloss.JoinHorizontal(lipgloss.Center, saveButton, cancelButton)
	dialogContent := lipgloss.JoinVertical(lipgloss.Left, title, inputBox, buttons)

	dialog := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1).
		Background(lipgloss.Color("235")).
		Render(dialogContent)

	return dialog
}

func (m UiModel) renderLinksDialog() string {
	dialogWidth := 40

	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Align(lipgloss.Center).
		Padding(0, 0).
		Width(dialogWidth - 4).
		Render("Seleccionar Enlace")

	links := []string{"GoodReads", "Real Academia Española"}
	var linkItems []string
	for i, link := range links {
		style := lipgloss.NewStyle().
			Width(dialogWidth-6).
			Padding(0, 1).
			Align(lipgloss.Left)
		if i == m.currentLinkIdx {
			style = style.
				Background(lipgloss.Color("236")).
				Foreground(lipgloss.Color("15"))
		} else {
			style = style.
				Foreground(lipgloss.Color("252"))
		}
		linkItems = append(linkItems, style.Render(link))
	}
	linksList := lipgloss.JoinVertical(lipgloss.Left, linkItems...)

	listBox := lipgloss.NewStyle().
		Width(dialogWidth-4).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 1).
		Render(linksList)

	goButton := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("28")).
		Padding(0, 2).
		Margin(0, 1).
		Render("Ir A (Enter)")

	cancelButton := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("160")).
		Padding(0, 2).
		Margin(0, 1).
		Render("Cancelar (Ctrl+C)")

	buttons := lipgloss.JoinHorizontal(lipgloss.Center, goButton, cancelButton)
	dialogContent := lipgloss.JoinVertical(lipgloss.Left, title, listBox, buttons)

	dialog := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1).
		Background(lipgloss.Color("235")).
		Render(dialogContent)

	return dialog
}

func (m UiModel) renderWithDialog(dialog string) string {
	background := m.renderMainContent()
	dialogCentered := lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		dialog,
	)
	return background + "\n" + dialogCentered
}
