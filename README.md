# YVD — YouTube Video Downloader

A terminal-based video downloader with a beautiful, inline TUI. Download from YouTube and 1000+ other sites via `yt-dlp`, with smart quality/format selection or a single direct command.

⭐ Hit the repo with a star if you're enjoying YVD ⭐

## Table of Contents
- [Prerequisites](#prerequisites)
- [Installation](#installation)
  - [1. Install Go](#1-install-go)
  - [2. Install the YVD CLI](#2-install-the-yvd-cli)
- [Usage](#usage)
  - [Interactive Mode](#interactive-mode)
  - [Direct Mode](#direct-mode)
- [Upgrading](#upgrading)
- [Troubleshooting](#troubleshooting)
- [About](#about)

---

## Prerequisites

| Dependency | Required | Notes |
|---|---|---|
| **Go** | ✅ Yes | v1.21 or newer |
| **yt-dlp** | Auto-managed | Downloaded automatically on first run if not on PATH |
| **ffmpeg** | For `video + audio` only | Needed to merge separate video and audio streams into mp4/mkv. Not needed for audio-only downloads. |

**Install ffmpeg:**

```bash
# macOS (Homebrew)
brew install ffmpeg

# Ubuntu / Debian / WSL
sudo apt install ffmpeg

# Arch Linux
sudo pacman -S ffmpeg

# Windows (Scoop)
scoop install ffmpeg
```

---

## Installation

### 1. Install Go
To use the YVD CLI, you need Go **v1.21 or newer** installed on your system.

There are two main installation methods that we recommend:

**Option 1 (Linux/WSL/macOS):** The Webi installer is the simplest way for most people. Just run this in your terminal:

```bash
curl -sS https://webi.sh/golang | sh
```
Read the output of the command and follow any instructions.

**Option 2 (any platform, including Windows/PowerShell):** Use the [official Golang installation instructions](https://go.dev/dl/).

After installing Golang, open a new shell session and run `go version` to make sure everything works. If it does, move on to step 2.

**Optional troubleshooting:**

If you're getting a `command not found` error after installation, it's most likely because the directory containing the `go` program isn't in your PATH:

```bash
# For Linux/WSL
echo 'export PATH=$PATH:$HOME/.local/opt/go/bin' >> ~/.bashrc && source ~/.bashrc

# For macOS
echo 'export PATH=$PATH:$HOME/.local/opt/go/bin' >> ~/.zshrc && source ~/.zshrc
```

### 2. Install the YVD CLI
The following command will download, build, and install the `yvd` command into your Go bin directory:

```bash
go install github.com/Musharraf1128/yvd_tui/cmd/yvd@latest
```

Run `yvd --help` to confirm the installation worked.

**Optional troubleshooting:**

If `yvd` is not found after installing, add `$HOME/go/bin` to your PATH:

```bash
# For Linux/WSL
echo 'export PATH=$PATH:$HOME/go/bin' >> ~/.bashrc && source ~/.bashrc

# For macOS
echo 'export PATH=$PATH:$HOME/go/bin' >> ~/.zshrc && source ~/.zshrc
```

---

## Usage

By default, downloads are saved to your system's `~/Downloads` folder.

> **Always quote your URLs.** YouTube URLs contain `?` and `&` characters which zsh and bash treat as special characters.

```bash
# ❌ Wrong
yvd https://www.youtube.com/watch?v=dQw4w9WgXcQ

# ✅ Correct
yvd 'https://www.youtube.com/watch?v=dQw4w9WgXcQ'
```

### Interactive Mode

Launch the full interactive TUI by passing just a URL:

```bash
yvd 'https://www.youtube.com/watch?v=dQw4w9WgXcQ'
```

Use arrow keys to select your preferred quality (`1080p`, `720p`, etc.) and format:

| Format | Description | Needs ffmpeg? |
|---|---|---|
| `video + audio` | Merged mp4/mkv — best quality | ✅ Yes |
| `video only` | Raw video stream, no audio | ❌ No |
| `audio only` | Best audio stream (m4a) | ❌ No |

### Direct Mode

Skip the TUI entirely by supplying both `--quality` / `-q` and `--format` / `-f`:

```bash
# Download best video up to 1080p, merged into mp4
yvd 'https://...' -q 1080 -f mp4

# Download audio only as mp3
yvd 'https://...' -q 0 -f mp3

# Supported audio formats: mp3, m4a, aac, flac, wav, ogg, opus
```

---

## Upgrading

To upgrade to the latest version, simply re-run the install command:

```bash
go install github.com/Musharraf1128/yvd_tui/cmd/yvd@latest
```

---

## Troubleshooting

**`does not contain package github.com/Musharraf1128/yvd_tui/cmd/yvd`**

The Go module proxy may have cached an old release. Bypass it:

```bash
GOPROXY=direct go install github.com/Musharraf1128/yvd_tui/cmd/yvd@latest
```

**Upgrade not picking up the latest version**

Force a fresh install:

```bash
# Locate and remove the old binary
rm "$(which yvd)"

# Reinstall directly from source
GOPROXY=direct go install github.com/Musharraf1128/yvd_tui/cmd/yvd@latest
```

**`video + audio` download failing or no audio in output**

Make sure `ffmpeg` is installed — YVD uses it to merge separate video and audio streams into a single file. See the [Prerequisites](#prerequisites) section above.

---

## About

Built with the [Charm](https://charm.sh) ecosystem — [Bubbletea](https://github.com/charmbracelet/bubbletea), [Bubbles](https://github.com/charmbracelet/bubbles), [Lipgloss](https://github.com/charmbracelet/lipgloss) — and [go-ytdlp](https://github.com/lrstanley/go-ytdlp).

License: MIT
