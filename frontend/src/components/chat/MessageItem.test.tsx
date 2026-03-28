import { render, screen } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { MessageItem } from './MessageItem'
import type { ChatMessage, AuthUser } from '@/lib/types'

const currentUser: AuthUser = {
  id: 'user-1',
  email: 'me@test.com',
  display_name: 'Me',
  avatar_url: '',
}

const baseMessage: ChatMessage = {
  id: 'msg-1',
  user_id: 'user-1',
  channel_id: 'ch-1',
  conversation_id: null,
  parent_id: null,
  content: 'Hello everyone!',
  is_edited: false,
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
  author_name: 'Me',
  author_avatar: '',
  reply_count: 0,
}

describe('MessageItem', () => {
  it('renders message content', () => {
    render(
      <MessageItem
        message={baseMessage}
        currentUser={currentUser}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
      />,
    )
    expect(screen.getByText('Hello everyone!')).toBeInTheDocument()
  })

  it('renders author name', () => {
    render(
      <MessageItem
        message={baseMessage}
        currentUser={currentUser}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
      />,
    )
    expect(screen.getByText('Me')).toBeInTheDocument()
  })

  it('shows (edited) label when message is edited', () => {
    const editedMsg = { ...baseMessage, is_edited: true }
    render(
      <MessageItem
        message={editedMsg}
        currentUser={currentUser}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
      />,
    )
    expect(screen.getByText('(edited)')).toBeInTheDocument()
  })

  it('does not show (edited) for non-edited messages', () => {
    render(
      <MessageItem
        message={baseMessage}
        currentUser={currentUser}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
      />,
    )
    expect(screen.queryByText('(edited)')).not.toBeInTheDocument()
  })

  it('shows reply count when there are replies', () => {
    const msgWithReplies = { ...baseMessage, reply_count: 5 }
    render(
      <MessageItem
        message={msgWithReplies}
        currentUser={currentUser}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
      />,
    )
    expect(screen.getByText('5 replies')).toBeInTheDocument()
  })

  it('shows singular reply text for 1 reply', () => {
    const msgWithReply = { ...baseMessage, reply_count: 1 }
    render(
      <MessageItem
        message={msgWithReply}
        currentUser={currentUser}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
      />,
    )
    expect(screen.getByText('1 reply')).toBeInTheDocument()
  })

  it('does not show action buttons for other users messages', () => {
    const otherUserMsg = { ...baseMessage, user_id: 'user-2', author_name: 'Other' }
    render(
      <MessageItem
        message={otherUserMsg}
        currentUser={currentUser}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
      />,
    )
    // Action buttons should not be in the DOM for other user's messages
    expect(screen.queryByTitle('Edit message')).not.toBeInTheDocument()
    expect(screen.queryByTitle('Delete message')).not.toBeInTheDocument()
  })

  it('renders initials when no avatar', () => {
    const msg = { ...baseMessage, author_name: 'John Doe', author_avatar: '' }
    render(
      <MessageItem
        message={msg}
        currentUser={currentUser}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
      />,
    )
    expect(screen.getByText('JD')).toBeInTheDocument()
  })

  it('renders bold markdown', () => {
    const msg = { ...baseMessage, content: 'This is **bold** text' }
    render(
      <MessageItem
        message={msg}
        currentUser={currentUser}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
      />,
    )
    const bold = screen.getByText('bold')
    expect(bold.tagName).toBe('STRONG')
  })
})
