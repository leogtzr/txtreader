package ui

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
	"txtreader/internal/progress"
	"txtreader/internal/text"
	"txtreader/internal/text/stats"
	"txtreader/internal/utils"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/viewport"
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
	showGotoLineDialog    bool
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
	showDeleteNoteDialog  bool     // Track delete note confirmation dialog
	deleteNoteConfirmIdx  int      // Track selected option in delete confirmation (0=No, 1=Yes)
	totalLines            int      // Total number of lines
	totalWords            int      // Total number of words
	longestLine           string   // The longest line content
	longestLineLength     int      // Length of the longest line
	longestWord           string
	topWords              []stats.WordCount
	cumulativeWords       []int   // Cumulative words up to each line
	totalReadingSeconds   float64 // Total reading time in seconds (loaded from progress)
	totalReadWords        int     // Total words read (loaded from progress)
	sessionReadingTime    float64 // Session reading time in seconds
	sessionWordsRead      int     // Session words read
	lastActionTime        time.Time
	vp                    viewport.Model // Viewport para manejar scroll en el tab de texto
	showHelpDialog        bool
}

const DefaultWPM = 250.0

const (
	brightWhiteColor  = lipgloss.Color("15")
	blueColor         = lipgloss.Color("27")
	cyanColor         = lipgloss.Color("51")
	lightGrayColor    = lipgloss.Color("250")
	darkGrayColor     = lipgloss.Color("237")
	mediumGrayColor   = lipgloss.Color("244")
	brightYellowColor = lipgloss.Color("226")
	redColor          = lipgloss.Color("160")
	royalBlueColor    = lipgloss.Color("63")
	greenColor        = lipgloss.Color("28")
	greyColor         = lipgloss.Color("235")
	grayColor         = lipgloss.Color("240")
)

const (
	keyQuit            = "q"
	keySave            = "s"
	keyNextLine        = "j"
	keyPrevLine        = "k"
	keyEsc             = "esc"
	keyMainTextTab     = "1"
	keyVocabTab        = "2"
	keyNotesTab        = "3"
	keyStatsTab        = "4"
	keyShowNoteDialog  = "n"
	keyEnter           = "enter"
	keyBackspace       = "backspace"
	keyEspace          = " "
	keyCancel          = "ctrl+c"
	keyDelete          = "d"
	keyOpenLinksDialog = "o"
	keyControlSave     = "ctrl+s"
	keyLeft            = "left"
	keyRight           = "right"
	keyAddToVocabulary = "w"
	keyCopyToClipboard = "c"
	keyGotoLineDialog  = "g"
	keyZero            = "0"
	keyDollarSign      = "$"
	keyHelp            = "?"
)

