import { useEffect, useRef } from 'react'
import type { ChatMessage, AuthUser } from '@/lib/types'
import { MessageItem } from './MessageItem'

interface MessageListProps {
  messages: ChatMessage[]
  currentUser: AuthUser
  onEditMessage: (id: string, content: string) => Promise<void>
  onDeleteMessage: (id: string) => Promise<void>
  typingNames: string[]
}

export function MessageList({
  messages,
  currentUser,
  onEditMessage,
  onDeleteMessage,
  typingNames,
}: MessageListProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const isAtBottomRef = useRef(true)

  const scrollToBottom = () => {
    if (containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight
    }
  }

  const handleScroll = () => {
    const el = containerRef.current
    if (!el) return
    isAtBottomRef.current =
      el.scrollHeight - el.scrollTop - el.clientHeight < 80
  }

  useEffect(() => {
    if (isAtBottomRef.current) {
      scrollToBottom()
    }
  }, [messages])

  // Scroll to bottom on first load
  useEffect(() => {
    scrollToBottom()
  }, [])

  return (
    <div
      ref={containerRef}
      onScroll={handleScroll}
      className="flex-1 overflow-y-auto"
    >
      {messages.length === 0 ? (
        <div className="flex h-full flex-col items-center justify-center text-muted-foreground">
          <div className="mb-3 flex h-16 w-16 items-center justify-center rounded-2xl bg-primary/5">
            <svg
              className="h-8 w-8 text-primary/40"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              strokeWidth={1.5}
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M8.625 12a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Zm0 0H8.25m4.125 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Zm0 0H12m4.125 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Zm0 0h-.375M21 12c0 4.556-4.03 8.25-9 8.25a9.764 9.764 0 0 1-2.555-.337A5.972 5.972 0 0 1 5.41 20.97a5.969 5.969 0 0 1-.474-.065 4.48 4.48 0 0 0 .978-2.025c.09-.457-.133-.901-.467-1.226C3.93 16.178 3 14.189 3 12c0-4.556 4.03-8.25 9-8.25s9 3.694 9 8.25Z"
              />
            </svg>
          </div>
          <p className="text-sm font-medium">No messages yet</p>
          <p className="mt-1 text-xs">Start the conversation!</p>
        </div>
      ) : (
        <div className="py-4">
          {messages.map((msg) => (
            <MessageItem
              key={msg.id}
              message={msg}
              currentUser={currentUser}
              onEdit={onEditMessage}
              onDelete={onDeleteMessage}
            />
          ))}
        </div>
      )}

      {/* Typing indicator */}
      {typingNames.length > 0 && (
        <div className="px-5 pb-2">
          <div className="flex items-center gap-2 text-xs text-muted-foreground">
            <span className="inline-flex gap-0.5">
              <span className="animate-bounce [animation-delay:0ms] h-1.5 w-1.5 rounded-full bg-primary/50" />
              <span className="animate-bounce [animation-delay:150ms] h-1.5 w-1.5 rounded-full bg-primary/50" />
              <span className="animate-bounce [animation-delay:300ms] h-1.5 w-1.5 rounded-full bg-primary/50" />
            </span>
            <span className="font-medium">
              {typingNames.join(', ')}{' '}
              {typingNames.length === 1 ? 'is' : 'are'} typing...
            </span>
          </div>
        </div>
      )}
    </div>
  )
}
