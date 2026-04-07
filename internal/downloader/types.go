package downloader

import "time"

// VideoInfo holds metadata about a video.
type VideoInfo struct {
	Title       string
	URL         string
	Duration    string
	Uploader    string
	Description string
	Thumbnail   string
	Formats     []Format
}

// Format represents a single available download format.
type Format struct {
	FormatID   string
	Extension  string
	Resolution string
	Filesize   int64
	Note       string
	HasVideo   bool
	HasAudio   bool
	Codec      string
	AudioCodec string
}

// DownloadOptions configures a download.
type DownloadOptions struct {
	URL          string
	Quality      string // e.g. "1080", "720", "480"
	AudioOnly    bool
	VideoOnly    bool
	OutputDir    string
	OutputFormat string // container format for merge, e.g. "mp4", "webm"
}

// ProgressInfo is a point-in-time snapshot of download progress.
type ProgressInfo struct {
	Status          string        // "starting" | "downloading" | "post_processing" | "finished" | "error"
	Percent         float64       // 0–100
	DownloadedBytes int
	TotalBytes      int
	ETA             time.Duration
	Speed           float64 // bytes per second
	Filename        string
}
