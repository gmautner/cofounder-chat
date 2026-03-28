import { useState, useRef, useCallback, useEffect } from 'react'
import { SendHorizontal } from 'lucide-react'

interface MessageInputProps {
  onSend: (content: string) => Promise<void>
  onTyping: () => void
  placeholder?: string
  disabled?: boolean
}

export function MessageInput({
  onSend,
  onTyping,
  placeholder = 'Type a message...',
  disabled = false,
}: MessageInputProps) {
  const [content, setContent] = useState('')
  const [sending, setSending] = useState(false)
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const typingTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const adjustHeight = useCallback(() => {
    const el = textareaRef.current
    if (!el) return
    el.style.height = 'auto'
    el.style.height = `${Math.min(el.scrollHeight, 160)}px`
  }, [])

  useEffect(() => {
    adjustHeight()
  }, [content, adjustHeight])

  const handleSend = async () => {
    const trimmed = content.trim()
    if (!trimmed || sending) return
    setSending(true)
    try {
      await onSend(trimmed)
      setContent('')
      if (textareaRef.current) {
        textareaRef.current.style.height = 'auto'
      }
    } finally {
      setSending(false)
      textareaRef.current?.focus()
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setContent(e.target.value)

    // Throttle typing indicator to every 2 seconds
    if (!typingTimerRef.current) {
      onTyping()
      typingTimerRef.current = setTimeout(() => {
        typingTimerRef.current = null
      }, 2000)
    }
  }

  return (
    <div className="border-t border-border bg-background px-5 py-3">
      <div className="flex items-end gap-2 rounded-xl border border-input bg-background px-4 py-2 focus-within:ring-2 focus-within:ring-ring/50 transition-shadow">
        <textarea
          ref={textareaRef}
          value={content}
          onChange={handleChange}
          onKeyDown={handleKeyDown}
          placeholder={placeholder}
          disabled={disabled}
          rows={1}
          className="flex-1 resize-none bg-transparent text-sm leading-relaxed text-foreground placeholder:text-muted-foreground focus:outline-none disabled:opacity-50"
        />
        <button
          onClick={handleSend}
          disabled={!content.trim() || sending || disabled}
          className="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-lg bg-primary text-primary-foreground transition-all hover:bg-primary/90 disabled:opacity-30 disabled:cursor-not-allowed cursor-pointer"
        >
          <SendHorizontal className="h-4 w-4" />
        </button>
      </div>
      <p className="mt-1.5 text-[11px] text-muted-foreground/60 px-1">
        <kbd className="font-mono">Enter</kbd> to send ·{' '}
        <kbd className="font-mono">Shift+Enter</kbd> for new line
      </p>
    </div>
  )
}