func InitialModel(filePath string) (UiModel, error) {
	m := UiModel{
		tabs:                 []string{"Texto", "Vocabulario", "Notas", "Estadísticas"},
		currentTab:           0,
		currentLine:          0,
		currentWordIdx:       0,
		currentVocabIdx:      0,
		currentNoteIdx:       0,
		currentLinkIdx:       0,
		selectedWord:         "",
		showGotoLineDialog:   false,
		lineInput:            "",
		tabWidths:            make([]int, 4), // Initialize for 4 tabs
		vocabulary:           []string{},
		notes:                []string{},
		showNoteDialog:       false,
		noteInput:            []string{""},
		showLinksDialog:      false,
		showDeleteNoteDialog: false,
		deleteNoteConfirmIdx: 0, // Default to "No"
		totalLines:           0,
		totalWords:           0,
		longestLine:          "",
		longestLineLength:    0,
		longestWord:          "",
		topWords:             []stats.WordCount{},
		cumulativeWords:      []int{},
		totalReadingSeconds:  0,
		totalReadWords:       0,
		sessionReadingTime:   0,
		sessionWordsRead:     0,
		lastActionTime:       time.Now(),
		showHelpDialog:       false,
	}

	m.filePath = filePath
	file, err := os.Open(m.filePath)
	if err != nil {
		return UiModel{}, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		m.lines = append(m.lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return UiModel{}, err
	}

	// Compute cumulative words
	m.cumulativeWords = make([]int, len(m.lines)+1)
	m.cumulativeWords[0] = 0
	for i := range m.lines {
		words := strings.Fields(m.lines[i])
		m.cumulativeWords[i+1] = m.cumulativeWords[i] + len(words)
	}
	m.totalWords = m.cumulativeWords[len(m.lines)]

	calculateStatistics(&m)

	// Load progress for the file
	line, vocab, notes, readingSeconds, readWords, err := progress.Load(m.filePath)
	if err != nil {
		return UiModel{}, err
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
	m.totalReadingSeconds = readingSeconds
	m.totalReadWords = readWords

	m.vp = viewport.New(0, 0) // Initialize to 0, Update() will set it.
	m.vp.MouseWheelEnabled = true
	m.vp.KeyMap = viewport.DefaultKeyMap()
	m.vp.KeyMap.Up.SetEnabled(false)
	m.vp.KeyMap.Down.SetEnabled(false)

	m.vp.SetContent(strings.Join(m.lines, "\n"))

	return m, nil
}

func calculateStatistics(m *UiModel) {
	m.totalLines = len(m.lines)
	var totalWords int
	var maxLen int
	var longest string
	var longestWordInLine string
	for _, line := range m.lines {
		words := strings.Fields(line)
		longestWord := stats.LongestWord(&words)
		if len(longestWord) > len(longestWordInLine) {
			longestWordInLine = longestWord
		}
		totalWords += len(words)
		if len(line) > maxLen {
			maxLen = len(line)
			longest = line
		}
	}

	m.totalWords = totalWords
	m.longestLine = longest
	m.longestLineLength = maxLen
	m.longestWord = longestWordInLine
	m.topWords = stats.TopNFrequentWords(m.lines)
}

func (m UiModel) Init() tea.Cmd {
	return nil
}

func (m UiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.showHelpDialog {
			switch msg.String() {
			case keyEsc, keyHelp, keyEnter, keyCancel:
				m.showHelpDialog = false
			}
			return m, nil
		}
		if m.showGotoLineDialog {
			switch msg.String() {
			case keyEsc:
				m.showGotoLineDialog = false
				m.lineInput = ""
			case keyEnter:
				if m.lineInput != "" {
					if lineNum, err := strconv.Atoi(m.lineInput); err == nil && lineNum > 0 && lineNum <= len(m.lines) {
						m.currentLine = lineNum - 1   // Convert to 0-based index
						m.currentWordIdx = 0          // Reset word index
						m.lastActionTime = time.Now() // Reset action time after jump
					}
				}
				m.showGotoLineDialog = false
				m.lineInput = ""
			case keyBackspace:
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
			case keyEsc:
				m.showNoteDialog = false
				m.noteInput = []string{""} // Reset note input
			case keyEnter:
				// Add new line to note input
				m.noteInput = append(m.noteInput, "")
			case keyBackspace:
				currentLine := len(m.noteInput) - 1
				if len(m.noteInput[currentLine]) > 0 {
					runes := []rune(m.noteInput[currentLine])
					m.noteInput[currentLine] = string(runes[:len(runes)-1]) // Elimina la última runa
				} else if currentLine > 0 {
					m.noteInput = m.noteInput[:currentLine]
				}
			case keyControlSave:
				// Save note
				note := strings.Join(m.noteInput, "\n")
				if note != "" {
					m.notes = append(m.notes, note)
				}
				m.showNoteDialog = false
				m.noteInput = []string{""} // Reset note input
			case keyCancel:
				// Cancel note
				m.showNoteDialog = false
				m.noteInput = []string{""} // Reset note input
			case keyEspace:
				currentLine := len(m.noteInput) - 1
				m.noteInput[currentLine] += " "
			default:
				if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
					currentLine := len(m.noteInput) - 1
					for _, r := range msg.Runes {
						m.noteInput[currentLine] += string(r)
					}
				}
			}
			return m, nil
		}
		if m.showLinksDialog {
			switch msg.String() {
			case keyEsc:
				m.showLinksDialog = false
				m.currentLinkIdx = 0
			case keyNextLine, "down":
				if m.currentLinkIdx < 1 { // Only two items: GoodReads, RAE
					m.currentLinkIdx++
				}
			case keyPrevLine, "up":
				if m.currentLinkIdx > 0 {
					m.currentLinkIdx--
				}
			case keyEnter:
				// Open the selected link in the default browser
				links := []string{"https://dle.rae.es/%s", "https://www.goodreads.com/search?q=%s"}
				if m.currentLinkIdx >= 0 && m.currentLinkIdx < len(links) {
					words := strings.Fields(m.lines[m.currentLine])
					var currentWord string
					if len(words) > 0 && m.currentWordIdx < len(words) {
						currentWord = words[m.currentWordIdx]
					}
					urlToSearch := fmt.Sprintf(links[m.currentLinkIdx], url.QueryEscape(text.SanitizeWord(currentWord)))
					if err := browserOpenURLCommand(runtime.GOOS, urlToSearch).Start(); err != nil {
						fmt.Printf("Error opening browser: %v\n", err)
					}
				}
				m.showLinksDialog = false
				m.currentLinkIdx = 0
			case keyCancel:
				// Cancel the dialog
				m.showLinksDialog = false
				m.currentLinkIdx = 0
			}
			return m, nil
		}
		if m.showDeleteNoteDialog {
			switch msg.String() {
			case keyEsc:
				m.showDeleteNoteDialog = false
				m.deleteNoteConfirmIdx = 0
			case keyLeft, "h":
				m.deleteNoteConfirmIdx = 0 // No
			case keyRight, "l":
				m.deleteNoteConfirmIdx = 1 // Yes
			case keyEnter:
				if m.deleteNoteConfirmIdx == 1 { // Yes selected
					// Delete the note
					if len(m.notes) > 0 && m.currentNoteIdx < len(m.notes) {
						m.notes = append(m.notes[:m.currentNoteIdx], m.notes[m.currentNoteIdx+1:]...)
						// Adjust currentNoteIdx if necessary
						if m.currentNoteIdx >= len(m.notes) && len(m.notes) > 0 {
							m.currentNoteIdx = len(m.notes) - 1
						}
						if len(m.notes) == 0 {
							m.currentNoteIdx = 0
						}
					}
				}
				m.showDeleteNoteDialog = false
				m.deleteNoteConfirmIdx = 0
			}
			return m, nil
		}
		switch msg.String() {
		case keyHelp:
			m.showHelpDialog = true
			return m, nil
		case keyCancel, keyQuit:
			// Save progress before quitting
			m.totalReadingSeconds += m.sessionReadingTime
			m.totalReadWords += m.sessionWordsRead
			if m.filePath != "" {
				if err := progress.Save(m.filePath, m.currentLine, m.vocabulary, m.notes, m.totalReadingSeconds, m.totalReadWords); err != nil {
					fmt.Printf("Error saving progress: %v\n", err)
				}
			}
			return m, tea.Quit
		case keySave:
			if m.filePath != "" {
				m.totalReadingSeconds += m.sessionReadingTime
				m.totalReadWords += m.sessionWordsRead
				if err := progress.Save(m.filePath, m.currentLine, m.vocabulary, m.notes, m.totalReadingSeconds, m.totalReadWords); err != nil {
					fmt.Printf("Error saving progress: %v\n", err)
				}
				m.sessionReadingTime = 0
				m.sessionWordsRead = 0
			}
		case keyMainTextTab:
			m.currentTab = 0
		case keyVocabTab:
			m.currentTab = 1
			m.currentVocabIdx = 0 // Reset vocabulary index when switching to Vocabulario tab
		case keyNotesTab:
			m.currentTab = 2
			m.currentNoteIdx = 0 // Reset note index when switching to Notas tab
		case keyStatsTab:
			m.currentTab = 3
		case keyShowNoteDialog:
			m.showNoteDialog = true
			m.noteInput = []string{""} // Initialize note input
		case keyGotoLineDialog:
			if m.currentTab == 0 {
				m.showGotoLineDialog = true
				m.lineInput = ""
			}
		case keyOpenLinksDialog:
			m.showLinksDialog = true
			m.currentLinkIdx = 0
		default:
			if m.currentTab == 0 {
				palabras := strings.Fields(m.lines[m.currentLine])
				switch msg.String() {
				case keyNextLine, "down":
					if m.currentLine < len(m.lines)-1 {
						delta := time.Since(m.lastActionTime).Seconds()
						if delta < 300 { // Ignore long idle periods (e.g., >5 min)
							m.sessionReadingTime += delta
							wordsInPrev := len(palabras)
							m.sessionWordsRead += wordsInPrev
						}
						m.lastActionTime = time.Now()
						//m.currentLine++
						m.currentLine = utils.Min(len(m.lines)-1, m.currentLine+1)
						m.currentWordIdx = 0
						m.syncViewportOffset() // Sincroniza viewport para centrar la nueva línea
					}
				case keyPrevLine, "up":
					if m.currentLine > 0 {
						m.lastActionTime = time.Now() // Update time but don't add to session (backtracking)
						//m.currentLine--
						m.currentLine = utils.Max(0, m.currentLine-1)
						m.currentWordIdx = 0
						m.syncViewportOffset() // Sincroniza
					}
				case keyLeft:
					if len(palabras) > 0 {
						m.currentWordIdx = (m.currentWordIdx - 1 + len(palabras)) % len(palabras)
					}
				case keyRight:
					if len(palabras) > 0 {
						m.currentWordIdx = (m.currentWordIdx + 1) % len(palabras)
					}
				case keyAddToVocabulary:
					if len(palabras) > 0 && m.currentWordIdx < len(palabras) {
						m.selectedWord = palabras[m.currentWordIdx]
						// Add to vocabulary if not already present:
						sanitizedWord := text.SanitizeWord(m.selectedWord)
						wordAlreadyInVocab := text.Contains(&m.vocabulary, sanitizedWord)
						if !wordAlreadyInVocab {
							m.vocabulary = append(m.vocabulary, sanitizedWord)
						}
					}
				case keyCopyToClipboard:
					if len(palabras) > 0 && m.currentWordIdx < len(palabras) {
						m.copiedToClipboardWord = palabras[m.currentWordIdx]
						err := clipboard.WriteAll(m.copiedToClipboardWord)
						if err != nil {
							fmt.Printf("Error copying word: %v\n", err)
						}
					}
				case keyZero:
					if len(palabras) > 0 {
						m.currentWordIdx = 0
					}
				case keyDollarSign:
					if len(palabras) > 0 {
						m.currentWordIdx = len(palabras) - 1
					}
				}

				// Delega a viewport.Update para keys default como pgup/pgdn (no para "j/k" ya que los manejas manualmente)
				var cmd tea.Cmd
				m.vp, cmd = m.vp.Update(msg)
				// Actualiza currentLine si YOffset cambió (ej. por pgup)
				halfHeight := m.vp.Height / 2
				m.currentLine = utils.Max(0, utils.Min(len(m.lines)-1, m.vp.YOffset+halfHeight))
				return m, cmd
			} else if m.currentTab == 1 {
				switch msg.String() {
				case keyNextLine:
					if m.currentVocabIdx < len(m.vocabulary)-1 {
						m.currentVocabIdx++
					}
				case keyPrevLine:
					if m.currentVocabIdx > 0 {
						m.currentVocabIdx--
					}
				case keyDelete:
					// Delete current vocabulary word
					if len(m.vocabulary) > 0 && m.currentVocabIdx < len(m.vocabulary) {
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
				case keyNextLine:
					if m.currentNoteIdx < len(m.notes)-1 {
						m.currentNoteIdx++
					}
				case keyPrevLine:
					if m.currentNoteIdx > 0 {
						m.currentNoteIdx--
					}
				case keyDelete:
					// Delete current note
					if len(m.notes) > 0 && m.currentNoteIdx < len(m.notes) {
						// Check if confirmation is required
						confirmDelete := os.Getenv("CONFIRM_NOTES_DELETE")
						if confirmDelete == "true" {
							// Show confirmation dialog
							m.showDeleteNoteDialog = true
							m.deleteNoteConfirmIdx = 0 // Default to "No"
						} else {
							// Delete immediately without confirmation
							m.notes = append(m.notes[:m.currentNoteIdx], m.notes[m.currentNoteIdx+1:]...)
							// Adjust currentNoteIdx if necessary
							if m.currentNoteIdx >= len(m.notes) && len(m.notes) > 0 {
								m.currentNoteIdx = len(m.notes) - 1
							}
							if len(m.notes) == 0 {
								m.currentNoteIdx = 0
							}
						}
					}
				}
			}
		}

	case tea.MouseMsg:
		if m.currentTab == 0 {
			// Delega a viewport para wheel up/down
			var cmd tea.Cmd
			m.vp, cmd = m.vp.Update(msg)
			// Actualiza currentLine al centro visible si YOffset cambió
			halfHeight := m.vp.Height / 2
			m.currentLine = utils.Max(0, utils.Min(len(m.lines)-1, m.vp.YOffset+halfHeight))
			return m, cmd
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		contentHeight := m.height - 5 // Ajuste existente para tab bar + status
		if contentHeight < 1 {
			contentHeight = 1
		}
		// 2 for borders/padding if any
		m.vp.Width = m.width - 2
		m.vp.Height = contentHeight
		// Sincroniza YOffset inicial para centrar currentLine (opcional, ver Paso 5)
		m.syncViewportOffset()
	}

	return m, nil
}

func (m *UiModel) syncViewportOffset() {
	halfHeight := m.vp.Height / 2
	newOffset := utils.Max(0, m.currentLine-halfHeight)
	maxOffset := utils.Max(0, len(m.lines)-m.vp.Height)
	m.vp.YOffset = utils.Min(newOffset, maxOffset)
}

func (m UiModel) View() string {
	if m.showHelpDialog {
		return m.renderWithDialog(m.renderHelpDialog())
	}
	if m.showGotoLineDialog {
		return m.renderWithDialog(m.renderGoToLineDialog())
	}
	if m.showNoteDialog {
		return m.renderWithDialog(m.renderNoteDialog())
	}
	if m.showLinksDialog {
		return m.renderWithDialog(m.renderLinksDialog())
	}
	if m.showDeleteNoteDialog {
		return m.renderWithDialog(m.renderDeleteNoteDialog())
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
				Foreground(brightWhiteColor).
				Background(blueColor).
				Border(lipgloss.RoundedBorder(), true).
				BorderForeground(cyanColor)
		} else {
			style = style.
				Italic(true).
				Foreground(lightGrayColor).
				Background(darkGrayColor).
				Border(lipgloss.NormalBorder(), true).
				BorderForeground(mediumGrayColor)
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
		Foreground(mediumGrayColor).
		Render(tabBar)
	content.WriteString(tabBar + "\n")

	// File label
	fileName := filepath.Base(m.filePath)
	fileLabel := lipgloss.NewStyle().
		Foreground(cyanColor).
		Background(darkGrayColor).
		Padding(0, 1).
		Width(m.width).
		Align(lipgloss.Left).
		Render(fmt.Sprintf("📄 %s", fileName))
	content.WriteString(fileLabel + "\n")

	// Content area
	contentHeight := m.height - 5 // Adjusted for tab bar height (3) + status (1)
	if contentHeight < 1 {
		contentHeight = 1
	}

	if m.currentTab == 0 {
		// Texto tab: show lines around current
		viewStart := m.vp.YOffset                                 // Usa offset del viewport
		viewEnd := utils.Min(len(m.lines), viewStart+m.vp.Height) // Visible height
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
							Background(brightYellowColor).
							Foreground(greyColor).
							Padding(0, 1).
							Render(word))
					} else {
						highlightedWords = append(highlightedWords, word)
					}
				}
				hlLine := strings.Join(highlightedWords, " ")
				// Apply line highlight
				hlLine = lipgloss.NewStyle().
					Background(darkGrayColor).
					Foreground(brightWhiteColor).
					Padding(0, 1).
					Render(hlLine)
				content.WriteString(hlLine + "\n")
			} else {
				content.WriteString(lipgloss.NewStyle().
					Foreground(lightGrayColor). // Light gray for non-current lines
					Render(m.lines[i]) + "\n")
			}
		}
	} else if m.currentTab == 1 {
		// Vocabulario tab: show vocabulary words with navigation
		if len(m.vocabulary) == 0 {
			content.WriteString(lipgloss.NewStyle().
				Foreground(lightGrayColor).
				Align(lipgloss.Right).
				Render("No Vocab\n"))
		} else {
			viewStart := utils.Max(0, m.currentVocabIdx-contentHeight/2)
			viewEnd := utils.Min(len(m.vocabulary), viewStart+contentHeight)
			for i := viewStart; i < viewEnd; i++ {
				word := m.vocabulary[i]
				if i == m.currentVocabIdx {
					content.WriteString(lipgloss.NewStyle().
						Background(darkGrayColor).
						Foreground(brightWhiteColor).
						Padding(0, 1).
						Render(word) + "\n")
				} else {
					content.WriteString(lipgloss.NewStyle().
						Foreground(lightGrayColor). // Light gray
						Render(word) + "\n")
				}
			}
		}
	} else if m.currentTab == 2 {
		// Notas tab: show notes with navigation and borders
		if len(m.notes) == 0 {
			content.WriteString(lipgloss.NewStyle().
				Foreground(lightGrayColor).
				Render("No Notes\n"))
		} else {
			viewStart := utils.Max(0, m.currentNoteIdx-contentHeight/2)
			viewEnd := utils.Min(len(m.notes), viewStart+contentHeight)
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
					BorderForeground(mediumGrayColor). // Medium gray border
					Padding(0, 1)
				if i == m.currentNoteIdx {
					noteStyle = noteStyle.
						Background(greyColor). // Darker gray background
						Foreground(brightWhiteColor) // Bright white text
				} else {
					noteStyle = noteStyle.
						Foreground(lightGrayColor) // Light gray
				}
				renderedNote := noteStyle.Render(noteText)
				renderedNotes = append(renderedNotes, renderedNote)
				usedHeight += noteHeight + 2 // Add border height
			}
			// Join notes vertically with a newline separator
			content.WriteString(lipgloss.JoinVertical(lipgloss.Left, renderedNotes...) + "\n")
		}
	} else if m.currentTab == 3 {
		// Estadísticas tab: show file statistics
		boldStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(brightWhiteColor).
			Background(greenColor)

		italicStyle := lipgloss.NewStyle().
			Italic(true)

		wpm := m.getCurrentWPM()

		statsLines := []string{
			"Líneas totales: " + boldStyle.Render(fmt.Sprintf("%d", m.totalLines)),
			"Palabras totales: " + boldStyle.Render(fmt.Sprintf("%d", m.totalWords)),
			"Línea más larga: " + boldStyle.Render(fmt.Sprintf("%d caracteres", m.longestLineLength)),
			italicStyle.Render(m.longestLine),
			"Palabra más larga: " + boldStyle.Render(m.longestWord),
			"Velocidad de lectura: " + boldStyle.Render(fmt.Sprintf("%.0f WPM", wpm)),
		}
		if len(m.topWords) > 0 {
			statsLines = append(statsLines, "Top palabras frecuentes:")
			for i, wc := range m.topWords {
				statsLines = append(statsLines, fmt.Sprintf("%d. %s: %d", i+1, wc.Word, wc.Count))
			}
		}
		statsText := strings.Join(statsLines, "\n")
		statsStyle := lipgloss.NewStyle().
			Foreground(brightWhiteColor). // Bright white
			Background(darkGrayColor). // Darker gray
			Padding(1, 2).
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(cyanColor) // Cyan border
		content.WriteString(statsStyle.Render(statsText) + "\n")
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
	timeLeft := m.remainingTimeString()
	status := lineInfo + selInfo + " | Tiempo restante: " + timeLeft
	statusStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Height(1).
		Bold(true).
		Foreground(brightWhiteColor).
		Background(darkGrayColor). // Very dark gray
		Width(m.width).
		Align(lipgloss.Left)
	content.WriteString(statusStyle.Render(status))

	return content.String()
}

