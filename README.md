# YVD — YouTube Video Downloader

This is the command line tool for downloading videos directly from YouTube and other sites via yt-dlp. It features a beautiful, inline terminal UI and smart format selection.

⭐ Hit the repo with a star if you're enjoying YVD ⭐

## Table of Contents
- [Installation](#installation)
  - [1. Install Go](#1-install-go)
  - [2. Install the YVD CLI](#2-install-the-yvd-cli)
- [Usage](#usage)
  - [Interactive Mode](#interactive-mode)
  - [Direct Mode](#direct-mode)
- [Troubleshooting Upgrading](#troubleshooting-upgrading)
- [About](#about)

---

## Installation

### 1. Install Go
To use the YVD CLI, you need an up-to-date Golang toolchain installed on your system.

There are two main installation methods that we recommend:

**Option 1 (Linux/WSL/macOS):** The Webi installer is the simplest way for most people. Just run this in your terminal:

```bash
curl -sS https://webi.sh/golang | sh
```
Read the output of the command and follow any instructions.

**Option 2 (any platform, including Windows/PowerShell):** Use the [official Golang installation instructions](https://go.dev/dl/). 

After installing Golang, open a new shell session and run `go version` to make sure everything works. If it does, move on to step 2.

**Optional troubleshooting:**

If you already had a version of Go installed a different way, on Linux/macOS you can run `which go` to find out where it's installed, and (if needed) remove the old version manually.

If you're getting a "command not found" error after installation, it's most likely because the directory containing the `go` program isn't in your PATH. You need to add the directory to your PATH by modifying your shell's configuration file:

```bash
# For Linux/WSL
echo 'export PATH=$PATH:$HOME/.local/opt/go/bin' >> ~/.bashrc
# Next, reload your shell configuration
source ~/.bashrc

# For macOS
echo 'export PATH=$PATH:$HOME/.local/opt/go/bin' >> ~/.zshrc
# Next, reload your shell configuration
source ~/.zshrc
```

### 2. Install the YVD CLI
The following command will download, build, and install the `yvd` command into your Go toolchain's bin directory. Go ahead and run it:

```bash
go install github.com/Musharraf1128/yvd_tui/cmd/yvd@latest
```
Run `yvd --help` on your command line to make sure the installation worked. If it did, you're ready to go!

**Optional troubleshooting:**

If you're getting a "command not found" error for `yvd help`, it's most likely because the directory containing the program isn't in your PATH. You probably need to add `$HOME/go/bin` to your PATH:

```bash
# For Linux/WSL
echo 'export PATH=$PATH:$HOME/go/bin' >> ~/.bashrc
# Next, reload your shell configuration
source ~/.bashrc

# For macOS
echo 'export PATH=$PATH:$HOME/go/bin' >> ~/.zshrc
# Next, reload your shell configuration
source ~/.zshrc
```

---

## Usage

By default, downloads will be saved directly to your system's `~/Downloads` folder.

**Always quote your URLs.** YouTube URLs contain `?` and `&` characters, which zsh and bash treat as glob operators.

```bash
# Wrong
yvd https://www.youtube.com/watch?v=dQw4w9WgXcQ

# Correct
yvd 'https://www.youtube.com/watch?v=dQw4w9WgXcQ'
```

### Interactive mode

Launch the full interactive TUI:

```bash
yvd 'https://www.youtube.com/watch?v=dQw4w9WgXcQ'
```

Use your arrow keys to select your preferred quality (`1080p`, `720p`, etc) and format (`video + audio`, `audio only`).

### Direct Mode

You can bypass the TUI entirely if you know what you want to download. Supply both `--quality` / `-q` and `--format` / `-f`:

```bash
# Download best video up to 1080p, merged into an mp4 container
yvd 'https://...' -q 1080 -f mp4

# Download audio only as mp3
yvd 'https://...' -q 0 -f mp3
```

---

### Upgrading

If you just installed the CLI, it's already upgraded!

To upgrade to the latest version in the future, simply re-run:

```bash
go install github.com/Musharraf1128/yvd_tui/cmd/yvd@latest
```

### Troubleshooting Upgrading

**1. Bypass the proxy**

If you keep getting the same upgrade message, you may be pulling from an old cache.

```bash
GOPROXY=direct go install github.com/Musharraf1128/yvd_tui/cmd/yvd@latest
```

**2. Reinstall**

If that doesn't work, try a fresh install:

Locate the binary file:

```bash
which yvd
```

Carefully remove the binary file after confirming the path is correct:

```bash
rm "$(which yvd)"
```

Clean install: Repeat the installation step.

---

## About

A beautiful, inline terminal UI tool for downloading videos natively via the Charm ecosystem (Bubbletea / Lipgloss) and `go-ytdlp`.

License: MIT
