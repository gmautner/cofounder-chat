# 001 - Single Workspace Chat Architecture

**Status:** Accepted

## Context

The user wants a Slack-like chat application but explicitly without multi-workspace complexity. All users share a single workspace and authenticate via Google OAuth.

## Decision

Build a single-workspace chat application with:
- Google OAuth as the sole authentication method
- Channels (public/private) and DMs as the two conversation types
- Single-level threads for focused discussions
- SSE (Server-Sent Events) for real-time updates
- PostgreSQL LISTEN/NOTIFY for database-driven event propagation

## Rationale

A single workspace dramatically simplifies the data model — no workspace_id foreign keys, no cross-workspace access control, no workspace management UI. This lets us focus on delivering a polished chat experience.

## Trade-offs

**Pros:**
- Simpler data model and API surface
- Faster development cycle
- No workspace switching complexity in the UI
- Google OAuth removes password management burden

**Cons:**
- Cannot support multiple isolated teams
- All users share the same workspace — access control is per-channel only
- No self-hosted multi-tenant option

## Alternatives Considered

- **Multi-workspace architecture:** Discarded per explicit user requirement — adds complexity without value for the use case
- **Username/password auth:** Discarded — Google OAuth is more secure and simpler for users
- **WebSockets for real-time:** Discarded in favor of SSE — simpler, works better with HTTP proxies, sufficient for chat updates
