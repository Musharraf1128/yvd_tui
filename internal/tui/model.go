package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/Musharraf1128/yvd_tui/internal/downloader"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ---------------------------------------------------------------------------
// State machine
// ---------------------------------------------------------------------------

type appState int

const (
	stateFetching appState = iota
	stateQualityPick
	stateFormatPick
	stateConfirm
	stateError
)

// ---------------------------------------------------------------------------
// Messages
// ---------------------------------------------------------------------------

type fetchDoneMsg struct {
	info *downloader.VideoInfo
	err  error
}

// ---------------------------------------------------------------------------
// Domain types
// ---------------------------------------------------------------------------

// QualityOption is a selectable quality in the picker.
type QualityOption struct {
	Label     string // "1080p", "720p", "audio only"
	Height    int    // 0 for audio-only / "best"
	AudioOnly bool
}

// DownloadSelection is the result returned when the user confirms.
type DownloadSelection struct {
	URL        string
	Quality    QualityOption
	FormatType string // "video + audio" | "video only" | "audio only"
	Info       *downloader.VideoInfo // carried forward for the download TUI
}

// ToDownloadOptions converts the selection into downloader.DownloadOptions.
func (s *DownloadSelection) ToDownloadOptions() downloader.DownloadOptions {
	opts := downloader.DownloadOptions{
		URL: s.URL,
	}
	if s.Quality.AudioOnly || s.FormatType == "audio only" {
		opts.AudioOnly = true
	} else if s.FormatType == "video only" {
		opts.VideoOnly = true
	}
	if s.Quality.Height > 0 {
		opts.Quality = fmt.Sprintf("%d", s.Quality.Height)
	}
	return opts
}

// ---------------------------------------------------------------------------
// Model
// ---------------------------------------------------------------------------

// Model is the root Bubbletea model.
type Model struct {
	state   appState
	url     string
	info    *downloader.VideoInfo
	err     error
	spinner spinner.Model

	// Quality picker
	qualities     []QualityOption
	qualityCursor int

	// Format picker (only for video qualities)
	formatTypes  []string
	formatCursor int

	// Confirmed selections
	selectedQuality QualityOption
	selectedFormat  string

	// Terminal dimensions
	width  int
	height int

	// Final result (nil until user confirms)
	Selection *DownloadSelection
	Cancelled bool
}

// New creates a fresh model ready to fetch info for the given URL.
func New(url string) Model {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	s.Style = spinnerStyle

	return Model{
		state:       stateFetching,
		url:         url,
		spinner:     s,
		formatTypes: []string{"video + audio", "video only"},
	}
}

// ---------------------------------------------------------------------------
// Init
// ---------------------------------------------------------------------------

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		doFetch(m.url),
	)
}

func doFetch(url string) tea.Cmd {
	return func() tea.Msg {
		info, err := downloader.FetchInfo(context.Background(), url)
		return fetchDoneMsg{info: info, err: err}
	}
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Global quit keys
		switch msg.String() {
		case "ctrl+c":
			m.Cancelled = true
			return m, tea.Quit
		case "q":
			if m.state != stateFetching {
				m.Cancelled = true
				return m, tea.Quit
			}
		}

		// Delegate to the active screen
		switch m.state {
		case stateQualityPick:
			return m.updateQualityPick(msg)
		case stateFormatPick:
			return m.updateFormatPick(msg)
		case stateConfirm:
			return m.updateConfirm(msg)
		case stateError:
			return m, tea.Quit
		}

	case fetchDoneMsg:
		if msg.err != nil {
			m.err = msg.err
			m.state = stateError
			return m, nil
		}
		m.info = msg.info
		m.qualities = buildQualityOptions(msg.info)
		m.qualityCursor = 0
		m.state = stateQualityPick
		return m, nil

	case spinner.TickMsg:
		if m.state == stateFetching {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

// --- screen-specific update handlers ---

func (m Model) updateQualityPick(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.qualityCursor > 0 {
			m.qualityCursor--
		}
	case "down", "j":
		if m.qualityCursor < len(m.qualities)-1 {
			m.qualityCursor++
		}
	case "enter", " ":
		m.selectedQuality = m.qualities[m.qualityCursor]
		if m.selectedQuality.AudioOnly {
			// Skip format picker — no choice needed
			m.selectedFormat = "audio only"
			m.state = stateConfirm
		} else {
			m.formatCursor = 0
			m.state = stateFormatPick
		}
	}
	return m, nil
}

func (m Model) updateFormatPick(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.formatCursor > 0 {
			m.formatCursor--
		}
	case "down", "j":
		if m.formatCursor < len(m.formatTypes)-1 {
			m.formatCursor++
		}
	case "enter", " ":
		m.selectedFormat = m.formatTypes[m.formatCursor]
		m.state = stateConfirm
	case "esc", "b":
		m.state = stateQualityPick
	}
	return m, nil
}

func (m Model) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "y":
		m.Selection = &DownloadSelection{
			URL:        m.url,
			Quality:    m.selectedQuality,
			FormatType: m.selectedFormat,
			Info:       m.info,
		}
		return m, tea.Quit
	case "esc", "b", "n":
		if m.selectedQuality.AudioOnly {
			m.state = stateQualityPick
		} else {
			m.state = stateFormatPick
		}
	}
	return m, nil
}

// ---------------------------------------------------------------------------
// View
// ---------------------------------------------------------------------------