func (m UiModel) getCurrentWPM() float64 {
	totalSec := m.totalReadingSeconds + m.sessionReadingTime
	totalWords := m.totalReadWords + m.sessionWordsRead
	if totalSec > 0 && totalWords > 0 {
		return float64(totalWords) / (totalSec / 60)
	}
	return DefaultWPM
}

func (m UiModel) remainingTimeString() string {
	wpm := m.getCurrentWPM()
	wordsRead := m.cumulativeWords[m.currentLine] + m.sessionWordsRead // Include session progress
	wordsLeft := m.totalWords - wordsRead
	if wordsLeft <= 0 {
		return "0m"
	}
	minutesLeft := float64(wordsLeft) / wpm
	hours := int(minutesLeft / 60)
	mins := int(minutesLeft) % 60
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dm", mins)
}

func (m UiModel) renderGoToLineDialog() string {
	dialogContent := fmt.Sprintf("Go to line: %s", m.lineInput)

	dialog := lipgloss.NewStyle().
		Width(30).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(royalBlueColor).
		Padding(1, 2).
		Align(lipgloss.Center).
		Background(greyColor).
		Render(dialogContent)

	return dialog
}

func (m UiModel) renderNoteDialog() string {
	dialogWidth := utils.Min(m.width*3/4, m.width-4)
	if dialogWidth > 80 {
		dialogWidth = 80
	}

	title := lipgloss.NewStyle().
		Foreground(brightWhiteColor).
		Align(lipgloss.Center).
		Padding(0, 0).
		Width(dialogWidth - 4).
		Render("Agregar nota")

	noteText := strings.Join(m.noteInput, "\n")
	inputBox := lipgloss.NewStyle().
		Width(dialogWidth - 4).
		Height(10).
		Border(lipgloss.NormalBorder()).
		BorderForeground(royalBlueColor).
		Padding(1).
		Render(noteText)

	saveButton := lipgloss.NewStyle().
		Foreground(brightWhiteColor).
		Background(greenColor).
		Padding(0, 2).
		Margin(0, 1).
		Render("Guardar (Ctrl+S)")

	cancelButton := lipgloss.NewStyle().
		Foreground(brightWhiteColor).
		Background(redColor).
		Padding(0, 2).
		Margin(0, 1).
		Render("Cancelar (Ctrl+C)")

	buttons := lipgloss.JoinHorizontal(lipgloss.Center, saveButton, cancelButton)
	dialogContent := lipgloss.JoinVertical(lipgloss.Left, title, inputBox, buttons)

	dialog := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(royalBlueColor).
		Padding(1).
		Background(greyColor).
		Render(dialogContent)

	return dialog
}

