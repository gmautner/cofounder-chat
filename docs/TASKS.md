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
| Channel CRUD (create, list, join, leave) | Done | Public/private channels, #general auto-join, browse public channels |
| DM conversations (1-on-1) | Done | Create, list with member info, dedup existing conversations |
| Message sending and display | Done | Real-time via SSE, basic Markdown rendering (bold, italic, code, links) |
| Message editing and deletion | Done | Author-only permissions, real-time updates via SSE |
| Sidebar with channels and DMs | Done | Navigation, active state, unread badges, collapsible sections |
| Frontend auth flow | Done | Login page (Google + dev), protected routes, auth context |
| Real-time SSE connection | Done | Auto-reconnect, typing indicators, message/reaction events |
| Browse channels dialog | Done | Discover and join public channels |
| New DM dialog | Done | User search, create conversations |
| Typing indicators | Done | Throttled sending, display in message area, auto-clear |

## Phase 3: Interactive Features

| Task | Status | Notes |
|------|--------|-------|
| Threads (single-level replies) | Pending | Thread panel, reply count on parent |
| Emoji reactions | Pending | Emoji picker, toggle, counts |
| @Mentions with autocomplete | Pending | Dropdown, highlight in message |
| Unread tracking and badges | Pending | Per-channel/DM badges implemented, need read position tracking UX |
| File attachments | Pending | Upload, preview, download, 15 MB limit |

## Phase 4: Polish

| Task | Status | Notes |
|------|--------|-------|
| Frontend design and UI polish | Pending | Slack-like layout done, responsive (mobile) + dark/light theme toggle pending |
| Error handling and edge cases | Pending | Network errors, reconnection, validation |
| Performance optimization | Pending | SSE connection management, query optimization |
