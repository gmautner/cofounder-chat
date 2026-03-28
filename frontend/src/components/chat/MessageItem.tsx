import { useState } from 'react'
import type { ChatMessage, AuthUser } from '@/lib/types'
import { cn } from '@/lib/utils'
import { Pencil, Trash2, Check, X } from 'lucide-react'

interface MessageItemProps {
  message: ChatMessage
  currentUser: AuthUser
  onEdit: (id: string, content: string) => Promise<void>
  onDelete: (id: string) => Promise<void>
}

function formatTime(iso: string): string {
  const date = new Date(iso)
  const now = new Date()
  const isToday = date.toDateString() === now.toDateString()
  const yesterday = new Date(now)
  yesterday.setDate(yesterday.getDate() - 1)
  const isYesterday = date.toDateString() === yesterday.toDateString()

  const time = date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })

  if (isToday) return time
  if (isYesterday) return `Yesterday ${time}`
  return `${date.toLocaleDateString([], { month: 'short', day: 'numeric' })} ${time}`
}

function renderContent(content: string) {
  // Basic markdown: **bold**, *italic*, `code`, and links
  const parts: (string | React.ReactElement)[] = []
  const regex = /(\*\*(.+?)\*\*)|(\*(.+?)\*)|(`(.+?)`)|((https?:\/\/[^\s]+))/g
  let lastIndex = 0
  let match: RegExpExecArray | null
  let key = 0

  while ((match = regex.exec(content)) !== null) {
    if (match.index > lastIndex) {
      parts.push(content.slice(lastIndex, match.index))
    }
    if (match[1]) {
      parts.push(<strong key={key++} className="font-semibold">{match[2]}</strong>)
    } else if (match[3]) {
      parts.push(<em key={key++}>{match[4]}</em>)
    } else if (match[5]) {
      parts.push(
        <code key={key++} className="rounded bg-muted px-1.5 py-0.5 text-[0.85em] font-mono">
          {match[6]}
        </code>
      )
    } else if (match[7]) {
      parts.push(
        <a key={key++} href={match[8]} target="_blank" rel="noopener noreferrer"
          className="text-primary underline underline-offset-2 hover:text-primary/80">
          {match[8]}
        </a>
      )
    }
    lastIndex = match.index + match[0].length
  }
  if (lastIndex < content.length) {
    parts.push(content.slice(lastIndex))
  }
  return parts.length > 0 ? parts : content
}

export function MessageItem({ message, currentUser, onEdit, onDelete }: MessageItemProps) {
  const [isEditing, setIsEditing] = useState(false)
  const [editContent, setEditContent] = useState(message.content)
  const [isHovered, setIsHovered] = useState(false)
  const isOwn = message.user_id === currentUser.id

  const handleSaveEdit = async () => {
    if (editContent.trim() && editContent !== message.content) {
      await onEdit(message.id, editContent.trim())
    }
    setIsEditing(false)
  }

  const handleCancelEdit = () => {
    setEditContent(message.content)
    setIsEditing(false)
  }

  const handleEditKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSaveEdit()
    } else if (e.key === 'Escape') {
      handleCancelEdit()
    }
  }

  const initials = message.author_name
    ?.split(' ')
    .map((n) => n[0])
    .join('')
    .slice(0, 2)
    .toUpperCase() || '?'

  return (
    <div
      className={cn(
        'group relative flex gap-3 px-5 py-1.5 transition-colors',
        isHovered && 'bg-accent/50',
      )}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      {/* Avatar */}
      <div className="mt-0.5 flex-shrink-0">
        {message.author_avatar ? (
          <img
            src={message.author_avatar}
            alt={message.author_name}
            className="h-9 w-9 rounded-lg object-cover"
          />
        ) : (
          <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary/10 text-xs font-semibold text-primary">
            {initials}
          </div>
        )}
      </div>

      {/* Content */}
      <div className="min-w-0 flex-1">
        <div className="flex items-baseline gap-2">
          <span className="text-sm font-semibold text-foreground">
            {message.author_name}
          </span>
          <span className="text-xs text-muted-foreground">
            {formatTime(message.created_at)}
          </span>
          {message.is_edited && (
            <span className="text-xs text-muted-foreground/60">(edited)</span>
          )}
        </div>

        {isEditing ? (
          <div className="mt-1">
            <textarea
              value={editContent}
              onChange={(e) => setEditContent(e.target.value)}
              onKeyDown={handleEditKeyDown}
              className="w-full resize-none rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              rows={2}
              autoFocus
            />
            <div className="mt-1 flex items-center gap-1.5">
              <button
                onClick={handleSaveEdit}
                className="inline-flex items-center gap-1 rounded-md bg-primary px-2.5 py-1 text-xs font-medium text-primary-foreground hover:bg-primary/90 cursor-pointer"
              >
                <Check className="h-3 w-3" /> Save
              </button>
              <button
                onClick={handleCancelEdit}
                className="inline-flex items-center gap-1 rounded-md px-2.5 py-1 text-xs font-medium text-muted-foreground hover:bg-muted cursor-pointer"
              >
                <X className="h-3 w-3" /> Cancel
              </button>
              <span className="text-xs text-muted-foreground">
                Esc to cancel · Enter to save
              </span>
            </div>
          </div>
        ) : (
          <p className="text-sm leading-relaxed text-foreground/90 whitespace-pre-wrap break-words">
            {renderContent(message.content)}
          </p>
        )}

        {message.reply_count > 0 && (
          <div className="mt-1 inline-flex items-center gap-1 text-xs font-medium text-primary cursor-pointer hover:underline">
            {message.reply_count} {message.reply_count === 1 ? 'reply' : 'replies'}
          </div>
        )}
      </div>

      {/* Action buttons on hover */}
      {isOwn && isHovered && !isEditing && (
        <div className="absolute right-4 -top-3 flex items-center gap-0.5 rounded-md border bg-background p-0.5 shadow-sm">
          <button
            onClick={() => {
              setEditContent(message.content)
              setIsEditing(true)
            }}
            className="rounded p-1.5 text-muted-foreground hover:bg-muted hover:text-foreground cursor-pointer"
            title="Edit message"
          >
            <Pencil className="h-3.5 w-3.5" />
          </button>
          <button
            onClick={() => onDelete(message.id)}
            className="rounded p-1.5 text-muted-foreground hover:bg-destructive/10 hover:text-destructive cursor-pointer"
            title="Delete message"
          >
            <Trash2 className="h-3.5 w-3.5" />
          </button>
        </div>
      )}
    </div>
  )
}
