# Cofounder Chat — Development Tasks

## Phase 1: Foundation

| Task | Status | Notes |
|------|--------|-------|
| Project scaffolding (Go backend + React frontend + Dockerfile) | Done | mise.toml, go.mod, Vite scaffold, project layout |
| Database schema (users, channels, messages, reactions, attachments) | Done | 7 migrations + sqlc queries, all running correctly |
| Google OAuth authentication | Done | OAuth flow, session management, user creation |
| Dev login endpoint | Done | DEV_MODE=1 only, for testing |

## Phase 2: Core Messaging

| Task | Status | Notes |
|------|--------|-------|
| Channel CRUD (create, list, join, leave) | Pending | Public and private channels, #general auto-join |
| DM conversations (1-on-1 and group) | Pending | Create, list, participant management |
| Message sending and display | Pending | Real-time via SSE, Markdown rendering |
| Message editing and deletion | Pending | Author-only permissions |
| Sidebar with channels and DMs | Pending | Navigation, active state |

## Phase 3: Interactive Features

| Task | Status | Notes |
|------|--------|-------|
| Threads (single-level replies) | Pending | Thread panel, reply count on parent |
| Typing indicators | Pending | Real-time via SSE, auto-clear timeout |
| Emoji reactions | Pending | Emoji picker, toggle, counts |
| @Mentions with autocomplete | Pending | Dropdown, highlight in message |
| Unread tracking and badges | Pending | Per-channel/DM badges, read position tracking |
| File attachments | Pending | Upload, preview, download, 15 MB limit |

## Phase 4: Polish

| Task | Status | Notes |
|------|--------|-------|
| Frontend design and UI polish | Pending | Slack-like layout, responsive (desktop + mobile), dark/light theme |
| Error handling and edge cases | Pending | Network errors, reconnection, validation |
| Performance optimization | Pending | SSE connection management, query optimization |
