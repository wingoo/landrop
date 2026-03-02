# landrop

`landrop` is a minimal LAN receiver for file and text drop.

Run it on a macOS/Windows machine, then upload from any device with browser or `curl`.
No iCloud/account required.

## Features

- Receive files from browser (`multipart/form-data`)
- Receive text from browser or `curl`
- Optional text stdout mode (`--text-stdout`): print received text in receiver CLI without saving file
- Optional server-side clipboard copy on macOS/Windows (`--clipboard`)
- Text endpoint has a separate safety limit (default 5MB)
- One-time token enabled by default (4-digit numeric, `?t=<token>`)
- `--once`: exit after first successful receive
- Startup QR code for quick phone access

## Install

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/wingoo/landrop/main/install/install.sh | bash
```

Optional:

```bash
LANDROP_REPO=wingoo/landrop VERSION=v0.1.0 curl -fsSL https://raw.githubusercontent.com/wingoo/landrop/main/install/install.sh | bash
```

### Windows (PowerShell)

```powershell
iwr -useb https://raw.githubusercontent.com/wingoo/landrop/main/install/install.ps1 | iex
```

### Build / install with Go directly

```bash
# install latest to $GOBIN (or $HOME/go/bin)
go install github.com/wingoo/landrop@latest

# or build from local source
git clone https://github.com/wingoo/landrop.git
cd landrop
go build -o landrop ./cmd/landrop
```

## Quick Start

```bash
landrop
```

Startup output includes:

- save directory
- listening address
- access URL with token
- endpoints
- QR code (scan with phone)

Default save dir:

- macOS/Linux: `~/Downloads/landrop`
- Windows: `%USERPROFILE%\Downloads\landrop`

## Usage

### Browser

1. Open printed `Access` URL in phone/another device browser.
2. Upload files or submit text.
3. Files/text are saved into save dir.

Use `landrop --text-stdout` if you want text to be printed directly in receiver CLI (no text file saved).

### curl upload file

```bash
curl -F "file=@/path/to/a.png" "http://<ip>:7777/upload?t=<token>"
```

Multiple files:

```bash
curl -F "file=@/path/to/a.png" -F "file=@/path/to/b.pdf" "http://<ip>:7777/upload?t=<token>"
```

### curl send text

```bash
curl -H "Content-Type: text/plain" --data-binary "hello from curl" "http://<ip>:7777/text?t=<token>"
```

## Flags

```text
--port <int>       listen port (default 7777)
--dir <path>       save directory
--once             exit after first successful receive
--text-stdout      print received text to CLI and do not save text file
--clipboard        enable clipboard copy on supported OS (macOS/Windows)
--no-token         disable token check (not recommended)
--no-qr            disable startup QR printing
```

## Security Notes

- LAN-only tool by design.
- Token is enabled by default.
- Do not expose this service to public internet.
- By default, server logs file/text receive metadata, but never prints text content.
- If `--text-stdout` is enabled, received text content will be printed to receiver CLI.

## API

- `GET /` web UI
- `POST /upload?t=...` multipart field: `file` (multiple supported)
- `POST /text?t=...` content-type:
  - `text/plain`
  - `application/x-www-form-urlencoded` with field `text`

Responses are JSON.

## Release Artifacts

- `landrop_darwin_arm64.tar.gz`
- `landrop_darwin_amd64.tar.gz`
- `landrop_linux_amd64.tar.gz`
- `landrop_linux_arm64.tar.gz`
- `landrop_windows_amd64.zip`
- `checksums.txt`
