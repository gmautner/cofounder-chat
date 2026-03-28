# Cofounder Chat — Product Requirements Document

## Overview

Cofounder Chat is a real-time team messaging application inspired by Slack. It provides a single shared workspace where all users sign in with their Google account. The app supports channels (public and private), direct messages (individual and group), threaded conversations, emoji reactions, file attachments, typing indicators, and unread badges — everything a team needs for day-to-day communication, without the complexity of multi-workspace management.

## Target Users

- Small to medium teams (5–200 people) who need a centralized place to communicate
- Organizations that use Google Workspace and want frictionless sign-in
- Teams that want a simple, self-hosted alternative to Slack without multi-workspace overhead

## Core Features

### 1. Authentication

- **Google OAuth sign-in only** — no username/password registration
- User profile automatically populated from Google account (name, email, avatar)
- Persistent sessions (stay logged in across browser restarts)
- Logout functionality

### 2. Channels

- **Public channels** — visible and joinable by any authenticated user
- **Private channels** — visible only to members; join by invitation
- Channel creation with name and optional description
- Channel listing in the sidebar with unread badges
- Channel member list
- Ability to leave a channel
- A default `#general` channel that all users automatically join on first sign-in

### 3. Direct Messages (DMs)

- **1-on-1 DMs** — private conversation between two users
- **Group DMs** — private conversation among 3+ users
- DM listing in the sidebar with unread badges
- User search to start new DM conversations
- Group DM creation by selecting multiple users

### 4. Messaging

- Real-time message delivery (messages appear instantly for all participants)
- Rich text input with Markdown support (bold, italic, code, links)
- Message display with sender avatar, name, and timestamp
- Message editing (by the author)
- Message deletion (by the author)

### 5. Threads

- Reply to any message to start a thread
- **Single-level only** — replies within a thread cannot be threaded further
- Thread replies count shown on the parent message (e.g., "5 replies")
- Thread panel opens alongside the channel/DM view
- Thread activity triggers unread indicators on the parent message

### 6. Typing Indicators

- Show "[user] is typing..." when another user is composing a message
- Displayed in both channels and DMs
- Automatically clears after a short timeout (3 seconds of inactivity)

### 7. Emoji Reactions

- React to any message (including thread replies) with emoji
- Multiple reactions per message
- Reaction count and list of who reacted
- Toggle own reaction on/off by clicking
- Emoji picker for selecting reactions

### 8. @Mentions

- Mention users with `@username` syntax
- Autocomplete dropdown when typing `@`
- Mentioned users receive visual highlighting in the message
- Mentions contribute to unread badge counts (mentioned messages have higher priority)

### 9. Unread Tracking & Badges

- Unread message count badge on each channel and DM in the sidebar
- Visual distinction between "has unread messages" and "has unread mentions"
- Mark as read when the user views a channel/DM
- Bold channel/DM name when there are unread messages

### 10. File Attachments

- Attach files to messages in channels and DMs
- Support for images, documents, and other common file types
- Image preview inline in the message
- Non-image files shown as downloadable links with file name and size
- Maximum file size: 15 MB per file

## User Flows

### First-Time User

1. User navigates to the app
2. Clicks "Sign in with Google"
3. Completes Google OAuth flow
4. Automatically joins `#general` channel
5. Sees the main chat interface with sidebar (channels, DMs) and message area

### Sending a Message

1. User selects a channel or DM from the sidebar
2. Types in the message input area
3. Other users in the same channel/DM see the typing indicator
4. User presses Enter or clicks Send
5. Message appears instantly for all participants
6. Typing indicator clears

### Starting a Thread

1. User hovers over a message
2. Clicks the "Reply in thread" action
3. Thread panel opens on the right side
4. User types and sends a reply
5. Parent message shows "1 reply" indicator
6. Other participants see the thread activity in their unread indicators

### Reacting to a Message

1. User hovers over a message
2. Clicks the emoji reaction button
3. Emoji picker opens
4. User selects an emoji
5. Reaction appears below the message with a count of 1
6. Other users can click the same reaction to increment the count

### File Attachment

1. User clicks the attachment button in the message input
2. File picker opens
3. User selects a file (up to 10 MB)
4. File uploads and a preview/link appears in the message input area
5. User sends the message with the attachment
6. Recipients see the file inline (images) or as a download link (other files)

## Non-Functional Requirements

- **Real-time:** Messages, typing indicators, reactions, and thread updates must appear within 1 second for all connected users
- **Responsive:** The UI must render well on both desktop and mobile screens (responsive layout)
- **Performance:** The app should handle up to 200 concurrent users without degradation
- **Security:** All API endpoints require authentication; users can only access channels they are members of; private channel content is never exposed to non-members
- **Data integrity:** Messages, reactions, and file references are persisted in the database; no data loss on server restart

## Out of Scope

- Multiple workspaces
- Video/audio conferencing
- Native mobile apps (the web app is responsive, but no dedicated iOS/Android app)
- Message search (may be added later)
- Bot integrations
- Custom emoji
- User roles/admin panel
- Message pinning
- Channel archiving
- Email notifications
- User status/presence (online/offline)
