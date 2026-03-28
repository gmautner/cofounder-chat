# 002 - Frontend Architecture: Single-Page Chat with SSE

**Status:** Accepted

## Context

Phase 2 required building the full chat frontend — sidebar, message list, message input, channel/DM management, and real-time updates. The backend already had all endpoints and SSE infrastructure ready.

## Decision

- **State management:** React Context for auth, local state in ChatPage for channels/conversations/messages. No external state library (Redux, Zustand, etc.).
- **Real-time:** Browser-native EventSource API with a custom `useSSE` hook that uses refs to avoid stale closures. Auto-reconnect on disconnect with 3-second backoff.
- **Routing:** React Router v7 with URL-based view selection (`/channels/:id`, `/dm/:id`). ChatPage parses the active view from `location.pathname`.
- **API layer:** Typed fetch wrapper (`lib/api.ts`) with all endpoints as named functions. Centralized error handling.
- **Theme:** Custom teal accent palette using CSS custom properties through shadcn's variable system. Deep dark sidebar with warm white content area.

## Rationale

- No external state library keeps the dependency count low and the mental model simple. Chat state naturally fits in a single page component since all views share the same channel/conversation/message/unread state.
- EventSource (SSE) is preferred over WebSocket per the tech-stack requirement. The `useRef` pattern for handlers prevents the SSE connection from being torn down on every state change.
- URL-based routing enables deep linking and browser history for channel/conversation views.

## Trade-offs

**Pros:**
- Simple mental model — one component owns all chat state
- No extra dependencies for state management
- URL-based routing enables bookmarking and sharing specific channels
- SSE auto-reconnect handles network interruptions gracefully

**Cons:**
- ChatPage component grows large as features are added (may need splitting later)
- All messages are stored in a single array (may need pagination for channels with many messages)
- No offline support or message queue (messages fail silently if SSE is disconnected)

## Alternatives Considered

- **Zustand/Jotai for state:** Discarded because the state shape is simple and localized to one page. External state would add complexity without clear benefit at this scale.
- **WebSocket instead of SSE:** Discarded per tech-stack decision — SSE avoids proxy configuration issues and the backend already has SSE infrastructure.
- **React Query for API calls:** Discarded because chat data is push-based (SSE events), not pull-based (periodic refetching). Query caching doesn't align well with real-time messaging.