func (m Model) View() string {
	if m.Selection != nil || m.Cancelled {
		return ""
	}
	switch m.state {
	case stateFetching:
		return m.viewFetching()
	case stateQualityPick:
		return m.viewQualityPick()
	case stateFormatPick:
		return m.viewFormatPick()
	case stateConfirm:
		return m.viewConfirm()
	case stateError:
		return m.viewError()
	}
	return ""
}

func banner() string {
	return headerStyle.Render("YVD — YouTube Video Downloader")
}

func (m Model) viewFetching() string {
	status := fmt.Sprintf("%s Fetching video info…", m.spinner.View())
	url := subtitleStyle.Render(clip(m.url, 70))
	help := helpStyle.Render("ctrl+c to quit")

	return appStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		banner(), url, "", status, "", help,
	))
}

func (m Model) viewQualityPick() string {
	var b strings.Builder

	b.WriteString(banner())
	b.WriteString("\n")
	b.WriteString(m.infoBox())
	b.WriteString("\n")
	b.WriteString(sectionHeaderStyle.Render("Select Quality"))
	b.WriteString("\n")

	for i, q := range m.qualities {
		var row string
		if i == m.qualityCursor {
			row = cursorStyle.Render("❯ ") + selectedItemStyle.Render(q.Label)
		} else {
			row = "  " + normalItemStyle.Render(q.Label)
		}
		b.WriteString(row + "\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓  navigate   •   enter  select   •   q  quit"))

	return appStyle.Render(b.String())
}

func (m Model) viewFormatPick() string {
	var b strings.Builder

	b.WriteString(banner())
	b.WriteString("\n")
	b.WriteString(subtitleStyle.Render(fmt.Sprintf("Quality selected: %s", m.selectedQuality.Label)))
	b.WriteString("\n\n")
	b.WriteString(sectionHeaderStyle.Render("Select Format"))
	b.WriteString("\n")

	for i, ft := range m.formatTypes {
		var row string
		if i == m.formatCursor {
			row = cursorStyle.Render("❯ ") + selectedItemStyle.Render(ft)
		} else {
			row = "  " + normalItemStyle.Render(ft)
		}
		b.WriteString(row + "\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓  navigate   •   enter  select   •   esc  back   •   q  quit"))

	return appStyle.Render(b.String())
}

func (m Model) viewConfirm() string {
	var b strings.Builder

	b.WriteString(banner())
	b.WriteString("\n")

	summary := lipgloss.JoinVertical(lipgloss.Left,
		infoLabelStyle.Render("Title")+infoValueStyle.Render(clip(m.info.Title, 55)),
		infoLabelStyle.Render("Quality")+infoValueStyle.Render(m.selectedQuality.Label),
		infoLabelStyle.Render("Format")+infoValueStyle.Render(m.selectedFormat),
	)
	b.WriteString(infoBoxStyle.Render(summary))
	b.WriteString("\n")

	b.WriteString(sectionHeaderStyle.Render("Ready to download"))
	b.WriteString("\n  ")
	b.WriteString("Press " + successStyle.Render("enter") + " to start  or  " + errorStyle.Render("esc") + " to go back\n")
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("enter/y  confirm   •   esc/b  back   •   q  quit"))

	return appStyle.Render(b.String())
}

func (m Model) viewError() string {
	pretty := downloader.PrettifyError(m.err)
	errMsg := errorStyle.Render("Error: " + pretty)
	help := helpStyle.Render("press any key to exit")

	return appStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		banner(), "", errMsg, "", help,
	))
}

// infoBox renders the video metadata card.
func (m Model) infoBox() string {
	content := lipgloss.JoinVertical(lipgloss.Left,
		infoLabelStyle.Render("Title")+infoValueStyle.Render(clip(m.info.Title, 55)),
		infoLabelStyle.Render("Uploader")+infoValueStyle.Render(m.info.Uploader),
		infoLabelStyle.Render("Duration")+infoValueStyle.Render(m.info.Duration),
	)
	return infoBoxStyle.Render(content)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// clip truncates s with an ellipsis if it exceeds maxLen.
func clip(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// buildQualityOptions derives the unique quality levels from raw formats.
func buildQualityOptions(info *downloader.VideoInfo) []QualityOption {
	heights := make(map[int]bool)
	hasAudioOnly := false

	for _, f := range info.Formats {
		if f.HasVideo {
			h := parseHeight(f.Resolution)
			if h > 0 {
				heights[h] = true
			}
		}
		if f.HasAudio && !f.HasVideo {
			hasAudioOnly = true
		}
	}

	// Sort heights descending (best quality first)
	var sorted []int
	for h := range heights {
		sorted = append(sorted, h)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(sorted)))

	var opts []QualityOption
	for _, h := range sorted {
		opts = append(opts, QualityOption{
			Label:  fmt.Sprintf("%dp", h),
			Height: h,
		})
	}

	// Fallback: if no video heights parsed, add a "best" option
	if len(opts) == 0 {
		opts = append(opts, QualityOption{Label: "best (auto)", Height: 0})
	}

	// Audio only at the bottom
	if hasAudioOnly || len(opts) > 0 {
		opts = append(opts, QualityOption{
			Label:     "audio only",
			AudioOnly: true,
		})
	}

	return opts
}

// parseHeight extracts the height from resolutions like "1920x1080" or "1080p".
func parseHeight(resolution string) int {
	var h int
	if strings.Contains(resolution, "x") {
		parts := strings.Split(resolution, "x")
		if len(parts) == 2 {
			fmt.Sscanf(parts[1], "%d", &h)
		}
	} else {
		fmt.Sscanf(strings.TrimSuffix(resolution, "p"), "%d", &h)
	}
	return h
}
