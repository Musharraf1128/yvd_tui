package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/Musharraf1128/yvd_tui/internal/downloader"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ---------------------------------------------------------------------------
// Messages
// ---------------------------------------------------------------------------

// ProgressMsg carries a real-time progress snapshot from the download goroutine.
type ProgressMsg struct {
	Progress downloader.ProgressInfo
}

// DownloadDoneMsg is sent by the download goroutine when it finishes.
type DownloadDoneMsg struct {
	Err error
}

// ---------------------------------------------------------------------------
// Model
// ---------------------------------------------------------------------------

// DownloadModel is a Bubbletea model that renders a live download progress screen.
// info may be nil when started from direct (non-interactive) mode.
type DownloadModel struct {
	info    *downloader.VideoInfo
	sel     *DownloadSelection
	latest  downloader.ProgressInfo
	bar     progress.Model
	spinner spinner.Model

	done    bool
	doneErr error
	started bool // true once first ProgressMsg received

	width int
}

// NewDownload returns a ready-to-run download progress model.
func NewDownload(info *downloader.VideoInfo, sel *DownloadSelection) DownloadModel {
	bar := progress.New(
		progress.WithSolidFill("#E88158"), // Orange
		progress.WithWidth(70),
	)

	s := spinner.New()
	s.Spinner = spinner.MiniDot
	s.Style = spinnerStyle

	return DownloadModel{
		info:    info,
		sel:     sel,
		bar:     bar,
		spinner: s,
		width:   80,
	}
}

func (m DownloadModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m DownloadModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		bw := msg.Width - 8
		if bw < 20 {
			bw = 20
		}
		m.bar.Width = bw
		return m, nil

	case tea.KeyMsg:
		// Any key exits when download is finished or errored.
		if m.done || m.doneErr != nil {
			return m, tea.Quit
		}
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case ProgressMsg:
		m.started = true
		m.latest = msg.Progress
		pct := clamp01(msg.Progress.Percent / 100.0)
		return m, m.bar.SetPercent(pct)

	case DownloadDoneMsg:
		if msg.Err != nil {
			m.doneErr = msg.Err
			return m, tea.Quit
		}
		m.done = true
		return m, tea.Batch(m.bar.SetPercent(1.0), tea.Quit)

	case progress.FrameMsg:
		newBar, cmd := m.bar.Update(msg)
		m.bar = newBar.(progress.Model)
		return m, cmd

	case spinner.TickMsg:
		if !m.started && !m.done && m.doneErr == nil {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

// ---------------------------------------------------------------------------
// View
// ---------------------------------------------------------------------------

func (m DownloadModel) View() string {
	var b strings.Builder

	b.WriteString(banner())
	b.WriteString("\n")

	// Info / meta box.
	if m.info != nil && m.sel != nil {
		content := lipgloss.JoinVertical(lipgloss.Left,
			infoLabelStyle.Render("Title")+infoValueStyle.Render(clip(m.info.Title, 55)),
			infoLabelStyle.Render("Quality")+infoValueStyle.Render(m.sel.Quality.Label),
			infoLabelStyle.Render("Format")+infoValueStyle.Render(m.sel.FormatType),
		)
		b.WriteString(infoBoxStyle.Render(content))
		b.WriteString("\n")
	} else if m.sel != nil {
		b.WriteString(subtitleStyle.Render(clip(m.sel.URL, 70)))
		b.WriteString("\n\n")
	}

	// ─── Error ───────────────────────────────────────────────────────────────
	if m.doneErr != nil {
		pretty := downloader.PrettifyError(m.doneErr)
		b.WriteString(errorStyle.Render("Error:"))
		b.WriteString("\n")
		b.WriteString("  " + infoValueStyle.Render(pretty))
		return appStyle.Render(b.String())
	}

	// ─── Success ─────────────────────────────────────────────────────────────
	if m.done {
		b.WriteString(successStyle.Render("Download complete!"))
		b.WriteString("\n")
		if m.latest.Filename != "" {
			name := filepath.Base(m.latest.Filename)
			b.WriteString("   " + infoValueStyle.Render("Saved: "+clip(name, 65)))
		}
		return appStyle.Render(b.String())
	}

	// ─── Downloading / Starting ───────────────────────────────────────────────
	if !m.started {
		b.WriteString(fmt.Sprintf("  %s  Preparing download…", m.spinner.View()))
	} else {
		b.WriteString(sectionHeaderStyle.Render(dlStatusLabel(m.latest.Status)))
		b.WriteString("\n\n")

		// Progress bar (full-width minus padding).
		b.WriteString("  ")
		b.WriteString(m.bar.View())
		b.WriteString("\n\n")

		// Stats row.
		stats := []string{fmt.Sprintf("%.1f%%", m.latest.Percent)}
		if m.latest.Speed > 0 {
			stats = append(stats, fmtBytes(m.latest.Speed)+"/s")
		}
		if m.latest.ETA > 0 {
			stats = append(stats, "ETA "+fmtDur(m.latest.ETA))
		}
		if m.latest.TotalBytes > 0 {
			stats = append(stats,
				fmtBytes(float64(m.latest.DownloadedBytes))+" / "+fmtBytes(float64(m.latest.TotalBytes)),
			)
		}
		b.WriteString("  " + infoValueStyle.Render(strings.Join(stats, "  •  ")))
		b.WriteString("\n")

		// Current filename.
		if m.latest.Filename != "" {
			name := filepath.Base(m.latest.Filename)
			b.WriteString("  " + subtitleStyle.Render(clip(name, 72)))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("ctrl+c to cancel"))
	return appStyle.Render(b.String())
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func dlStatusLabel(status string) string {
	switch status {
	case "starting":
		return "Starting…"
	case "downloading":
		return "Downloading…"
	case "post_processing":
		return "Post-processing (ffmpeg)…"
	case "finished":
		return "Finishing…"
	default:
		return "Downloading…"
	}
}

// fmtBytes formats a byte count as a human-readable string (e.g. "12.3 MiB").
func fmtBytes(b float64) string {
	const unit = 1024.0
	if b < unit {
		return fmt.Sprintf("%.0f B", b)
	}
	div, exp := unit, 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", b/div, "KMGTPE"[exp])
}

// fmtDur formats a time.Duration as a short human-readable string.
func fmtDur(d time.Duration) string {
	d = d.Round(time.Second)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh%02dm%02ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm%02ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
