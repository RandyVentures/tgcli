# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

## [0.1.0] - 2026-03-12

### Added
- Initial release
- Bot API authentication (`tgcli auth`)
- Health check (`tgcli doctor`)
- Send text messages (`tgcli send text`)
- Send files/photos (`tgcli send file`)
- Edit messages (`tgcli send edit`)
- Delete messages (`tgcli send delete`)
- Forward messages (`tgcli send forward`)
- Receive messages (`tgcli sync --follow`)
- List chats (`tgcli chats list`)
- Chat info (`tgcli chats info`)
- List messages (`tgcli messages list`)
- Search messages (`tgcli messages search`)
- List groups (`tgcli groups list`)
- List channels (`tgcli channels list`)
- JSON output support (`--json` flag)
- SQLite local storage
- Security hardening (file permissions, input validation)
- E2E testing support for OpenClaw

### Security
- Store directory permissions: 0700
- Database file permissions: 0600
- LIKE pattern escaping to prevent injection
- Input validation (message length, file size)
- Result limits to prevent abuse