func (m UiModel) renderLinksDialog() string {
	dialogWidth := 40

	title := lipgloss.NewStyle().
		Foreground(brightWhiteColor).
		Align(lipgloss.Center).
		Padding(0, 0).
		Width(dialogWidth - 4).
		Render("Seleccionar Enlace")

	links := []string{"Real Academia Española", "GoodReads"}
	var linkItems []string
	for i, link := range links {
		style := lipgloss.NewStyle().
			Width(dialogWidth-6).
			Padding(0, 1).
			Align(lipgloss.Left)
		if i == m.currentLinkIdx {
			style = style.
				Background(darkGrayColor).
				Foreground(brightWhiteColor)
		} else {
			style = style.Foreground(lightGrayColor)
		}
		linkItems = append(linkItems, style.Render(link))
	}
	linksList := lipgloss.JoinVertical(lipgloss.Left, linkItems...)

	listBox := lipgloss.NewStyle().
		Width(dialogWidth-4).
		Border(lipgloss.NormalBorder()).
		BorderForeground(royalBlueColor).
		Padding(0, 1).
		Render(linksList)

	goButton := lipgloss.NewStyle().
		Foreground(brightWhiteColor).
		Background(greenColor).
		Padding(0, 2).
		Margin(0, 1).
		Render("Ir A (Enter)")

	cancelButton := lipgloss.NewStyle().
		Foreground(brightWhiteColor).
		Background(redColor).
		Padding(0, 2).
		Margin(0, 1).
		Render("Cancelar (Ctrl+C)")

	buttons := lipgloss.JoinHorizontal(lipgloss.Center, goButton, cancelButton)
	dialogContent := lipgloss.JoinVertical(lipgloss.Left, title, listBox, buttons)

	dialog := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(royalBlueColor).
		Padding(1).
		Background(greyColor).
		Render(dialogContent)

	return dialog
}

