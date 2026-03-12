# tgcli specification

This document defines the v1 plan for `tgcli`: a Telegram CLI that syncs messages locally, supports fast search, sending, and chat/group/channel management. Implementation will use `gotd` (pure Go Telegram MTProto client) under the hood.

## Goals

- **Explicit authentication step**: `tgcli auth` prompts for phone number, sends code, completes login
- **Auth starts syncing immediately**: after successful authentication, `tgcli auth` begins initial sync (dialogs + recent messages)
- **Non-interactive sync**: `tgcli sync` never prompts for auth; fails with clear error if not authenticated
- **Fast offline message search**: local SQLite + FTS5 index
- **Human-first output**: readable tables by default, `--json` opt-in for scripting/testing
- **Single-instance safety**: store locking to avoid multi-instance session conflicts
- **Group & channel management**: list, inspect, send, manage participants
- **E2E testing ready**: designed for OpenClaw integration testing

## Non-goals (v1)

- Secret chats (not supported by bot/user API)
- Complex media processing (initial version: basic file send/receive)
- Full message formatting parity (markdown/entities will improve over time)

## Terminology

- **Chat ID**: Telegram chat identifier (user ID, group/channel ID, etc.)
- **Store directory**: directory containing all local state, default `~/.tgcli`
- **Dialog**: Telegram term for a chat entry (DM, group, channel, etc.)

## Storage layout

Default store: `~/.tgcli` (override with `--store DIR`).

Proposed files:

- `~/.tgcli/session.json` — Telegram session data (auth key, user info)
- `~/.tgcli/tgcli.db` — SQLite DB (messages, chats, FTS, local metadata)
- `~/.tgcli/media/...` — downloaded media (on-demand or background)
- `~/.tgcli/LOCK` — store lock to prevent concurrent access

## Concurrency + locking

Every command that accesses the Telegram session must acquire an exclusive lock in the store dir.

Behavior:
- If lock is held: fail fast with clear message (include PID if available)
- Prevents running multiple `tgcli` instances against same session

## Authentication model

### Commands

- `tgcli auth` (interactive)
  - Prompts for phone number
  - Sends authentication code via Telegram
  - User enters code
  - After success: starts initial sync (bootstrap) immediately
  - Exits after initial sync completes, unless `--follow` is set

- `tgcli sync` (non-interactive)
  - Requires existing authenticated session in `session.json`
  - Never prompts for auth; if not authenticated, prints "run `tgcli auth`"
  - `--once` performs bounded sync and exits
  - Default (or `--follow`) stays connected and continues capturing messages

### UX principle

Only `tgcli auth` is expected to prompt for credentials. `tgcli sync` should be safe to run in scripts/daemons without surprising interactivity.

## Sync model

`tgcli` captures messages via gotd event handlers:

- Initial dialogs fetch: get all chats/groups/channels
- Message history: fetch recent messages for each dialog
- Real-time updates: new incoming/outgoing messages while connected
- Connection lifecycle: handle reconnects with backoff

### Bootstrap sync (after auth)

Immediately after authentication success, `tgcli auth` runs bootstrap sync:

- Fetches all dialogs (chats, groups, channels)
- Fetches recent message history for each dialog (configurable limit)
- Updates chat names, participant counts, metadata
- Exits once initial sync completes, unless `--follow`

### Continuous sync

`tgcli sync --follow` keeps running:
- Persists new messages as they arrive
- Handles safe reconnect with backoff on disconnect
- Maintains real-time message capture

## Database schema (tgcli.db)

### Tables (proposed)

- `chats`
  - `id` (PK), `type` (`user|group|channel|supergroup`), `title`, `username`, `last_message_id`, `last_message_ts`, `unread_count`, ...
  
- `users`
  - `id` (PK), `first_name`, `last_name`, `username`, `phone`, `is_bot`, ...
  
