# 🗃️ tgcli — Telegram CLI: sync, search, send

Telegram CLI built on top of `gotd`, focused on:

- Best-effort local sync of message history + continuous capture
- Fast offline search
- Sending messages (text, files, reactions)
- Chat, group, and channel management
- **E2E testing for OpenClaw**

This is a third-party tool that uses the Telegram MTProto API via gotd and is not affiliated with Telegram.

## Status

🚧 **In Development** — Core implementation in progress.

See [docs/spec.md](docs/spec.md) for the full design specification.

## Install / Build

### Build locally

```bash
go build -tags sqlite_fts5 -o ./dist/tgcli ./cmd/tgcli
```

Run (local build):

```bash
./dist/tgcli --help
```

## Quick start

Default store directory is `~/.tgcli` (override with `--store DIR`).

```bash
# 1) Authenticate (prompts for phone + code), then bootstrap sync
./dist/tgcli auth

# 2) Keep syncing (never prompts; requires prior auth)
./dist/tgcli sync --follow

# Diagnostics
./dist/tgcli doctor

# Search messages
./dist/tgcli messages search "meeting"

# Send a message
./dist/tgcli send text --to 123456789 --message "hello"

# Send a file
./dist/tgcli send file --to 123456789 --file ./pic.jpg --caption "check this"

# List chats
./dist/tgcli chats list

# List messages in a chat
./dist/tgcli messages list --chat 123456789
```

## High-level UX

- `tgcli auth`: interactive login (phone + code), then immediately performs initial data sync
- `tgcli sync`: non-interactive sync loop (never prompts; errors if not authenticated)
- Output is human-readable by default; pass `--json` for machine-readable output

## Storage

Defaults to `~/.tgcli` (override with `--store DIR`).

```
~/.tgcli/
├── session.json    # Telegram session data
├── tgcli.db        # Messages, chats, FTS index
├── media/          # Downloaded media files
└── LOCK            # Single-instance lock
```

## Environment overrides

- `TGCLI_PHONE`: pre-fill phone number for auth (for testing)
- `TGCLI_APP_ID`: override Telegram API app ID
- `TGCLI_APP_HASH`: override Telegram API app hash

## E2E Testing

Designed for OpenClaw integration testing:

```bash
# Authenticate once
tgcli auth

# Send test message
tgcli send text --to 123456789 --message "E2E test" --json

# Verify receipt
tgcli messages list --chat 123456789 --json | jq '.messages[-1]'

# Search
tgcli messages search "E2E test" --json
```

All commands support `--json` for easy parsing in test scripts.

## Prior Art / Credit

This project is heavily inspired by the excellent `wacli` by Peter Steinberger:
- [wacli](https://github.com/steipete/wacli)

## Development

```bash
# Run tests
go test -tags sqlite_fts5 ./...

# Build
go build -tags sqlite_fts5 -o ./dist/tgcli ./cmd/tgcli

# Run
./dist/tgcli --help
```

## License

See [LICENSE](LICENSE).
