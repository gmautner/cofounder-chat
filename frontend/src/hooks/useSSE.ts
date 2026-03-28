import { useEffect, useRef } from 'react'

export interface SSEHandlers {
  onNewMessage?: (data: {
    message: Record<string, unknown>
    channel_id?: string
    conversation_id?: string
    author_name: string
    author_avatar: string
  }) => void
  onMessageUpdated?: (data: Record<string, unknown>) => void
  onMessageDeleted?: (data: {
    id: string
    channel_id?: string | null
    conversation_id?: string | null
  }) => void
  onReactionAdded?: (data: Record<string, unknown>) => void
  onReactionRemoved?: (data: Record<string, unknown>) => void
  onTyping?: (data: {
    user_id: string
    display_name: string
    channel_id: string
    conversation_id: string
  }) => void
}

export function useSSE(handlers: SSEHandlers) {
  const handlersRef = useRef(handlers)
  handlersRef.current = handlers

  useEffect(() => {
    let es: EventSource | null = null
    let reconnectTimer: ReturnType<typeof setTimeout> | null = null

    function connect() {
      es = new EventSource('/api/events')

      es.addEventListener('new_message', (e) => {
        try {
          handlersRef.current.onNewMessage?.(JSON.parse(e.data))
        } catch { /* ignore parse errors */ }
      })

      es.addEventListener('message_updated', (e) => {
        try {
          handlersRef.current.onMessageUpdated?.(JSON.parse(e.data))
        } catch { /* ignore */ }
      })

      es.addEventListener('message_deleted', (e) => {
        try {
          handlersRef.current.onMessageDeleted?.(JSON.parse(e.data))
        } catch { /* ignore */ }
      })

      es.addEventListener('reaction_added', (e) => {
        try {
          handlersRef.current.onReactionAdded?.(JSON.parse(e.data))
        } catch { /* ignore */ }
      })

      es.addEventListener('reaction_removed', (e) => {
        try {
          handlersRef.current.onReactionRemoved?.(JSON.parse(e.data))
        } catch { /* ignore */ }
      })

      es.addEventListener('typing', (e) => {
        try {
          handlersRef.current.onTyping?.(JSON.parse(e.data))
        } catch { /* ignore */ }
      })

      es.onerror = () => {
        es?.close()
        // Reconnect after 3 seconds
        reconnectTimer = setTimeout(connect, 3000)
      }
    }

    connect()

    return () => {
      es?.close()
      if (reconnectTimer) clearTimeout(reconnectTimer)
    }
  }, [])
}
