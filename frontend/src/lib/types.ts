export interface User {
  id: string
  google_id: string
  email: string
  display_name: string
  avatar_url: string
  created_at: string
  updated_at: string
}

export interface AuthUser {
  id: string
  email: string
  display_name: string
  avatar_url: string
}

export interface Channel {
  id: string
  name: string
  description: string
  is_private: boolean
  created_by: string
  created_at: string
  updated_at: string
}

export interface ChatMessage {
  id: string
  user_id: string
  channel_id: string | null
  conversation_id: string | null
  parent_id: string | null
  content: string
  is_edited: boolean
  created_at: string
  updated_at: string
  author_name: string
  author_avatar: string
  reply_count: number
}

export interface Conversation {
  id: string
  created_at: string
  members: ConversationMember[]
}

export interface ConversationMember {
  id: string
  display_name: string
  avatar_url: string
  email: string
}

export interface UnreadCount {
  id: string
  type: 'channel' | 'conversation'
  count: number
}

export interface TypingUser {
  userId: string
  displayName: string
  timeout: ReturnType<typeof setTimeout>
}