func (m UiModel) renderDeleteNoteDialog() string {
	dialogWidth := 50

	title := lipgloss.NewStyle().
		Foreground(brightWhiteColor).
		Align(lipgloss.Center).
		Padding(0, 0).
		Width(dialogWidth - 4).
		Bold(true).
		Render("¿Estás seguro de eliminar esta Nota?")

	// Buttons: No and Yes
	noButton := lipgloss.NewStyle().
		Foreground(brightWhiteColor).
		Padding(0, 3).
		Margin(0, 2)
	yesButton := lipgloss.NewStyle().
		Foreground(brightWhiteColor).
		Padding(0, 3).
		Margin(0, 2)

	if m.deleteNoteConfirmIdx == 0 {
		// No is selected
		noButton = noButton.
			Background(redColor).
			Bold(true)
		yesButton = yesButton.
			Background(grayColor)
	} else {
		// Yes is selected
		noButton = noButton.
			Background(grayColor)
		yesButton = yesButton.
			Background(greenColor).
			Bold(true)
	}

	buttons := lipgloss.JoinHorizontal(
		lipgloss.Center,
		noButton.Render("No"),
		yesButton.Render("Sí"),
	)

	hint := lipgloss.NewStyle().
		Foreground(mediumGrayColor).
		Align(lipgloss.Center).
		Width(dialogWidth - 4).
		Render("← → para navegar | Enter para confirmar | Esc para cancelar")

	dialogContent := lipgloss.JoinVertical(lipgloss.Center, title, "", buttons, "", hint)

	dialog := lipgloss.NewStyle().
		Width(dialogWidth).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(redColor). // Red border
		Padding(1).
		Background(greyColor).
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

func (m UiModel) renderHelpDialog() string {
	dialogWidth := utils.Min(m.width-4, 70)

	title := lipgloss.NewStyle().
		Foreground(brightWhiteColor).
		Background(blueColor).
		Bold(true).
		Align(lipgloss.Center).
		Padding(0, 1).
		Width(dialogWidth - 4).
		Render("⌨️  ATAJOS DE TECLADO")

	// Definir secciones de ayuda
	sections := []struct {
		title string
		keys  [][]string
	}{
		{
			title: "NAVEGACIÓN",
			keys: [][]string{
				{"j / ↓", "Línea siguiente"},
				{"k / ↑", "Línea anterior"},
				{"← / →", "Palabra anterior/siguiente"},
				{"0", "Primera palabra de la línea"},
				{"$", "Última palabra de la línea"},
				{"g", "Ir a línea específica"},
				{"PgUp/PgDn", "Página arriba/abajo"},
			},
		},
		{
			title: "TABS",
			keys: [][]string{
				{"1", "Tab Texto"},
				{"2", "Tab Vocabulario"},
				{"3", "Tab Notas"},
				{"4", "Tab Estadísticas"},
			},
		},
		{
			title: "ACCIONES",
			keys: [][]string{
				{"w", "Agregar palabra al vocabulario"},
				{"c", "Copiar palabra al portapapeles"},
				{"n", "Crear nueva nota"},
				{"o", "Abrir enlaces (RAE/GoodReads)"},
				{"d", "Eliminar (vocabulario/nota)"},
				{"s", "Guardar progreso"},
			},
		},
		{
			title: "GENERAL",
			keys: [][]string{
				{"?", "Mostrar esta ayuda"},
				{"Esc", "Cerrar diálogos"},
				{"q / Ctrl+C", "Salir"},
			},
		},
	}

	var helpContent strings.Builder

	keyStyle := lipgloss.NewStyle().
		Foreground(brightYellowColor).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(lightGrayColor)

	sectionTitleStyle := lipgloss.NewStyle().
		Foreground(cyanColor).
		Bold(true).
		Underline(true).
		MarginTop(1)

	for _, section := range sections {
		helpContent.WriteString(sectionTitleStyle.Render(section.title) + "\n")
		for _, key := range section.keys {
			line := fmt.Sprintf("  %s  %s",
				keyStyle.Render(fmt.Sprintf("%-12s", key[0])),
				descStyle.Render(key[1]))
			helpContent.WriteString(line + "\n")
		}
	}

	contentBox := lipgloss.NewStyle().
		Width(dialogWidth - 4).
		MaxHeight(m.height - 10).
		Padding(1).
		Render(helpContent.String())

	closeHint := lipgloss.NewStyle().
		Foreground(mediumGrayColor).
		Italic(true).
		Align(lipgloss.Center).
		Width(dialogWidth - 4).
		Render("Presiona Esc o ? para cerrar")

	dialogContent := lipgloss.JoinVertical(lipgloss.Left, title, contentBox, closeHint)

	dialog := lipgloss.NewStyle().
		Width(dialogWidth).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(cyanColor).
		Padding(1).
		Background(greyColor).
		Render(dialogContent)

	return dialog
}

func browserOpenURLCommand(osName, url string) *exec.Cmd {
	switch osName {
	case "linux":
		return exec.Command("xdg-open", url)
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		return exec.Command("open", url)
	default:
		return nil
	}
}
