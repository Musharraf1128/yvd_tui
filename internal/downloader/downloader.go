package downloader

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	ytdlp "github.com/lrstanley/go-ytdlp"
)

// PrettifyError converts a raw go-ytdlp error into a concise, human-readable
// message. The library decorates errors as "exit code N: ...\n\n<stderr lines>",
// so this extracts the first meaningful ERROR:/WARNING: line emitted by yt-dlp.
// If none is found it falls back to the first line of the error string.
func PrettifyError(err error) string {
	if err == nil {
		return ""
	}
	raw := err.Error()
	// Walk each line looking for the yt-dlp ERROR: message.
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "ERROR:") || strings.HasPrefix(line, "WARNING:") {
			// Attach a hint for the most common network failure.
			if strings.Contains(line, "Connection refused") || strings.Contains(line, "Failed to establish") {
				return line + "\n\nHint: Check your internet connection or try a VPN/proxy."
			}
			return line
		}
	}
	// No yt-dlp error line found — return only the first line to avoid noise.
	return strings.SplitN(raw, "\n", 2)[0]
}


// EnsureInstalled makes sure yt-dlp binary is available.
// First checks if it's on PATH (pip/package manager install), then
// falls back to go-ytdlp's auto-download.
func EnsureInstalled(ctx context.Context) error {
	if _, err := exec.LookPath("yt-dlp"); err == nil {
		return nil // already installed system-wide
	}
	_, err := ytdlp.Install(ctx, nil)
	return err
}

// FetchInfo retrieves video metadata without downloading.
func FetchInfo(ctx context.Context, url string) (*VideoInfo, error) {
	dl := ytdlp.New().
		SkipDownload().
		DumpJSON().
		NoWarnings().
		Quiet()

	result, err := dl.Run(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch video info: %w", err)
	}

	extracted, err := result.GetExtractedInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to parse video info: %w", err)
	}

	if len(extracted) == 0 {
		return nil, fmt.Errorf("no video info returned")
	}

	ei := extracted[0]

	info := &VideoInfo{
		Title:    derefStr(ei.Title),
		URL:      derefStr(ei.WebpageURL),
		Duration: formatDuration(derefFloat(ei.Duration)),
		Uploader: derefStr(ei.Uploader),
	}

	if ei.Description != nil {
		info.Description = truncate(*ei.Description, 200)
	}
	if ei.Thumbnail != nil {
		info.Thumbnail = *ei.Thumbnail
	}

	// Parse formats, de-duplicating by (resolution, ext, note).
	seen := make(map[string]bool)
	for _, f := range ei.Formats {
		hasVideo := derefStr(f.VCodec) != "" && derefStr(f.VCodec) != "none"
		hasAudio := derefStr(f.ACodec) != "" && derefStr(f.ACodec) != "none"

		resolution := derefStr(f.Resolution)
		if resolution == "" && f.Height != nil {
			resolution = fmt.Sprintf("%dp", int(*f.Height))
		}

		ext := derefStr(f.Extension)
		note := derefStr(f.FormatNote)
		formatID := derefStr(f.FormatID)

		key := fmt.Sprintf("%s-%s-%s", resolution, ext, note)
		if seen[key] {
			continue
		}
		seen[key] = true

		var filesize int64
		if f.FileSize != nil {
			filesize = int64(*f.FileSize)
		} else if f.FileSizeApprox != nil {
			filesize = int64(*f.FileSizeApprox)
		}

		info.Formats = append(info.Formats, Format{
			FormatID:   formatID,
			Extension:  ext,
			Resolution: resolution,
			Filesize:   filesize,
			Note:       note,
			HasVideo:   hasVideo,
			HasAudio:   hasAudio,
			Codec:      derefStr(f.VCodec),
			AudioCodec: derefStr(f.ACodec),
		})
	}

	return info, nil
}

// Download starts a video download, calling onProgress with real-time updates.
// onProgress may be nil (no callbacks will be made).
func Download(ctx context.Context, opts DownloadOptions, onProgress func(ProgressInfo)) error {
	dl := ytdlp.New().NoWarnings()

	// Format selector.
	if fs := buildFormatString(opts); fs != "" {
		dl = dl.Format(fs)
	}

	// Request a specific container when merging video+audio (needs ffmpeg).
	if !opts.AudioOnly && !opts.VideoOnly && opts.OutputFormat != "" {
		dl = dl.MergeOutputFormat(opts.OutputFormat)
	}

	// Output filename template.
	tmpl := "%(title)s.%(ext)s"
	if opts.OutputDir != "" {
		tmpl = opts.OutputDir + "/" + tmpl
	}
	dl = dl.Output(tmpl)

	// Wire up real-time progress callback.
	dl.ProgressFunc(250*time.Millisecond, func(u ytdlp.ProgressUpdate) {
		if onProgress == nil {
			return
		}
		var speed float64
		if dur := u.Duration().Seconds(); dur > 0 {
			speed = float64(u.DownloadedBytes) / dur
		}
		onProgress(ProgressInfo{
			Status:          string(u.Status),
			Percent:         u.Percent(),
			DownloadedBytes: u.DownloadedBytes,
			TotalBytes:      u.TotalBytes,
			ETA:             u.ETA(),
			Speed:           speed,
			Filename:        u.Filename,
		})
	})

	_, err := dl.Run(ctx, opts.URL)
	return err
}

// buildFormatString converts DownloadOptions into a yt-dlp format selector string.
func buildFormatString(opts DownloadOptions) string {
	if opts.AudioOnly {
		// Best audio-only stream; prefer m4a for broad compatibility.
		return "bestaudio[ext=m4a]/bestaudio"
	}

	heightFilter := ""
	if opts.Quality != "" {
		heightFilter = fmt.Sprintf("[height<=%s]", opts.Quality)
	}

	if opts.VideoOnly {
		return fmt.Sprintf("bestvideo%s", heightFilter)
	}

	// Video + audio merged (requires ffmpeg).
	return fmt.Sprintf("bestvideo%s+bestaudio/best%s", heightFilter, heightFilter)
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefFloat(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}

func formatDuration(secs float64) string {
	total := int(secs)
	h := total / 3600
	m := (total % 3600) / 60
	s := total % 60

	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	}
	return fmt.Sprintf("%dm %ds", m, s)
}

func truncate(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
