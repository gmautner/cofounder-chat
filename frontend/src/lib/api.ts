import type {
  AuthUser,
  Channel,
  ChatMessage,
  Conversation,
  UnreadCount,
  User,
} from './types'

class ApiError extends Error {
  status: number
  constructor(message: string, status: number) {
    super(message)
    this.status = status
    this.name = 'ApiError'
  }
}

async function fetchJSON<T>(url: string, options?: RequestInit): Promise<T> {
  const res = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: 'Request failed' }))
    throw new ApiError(body.error || `HTTP ${res.status}`, res.status)
  }
  return res.json()
}

export const api = {
  // Auth
  getMe: () => fetchJSON<AuthUser>('/api/me'),
  logout: () => fetch('/auth/logout', { method: 'POST' }),
  devLogin: (email: string) =>
    fetch('/api/dev/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email }),
    }),

  // Channels
  listChannels: () => fetchJSON<Channel[]>('/api/channels'),
  browseChannels: () => fetchJSON<Channel[]>('/api/channels/browse'),
  createChannel: (name: string, description: string, isPrivate: boolean) =>
    fetchJSON<Channel>('/api/channels', {
      method: 'POST',
      body: JSON.stringify({ name, description, is_private: isPrivate }),
    }),
  getChannel: (id: string) => fetchJSON<Channel>(`/api/channels/${id}`),
  joinChannel: (id: string) =>
    fetchJSON<{ status: string }>(`/api/channels/${id}/join`, { method: 'POST' }),
  leaveChannel: (id: string) =>
    fetchJSON<{ status: string }>(`/api/channels/${id}/leave`, { method: 'POST' }),
  getChannelMembers: (id: string) => fetchJSON<User[]>(`/api/channels/${id}/members`),
  getChannelMessages: (id: string) =>
    fetchJSON<ChatMessage[]>(`/api/channels/${id}/messages`),
  sendChannelMessage: (channelId: string, content: string, parentId?: string) =>
    fetchJSON<ChatMessage>(`/api/channels/${channelId}/messages`, {
      method: 'POST',
      body: JSON.stringify({ content, parent_id: parentId }),
    }),
  markChannelRead: (id: string) =>
    fetch(`/api/channels/${id}/read`, { method: 'POST' }),

  // Conversations
  listConversations: () => fetchJSON<Conversation[]>('/api/conversations'),
  createConversation: (userId: string) =>
    fetchJSON<{ id: string; existing?: boolean }>('/api/conversations', {
      method: 'POST',
      body: JSON.stringify({ user_id: userId }),
    }),
  getConversationMessages: (id: string) =>
    fetchJSON<ChatMessage[]>(`/api/conversations/${id}/messages`),
  sendConversationMessage: (
    convId: string,
    content: string,
    parentId?: string,
  ) =>
    fetchJSON<ChatMessage>(`/api/conversations/${convId}/messages`, {
      method: 'POST',
      body: JSON.stringify({ content, parent_id: parentId }),
    }),
  getConversationMembers: (id: string) =>
    fetchJSON<User[]>(`/api/conversations/${id}/members`),
  markConversationRead: (id: string) =>
    fetch(`/api/conversations/${id}/read`, { method: 'POST' }),

  // Messages
  updateMessage: (id: string, content: string) =>
    fetchJSON<ChatMessage>(`/api/messages/${id}`, {
      method: 'PUT',
      body: JSON.stringify({ content }),
    }),
  deleteMessage: (id: string) =>
    fetch(`/api/messages/${id}`, { method: 'DELETE' }),
  getThread: (id: string) =>
    fetchJSON<{ parent: ChatMessage; replies: ChatMessage[] }>(
      `/api/messages/${id}/thread`,
    ),

  // Users
  listUsers: () => fetchJSON<User[]>('/api/users'),
  searchUsers: (q: string) =>
    fetchJSON<User[]>(`/api/users/search?q=${encodeURIComponent(q)}`),

  // Unread
  getUnreadCounts: () => fetchJSON<UnreadCount[]>('/api/unread'),

  // Typing
  sendTyping: (channelId?: string, conversationId?: string) =>
    fetch('/api/typing', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        channel_id: channelId || '',
        conversation_id: conversationId || '',
      }),
    }),
}
