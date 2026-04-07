package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Musharraf1128/yvd_tui/internal/downloader"
	"github.com/Musharraf1128/yvd_tui/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	quality string
	format  string
)

var rootCmd = &cobra.Command{
	Use:   "yvd [video-url]",
	Short: "YVD — YouTube Video Downloader TUI",
	Long: `YVD is a terminal-based video downloader with an interactive TUI.

Download videos by providing a URL. If no flags are given, an interactive
picker will let you choose quality and format. Supply both --quality and
--format to download directly without any prompts.

NOTE: Always quote URLs to prevent shell glob expansion on '?':

Examples:
  yvd 'https://www.youtube.com/watch?v=dQw4w9WgXcQ'
  yvd 'https://www.youtube.com/watch?v=dQw4w9WgXcQ' -q 1080 -f mp4
  yvd 'https://www.youtube.com/watch?v=dQw4w9WgXcQ' -q 0   -f mp3`,
	Args: cobra.MaximumNArgs(1),
	Run:  runRoot,
}

func init() {
	rootCmd.Flags().StringVarP(&quality, "quality", "q", "", "Max video height (e.g. 1080, 720). Use 0 for audio-only.")
	rootCmd.Flags().StringVarP(&format, "format", "f", "", "Container/output format (e.g. mp4, webm, mp3, m4a)")
}

func runRoot(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println("YVD — YouTube Video Downloader")
		fmt.Println()
		fmt.Println("Usage: yvd <video-url> [flags]")
		fmt.Println()
		fmt.Println("Run 'yvd --help' for more information.")
		return
	}

	url := args[0]

	// Ensure yt-dlp is present before doing anything else.
	fmt.Print("Checking yt-dlp... ")
	if err := downloader.EnsureInstalled(context.Background()); err != nil {
		fmt.Printf("FAILED\n   %v\n", err)
		os.Exit(1)
	}
	fmt.Println("OK")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ─── Direct mode (Phase 5) ─────────────────────────────────────────────────
	// Both --quality and --format provided → skip TUI completely.
	if quality != "" && format != "" {
		runDirectMode(ctx, url, quality, format)
		return
	}

	// ─── Interactive mode ──────────────────────────────────────────────────────
	// Step 1: Selection TUI — fetch metadata → pick quality → pick format → confirm.
	selModel := tui.New(url)
	p1 := tea.NewProgram(selModel)

	finalSel, err := p1.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	result, ok := finalSel.(tui.Model)
	if !ok || result.Cancelled || result.Selection == nil {
		fmt.Println("Download cancelled.")
		return
	}

	sel := result.Selection
	opts := sel.ToDownloadOptions()
	opts.OutputDir = getDefaultOutputDir()

	// Step 2: Download TUI — progress bar with real-time stats.
	dlModel := tui.NewDownload(sel.Info, sel)
	p2 := tea.NewProgram(dlModel)

	// Download runs in its own goroutine; progress messages are sent to p2.
	go func() {
		dlErr := downloader.Download(ctx, opts, func(pi downloader.ProgressInfo) {
			p2.Send(tui.ProgressMsg{Progress: pi})
		})
		p2.Send(tui.DownloadDoneMsg{Err: dlErr})
	}()

	if _, err := p2.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

// ---------------------------------------------------------------------------
// Direct mode (Phase 5)
// ---------------------------------------------------------------------------

// audioFormats is the set of format identifiers treated as audio-only.
var audioFormats = map[string]bool{
	"mp3": true, "m4a": true, "aac": true,
	"flac": true, "wav": true, "ogg": true, "opus": true,
}

// runDirectMode downloads without launching the interactive TUI.
// It renders a simple ASCII progress bar that plays well with scripts / redirection.
func runDirectMode(ctx context.Context, url, qualityStr, formatStr string) {
	isAudio := audioFormats[strings.ToLower(formatStr)]

	opts := downloader.DownloadOptions{
		URL:          url,
		AudioOnly:    isAudio,
		OutputFormat: formatStr,
		OutputDir:    getDefaultOutputDir(),
	}
	if !isAudio {
		opts.Quality = qualityStr
	}

	fmt.Printf("\nDownloading directly — quality: %sp  format: %s\n", qualityStr, formatStr)
	fmt.Printf("   %s\n\n", url)

	var lastLine string
	dlErr := downloader.Download(ctx, opts, func(pi downloader.ProgressInfo) {
		var line string
		switch pi.Status {
		case "post_processing":
			line = "   [post-processing (ffmpeg)…]                              "
		case "finished":
			line = "   [finishing…]                                             "
		default:
			bar := asciiBar(pi.Percent, 30)
			spd := ""
			if pi.Speed > 0 {
				spd = " @ " + cmdFmtBytes(pi.Speed) + "/s"
			}
			eta := ""
			if pi.ETA > 0 {
				eta = " ETA " + cmdFmtDur(pi.ETA)
			}
			line = fmt.Sprintf("   [%s] %5.1f%%%s%s", bar, pi.Percent, spd, eta)
		}
		if line != lastLine {
			fmt.Printf("\r%s", line)
			lastLine = line
		}
	})

	fmt.Println() // newline after inline progress
	if dlErr != nil {
		fmt.Printf("Error: %v\n", dlErr)
		os.Exit(1)
	}
	fmt.Println("Download complete!")
}

// asciiBar builds a simple block-character progress bar string of the given width.
func asciiBar(pct float64, width int) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	filled := int(pct / 100.0 * float64(width))
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

// cmdFmtBytes formats a byte count as a human-readable string (used in direct mode).
func cmdFmtBytes(b float64) string {
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

// cmdFmtDur formats a duration as "1m30s" (used in direct mode).
func cmdFmtDur(d time.Duration) string {
	d = d.Round(time.Second)
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	if m > 0 {
		return fmt.Sprintf("%dm%ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

// getDefaultOutputDir determines the best place to save downloads.
// Tries ~/Downloads -> ~/downloads -> ~/ (home dir) -> "." (current dir)
func getDefaultOutputDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}

	paths := []string{
		filepath.Join(home, "Downloads"),
		filepath.Join(home, "downloads"),
		home,
	}

	for _, p := range paths {
		if stat, err := os.Stat(p); err == nil && stat.IsDir() {
			return p
		}
	}

	return "."
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
