# tgcli specification

This document defines `tgcli`: a Telegram CLI that receives messages, supports search, sending, and chat management. Built for **OpenClaw E2E testing**.

## Implementation Note

**Using Bot API** (not MTProto/gotd) for simplicity and reliability:
- Lightweight, fast builds
- No phone authentication required
- Perfect for automated testing
- Get a token from @BotFather

## Goals

- **Simple authentication**: `tgcli auth` validates bot token
- **Receive messages**: Listen for messages sent to the bot
- **Send messages**: Text, files, photos with reply support
- **Edit/delete messages**: Modify sent messages
- **Fast offline search**: Local SQLite database
- **Human-first output**: Readable tables by default, `--json` for scripting
- **Single-instance safety**: Store locking to avoid conflicts
- **E2E testing ready**: Designed for OpenClaw integration testing

## Non-goals

- Secret chats (not supported by Bot API)
- Phone-based authentication (using bot tokens instead)
- Full message history (Bot API only receives messages TO the bot)
- Reactions (requires Bot API 7.0+)

## Storage layout

Default store: `~/.tgcli` (override with `--store DIR`).

Files:
- `~/.tgcli/tgcli.db` — SQLite DB (messages, chats, metadata)
- `~/.tgcli/LOCK` — Store lock to prevent concurrent access

Permissions: Directory `0700`, files `0600` (owner only).

## Authentication

Uses Telegram Bot API tokens (from @BotFather):

```bash
export TGCLI_BOT_TOKEN="123456:ABC..."
tgcli auth   # Validates token, shows bot info
tgcli doctor # Health check
```

## Commands

### Core
| Command | Description |
|---------|-------------|
| `auth` | Validate bot token, show bot info |
| `doctor` | Health check (token, store, connectivity) |
| `version` | Show version |

### Messages
| Command | Description |
|---------|-------------|
| `sync [--follow]` | Receive messages (continuous with --follow) |
| `send text --to ID --message MSG` | Send text message |
| `send file --to ID --file PATH` | Send file/document |
| `send edit --chat ID --message-id ID --text MSG` | Edit message |
| `send delete --chat ID --message-id ID` | Delete message |
| `send forward --to ID --from ID --message-id ID` | Forward message |
| `messages list --chat ID` | List messages in chat |
| `messages search QUERY` | Search messages |

### Chats
| Command | Description |
|---------|-------------|
| `chats list` | List stored chats |
| `chats info --chat ID` | Get chat details |
| `groups list` | List stored groups |
| `channels list` | List stored channels |

All commands support `--json` for machine-readable output.

## E2E Testing Flow

Designed for OpenClaw CI/CD:

```bash
# Setup
export TGCLI_BOT_TOKEN="$TEST_BOT_TOKEN"
tgcli doctor

# Send test message
RESULT=$(tgcli send text --to "$CHAT_ID" --message "test" --json)
MSG_ID=$(echo "$RESULT" | jq -r '.message_id')

# Verify
tgcli messages list --chat "$CHAT_ID" --json | jq '.[-1]'

# Cleanup
tgcli send delete --chat "$CHAT_ID" --message-id "$MSG_ID"
```

## Implementation Status

### Phase 1: Core (MVP) ✅
- [x] Project structure
- [x] Auth (bot token validation)
- [x] Basic sync (receive messages)
- [x] Send text messages
- [x] List chats
- [x] List messages
- [x] JSON output

### Phase 2: Extended ✅
- [x] File upload
- [x] Replies (--reply-to)
- [x] Message search
- [x] Edit messages
- [x] Delete messages
- [x] Forward messages
- [x] Chat info
- [ ] Reactions (Bot API 7.0+ needed)
- [ ] File download

### Phase 3: Advanced ✅
- [x] Continuous sync (--follow)
- [x] Security hardening (permissions, input validation)
- [ ] History backfill (Bot API limitation)
- [ ] Media download

## Security

- Token from environment variable (never logged)
- Parameterized SQL queries
- File permissions: `0700` for dir, `0600` for files
- LIKE pattern escaping (prevents injection)
- Input validation (max lengths, file sizes)
- Result limits (max 1000)

## Technology

- **Language**: Go
- **Telegram library**: `telegram-bot-api` v5.5
- **CLI framework**: `cobra`
- **Database**: SQLite (pure Go via modernc.org/sqlite)

## License

MIT