- `messages`
  - `id` (PK), `chat_id`, `from_user_id`, `date`, `text`, `reply_to_message_id`, `media_type`, `media_path`, ...
  
- `messages_fts`
  - FTS5 virtual table over `messages.text` for fast search

### Indexes

- `messages(chat_id, date DESC)` — chat message list queries
- `messages(from_user_id, date DESC)` — user message history
- `chats(last_message_ts DESC)` — recent chats list

## Commands (MVP)

### Core

```bash
# Authentication
tgcli auth                          # Phone + code auth
tgcli doctor                        # Diagnostics (session status, DB stats)

# Sync
tgcli sync                          # One-time sync
tgcli sync --follow                 # Continuous sync
```

### Messages

```bash
# List/search
tgcli messages list --chat <id>           # List messages in chat
tgcli messages search "query"             # FTS search across all chats
tgcli messages search --chat <id> "query" # Search within specific chat

# Send
tgcli send text --to <id> --message "hello"
tgcli send text --to <id> --message "hi" --reply-to <msg_id>
tgcli send file --to <id> --file pic.jpg --caption "check this"
tgcli send reaction --chat <id> --message-id <id> --emoji "👍"
```

### Chats

```bash
tgcli chats list                    # List all chats (DMs, groups, channels)
tgcli chats info --chat <id>        # Chat details (title, members, etc.)
tgcli chats unread                  # List chats with unread messages
```

### Groups

```bash
tgcli groups list                   # List groups
tgcli groups info --chat <id>       # Group details
tgcli groups members --chat <id>    # List group members
```

### Channels

```bash
tgcli channels list                 # List channels
tgcli channels info --chat <id>     # Channel details
```

### Media

```bash
tgcli media download --chat <id> --message-id <id>  # Download media from message
```

### All commands support `--json` flag for machine-readable output

## Output formats

### Human-readable (default)

```
CHATS
ID          Type    Title              Last Message
123456789   user    John Doe          2 hours ago
-100987654  group   Dev Team          5 minutes ago
```

### JSON (with `--json` flag)

```json
{
  "chats": [
    {
      "id": 123456789,
      "type": "user",
      "title": "John Doe",
      "last_message_ts": 1234567890
    }
  ]
}
```

## E2E Testing Flow

Designed to support OpenClaw E2E tests:

1. **Setup**: `tgcli auth` with test account
2. **Send**: `tgcli send text --to <test_chat> --message "test" --json`
3. **Verify**: `tgcli messages list --chat <test_chat> --json | jq '.messages[-1].text'`
4. **Cleanup**: Database queries or message deletion

## Implementation phases

### Phase 1: Core (MVP)
- ✅ Project structure
- ✅ Auth (phone + code)
- ✅ Basic sync (dialogs + messages)
- ✅ Send text messages
- ✅ List chats
- ✅ List messages
- ✅ JSON output

### Phase 2: Extended
- ✅ File upload/download
- ✅ Reactions
- ✅ Replies
- ✅ Message search (FTS)
- ✅ Groups/channels details

### Phase 3: Advanced
- ✅ Continuous sync (--follow)
- ✅ History backfill
- ✅ Advanced media handling
- ✅ Message editing/deletion

## Technology choices

- **Language**: Go (consistency with wacli, great tooling, easy cross-platform builds)
- **Telegram library**: `gotd` (pure Go, no CGo, modern MTProto implementation)
- **CLI framework**: `cobra` (same as wacli, battle-tested)
- **Database**: SQLite with FTS5 (fast, embedded, proven)
- **Output**: `text/tabwriter` for human-readable, `encoding/json` for machine-readable

## Development principles

- **Clean code**: Easy to read and understand
- **Clear errors**: Actionable error messages
- **Testable**: Unit tests for core logic, E2E tests for integration
- **Well-documented**: Comments for non-obvious logic
- **OpenClaw-aligned**: Cover all Telegram features that OpenClaw supports
