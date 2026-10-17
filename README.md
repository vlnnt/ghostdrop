# GhostDrop

**Language:** English | [Русский](README.ru.md)

A simple Discord bot for anonymously sharing images, GIFs, videos, and audio files.

## Features

- Anonymous media sharing with `/anon`.
- Supports images, GIFs, videos, and audio.
- Choose any text channel.
- Simple web-based configuration.
- Windows and Linux support.

## Installation

1. Download the latest version from [Releases](../../releases).
2. Create a bot in the [Discord Developer Portal](https://discord.com/developers/applications).
3. Invite the bot to your server with `bot` and `applications.commands` scopes.
4. Run `ghostdrop.exe` on Windows or `./ghostdrop` on Linux.
5. Open `http://127.0.0.1:8080` and enter your bot token.
6. Click **Save & Start GhostDrop**.

## Usage

```text
/anon media:<file> channel:<channel>
```

**Example:**

```text
/anon media:funny.gif channel:#general
```

The bot sends the file without publicly revealing the sender.

**Supported:** PNG, JPG, GIF, WebP, MP4, WebM, MOV, MP3, WAV, OGG, FLAC and M4A.

**Maximum file size:** 20 MiB.

## Configuration

Settings are saved automatically in `config.yaml` next to the executable.

Leave Guild ID empty to enable global commands.

Keep your bot token and configuration file private.