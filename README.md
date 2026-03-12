# 🗃️ tgcli — Telegram CLI: sync, search, send.

Telegram CLI built on top of `telegram-bot-api`, focused on:

- Receiving and storing messages locally
- Fast offline search
- Sending messages (text, files, photos)
- JSON output for E2E test automation

**Built for [OpenClaw](https://github.com/anthropics/openclaw) E2E testing** — provides a scriptable CLI for automated Telegram channel testing.

This is a third-party tool that uses the Telegram Bot API and is not affiliated with Telegram.

## Status

Core implementation complete. Ready for integration testing.

See `docs/spec.md` for the full design specification.

## Install / Build

### Build locally

```bash
go build -o ./dist/tgcli ./cmd/tgcli
```

Run:

```bash
./dist/tgcli --help
```

## Quick start

Default store directory is `~/.tgcli` (override with `--store DIR`).

```bash
# 1) Set your bot token (get one from @BotFather on Telegram)
export TGCLI_BOT_TOKEN="your_bot_token_here"

# 2) Validate token and check connectivity
./dist/tgcli doctor
./dist/tgcli auth

# 3) Listen for incoming messages
./dist/tgcli sync --follow

# 4) Send a message
./dist/tgcli send text --to 123456789 --message "Hello from tgcli!"

# 5) Send a file
./dist/tgcli send file --to 123456789 --file ./pic.jpg --caption "Check this out"

# 6) List stored chats
./dist/tgcli chats list

# 7) Search messages
./dist/tgcli messages search "meeting"
```

## E2E Testing with OpenClaw

`tgcli` is designed for automated E2E testing of Telegram integrations. All commands support `--json` for machine-readable output.

### Example: CI/CD Test Script

```bash
#!/bin/bash
set -e

export TGCLI_BOT_TOKEN="$TEST_BOT_TOKEN"
TEST_CHAT_ID="$TELEGRAM_TEST_CHAT_ID"

# Send test message
RESULT=$(./dist/tgcli send text --to "$TEST_CHAT_ID" --message "E2E test $(date)" --json)
MESSAGE_ID=$(echo "$RESULT" | jq -r '.message_id')
echo "Sent message ID: $MESSAGE_ID"

# Wait for response and sync
sleep 5
./dist/tgcli sync

# Verify message was stored
./dist/tgcli messages list --chat "$TEST_CHAT_ID" --json | jq '.[-1]'

# Search for expected content
FOUND=$(./dist/tgcli messages search "E2E test" --json | jq 'length')
if [ "$FOUND" -gt 0 ]; then
  echo "✅ E2E test passed"
else
  echo "❌ E2E test failed"
  exit 1
fi
```

### GitHub Actions Example

```yaml
name: Telegram E2E Tests

on: [push, pull_request]

jobs:
  telegram-e2e:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      
      - name: Build tgcli
        run: go build -o ./dist/tgcli ./cmd/tgcli
      
      - name: Run E2E Tests
        env:
          TGCLI_BOT_TOKEN: ${{ secrets.TELEGRAM_BOT_TOKEN }}
          TEST_CHAT_ID: ${{ secrets.TELEGRAM_TEST_CHAT_ID }}
        run: |
          ./dist/tgcli doctor
          ./dist/tgcli send text --to "$TEST_CHAT_ID" --message "CI test" --json
```

## High-level UX

- `tgcli auth`: validates bot token and displays bot info
- `tgcli doctor`: checks token, store, and connectivity
- `tgcli sync`: receives messages (use `--follow` for continuous)
- `tgcli send`: sends text or files to a chat
- Output is human-readable by default; pass `--json` for machine-readable output

## Storage

Defaults to `~/.tgcli` (override with `--store DIR`).

```
~/.tgcli/
├── tgcli.db     # SQLite database (chats, messages)
└── media/       # Downloaded media files (future)
```

## Environment Variables

- `TGCLI_BOT_TOKEN`: **Required.** Your Telegram bot token from @BotFather.

## Commands Reference

| Command | Description |
|---------|-------------|
| `auth` | Validate bot token and show bot info |
| `doctor` | Check configuration and connectivity |
| `sync [--follow]` | Receive messages (continuous with --follow) |
| `send text --to ID --message MSG` | Send text message |
| `send file --to ID --file PATH` | Send file/document |
| `send edit --chat ID --message-id ID --text MSG` | Edit a message |
| `send delete --chat ID --message-id ID` | Delete a message |
| `send forward --to ID --from ID --message-id ID` | Forward a message |
| `chats list` | List stored chats |
| `chats info --chat ID` | Get detailed chat info |
| `messages list --chat ID` | List messages in a chat |
| `messages search QUERY` | Search messages |
| `groups list` | List stored groups |
| `channels list` | List stored channels |

All commands support `--json` for machine-readable output.

## Bot API vs MTProto

This CLI uses the **Telegram Bot API** (not MTProto) for simplicity and reliability:

| Aspect | Bot API (tgcli) | MTProto (gotd) |
|--------|-----------------|----------------|
| Auth | Bot token from @BotFather | Phone + code |
| Build | Fast, lightweight | Slow, 4GB+ RAM |
| Messages | Only messages TO the bot | All messages |
| Use case | Bots, automation, testing | Full client |

For E2E testing, Bot API is ideal — you control the test bot and can verify message flow.

## Prior Art / Credit

This project is inspired by the excellent `wacli` by Peter Steinberger:

- [`wacli`](https://github.com/steipete/wacli) — WhatsApp CLI

## License

See `LICENSE`.
