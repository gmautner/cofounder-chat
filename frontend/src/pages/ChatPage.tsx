import { useState, useEffect, useCallback, useRef } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { useAuth } from '@/hooks/useAuth'
import { useSSE } from '@/hooks/useSSE'
import { api } from '@/lib/api'
import type {
  Channel,
  Conversation,
  ChatMessage,
  UnreadCount,
  TypingUser,
} from '@/lib/types'
import { Sidebar } from '@/components/chat/Sidebar'
import { MessageList } from '@/components/chat/MessageList'
import { MessageInput } from '@/components/chat/MessageInput'
import { CreateChannelDialog } from '@/components/chat/CreateChannelDialog'
import { NewDMDialog } from '@/components/chat/NewDMDialog'
import { BrowseChannelsDialog } from '@/components/chat/BrowseChannelsDialog'
import { Hash, Lock, Users, MessageSquare } from 'lucide-react'

export function ChatPage() {
  const { user, logout } = useAuth()
  const location = useLocation()
  const navigate = useNavigate()

  // Parse active view from URL
  const channelMatch = location.pathname.match(/^\/channels\/([^/]+)/)
  const dmMatch = location.pathname.match(/^\/dm\/([^/]+)/)
  const activeChannelId = channelMatch?.[1] ?? null
  const activeConversationId = dmMatch?.[1] ?? null

  // Core state
  const [channels, setChannels] = useState<Channel[]>([])
  const [conversations, setConversations] = useState<Conversation[]>([])
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [unreadCounts, setUnreadCounts] = useState<UnreadCount[]>([])
  const [typingUsers, setTypingUsers] = useState<Map<string, TypingUser>>(
    new Map(),
  )

  // Dialog state
  const [showCreateChannel, setShowCreateChannel] = useState(false)
  const [showBrowseChannels, setShowBrowseChannels] = useState(false)
  const [showNewDM, setShowNewDM] = useState(false)

  // Loading state
  const [loadingMessages, setLoadingMessages] = useState(false)

  // Refs for SSE handler closures
  const activeChannelIdRef = useRef(activeChannelId)
  const activeConversationIdRef = useRef(activeConversationId)
  activeChannelIdRef.current = activeChannelId
  activeConversationIdRef.current = activeConversationId

  // Load initial data
  useEffect(() => {
    api.listChannels().then(setChannels).catch(console.error)
    api.listConversations().then(setConversations).catch(console.error)
    api.getUnreadCounts().then(setUnreadCounts).catch(console.error)
  }, [])

  // Redirect to #general if at root
  useEffect(() => {
    if (
      !activeChannelId &&
      !activeConversationId &&
      channels.length > 0 &&
      location.pathname === '/'
    ) {
      const general = channels.find((c) => c.name === 'general')
      if (general) {
        navigate(`/channels/${general.id}`, { replace: true })
      } else {
        navigate(`/channels/${channels[0].id}`, { replace: true })
      }
    }
  }, [channels, activeChannelId, activeConversationId, navigate, location.pathname])

  // Load messages when view changes
  useEffect(() => {
    setMessages([])
    setTypingUsers(new Map())

    if (activeChannelId) {
      setLoadingMessages(true)
      api
        .getChannelMessages(activeChannelId)
        .then(setMessages)
        .catch(console.error)
        .finally(() => setLoadingMessages(false))
      api.markChannelRead(activeChannelId)
      // Clear unread for this channel
      setUnreadCounts((prev) =>
        prev.filter(
          (u) => !(u.id === activeChannelId && u.type === 'channel'),
        ),
      )
    } else if (activeConversationId) {
      setLoadingMessages(true)
      api
        .getConversationMessages(activeConversationId)
        .then(setMessages)
        .catch(console.error)
        .finally(() => setLoadingMessages(false))
      api.markConversationRead(activeConversationId)
      setUnreadCounts((prev) =>
        prev.filter(
          (u) =>
            !(u.id === activeConversationId && u.type === 'conversation'),
        ),
      )
    }
  }, [activeChannelId, activeConversationId])

  // SSE real-time events
  useSSE({
    onNewMessage: useCallback(
      (data: { message: Record<string, unknown>; channel_id?: string; conversation_id?: string; author_name: string; author_avatar: string }) => {
        const msg = data.message as unknown as ChatMessage
        const eventChannelId = data.channel_id
        const eventConvId = data.conversation_id

        // Build enriched message
        const enriched: ChatMessage = {
          ...msg,
          author_name: data.author_name || msg.author_name,
          author_avatar: data.author_avatar || msg.author_avatar,
          reply_count: msg.reply_count ?? 0,
        }

        // If the message is for the active view, append it
        if (
          (activeChannelIdRef.current && eventChannelId === activeChannelIdRef.current) ||
          (activeConversationIdRef.current && eventConvId === activeConversationIdRef.current)
        ) {
          setMessages((prev) => {
            // Avoid duplicates
            if (prev.some((m) => m.id === enriched.id)) return prev
            return [...prev, enriched]
          })
        } else {
          // Increment unread
          const id = eventChannelId || eventConvId
          const type = eventChannelId ? 'channel' : 'conversation'
          if (id) {
            setUnreadCounts((prev) => {
              const existing = prev.find(
                (u) => u.id === id && u.type === type,
              )
              if (existing) {
                return prev.map((u) =>
                  u.id === id && u.type === type
                    ? { ...u, count: u.count + 1 }
                    : u,
                )
              }
              return [...prev, { id, type: type as 'channel' | 'conversation', count: 1 }]
            })
          }
        }

        // Refresh conversations list if it's a DM
        if (eventConvId) {
          api.listConversations().then(setConversations).catch(console.error)
        }
      },
      [],
    ),

    onMessageUpdated: useCallback((data: Record<string, unknown>) => {
      const updated = data as unknown as ChatMessage
      setMessages((prev) =>
        prev.map((m) =>
          m.id === updated.id
            ? {
                ...m,
                content: updated.content,
                is_edited: updated.is_edited,
                updated_at: updated.updated_at,
              }
            : m,
        ),
      )
    }, []),

    onMessageDeleted: useCallback((data: { id: string; channel_id?: string | null; conversation_id?: string | null }) => {
      setMessages((prev) => prev.filter((m) => m.id !== data.id))
    }, []),

    onTyping: useCallback(
      (data: { user_id: string; display_name: string; channel_id: string; conversation_id: string }) => {
        if (data.user_id === user?.id) return

        const relevantChannel =
          activeChannelIdRef.current &&
          data.channel_id === activeChannelIdRef.current
        const relevantConv =
          activeConversationIdRef.current &&
          data.conversation_id === activeConversationIdRef.current

        if (!relevantChannel && !relevantConv) return

        setTypingUsers((prev) => {
          const next = new Map(prev)
          const existing = next.get(data.user_id)
          if (existing?.timeout) clearTimeout(existing.timeout)

          const timeout = setTimeout(() => {
            setTypingUsers((p) => {
              const n = new Map(p)
              n.delete(data.user_id)
              return n
            })
          }, 3000)

          next.set(data.user_id, {
            userId: data.user_id,
            displayName: data.display_name,
            timeout,
          })
          return next
        })
      },
      [user?.id],
    ),
  })

  // Message actions
  const handleSendMessage = async (content: string) => {
    if (activeChannelId) {
      await api.sendChannelMessage(activeChannelId, content)
    } else if (activeConversationId) {
      await api.sendConversationMessage(activeConversationId, content)
    }
  }

  const handleEditMessage = async (id: string, content: string) => {
    await api.updateMessage(id, content)
  }

  const handleDeleteMessage = async (id: string) => {
    await api.deleteMessage(id)
  }

  const handleTyping = () => {
    api.sendTyping(
      activeChannelId || undefined,
      activeConversationId || undefined,
    )
  }

  // Channel/DM creation handlers
  const handleChannelCreated = (channel: Channel) => {
    setChannels((prev) => [...prev, channel].sort((a, b) => a.name.localeCompare(b.name)))
    navigate(`/channels/${channel.id}`)
  }

  const handleChannelJoined = (channel: Channel) => {
    setChannels((prev) => [...prev, channel].sort((a, b) => a.name.localeCompare(b.name)))
    navigate(`/channels/${channel.id}`)
  }

  const handleConversationCreated = (convId: string) => {
    api.listConversations().then(setConversations).catch(console.error)
    navigate(`/dm/${convId}`)
  }

  // Get active channel/conversation info for header
  const activeChannel = channels.find((c) => c.id === activeChannelId)
  const activeConversation = conversations.find(
    (c) => c.id === activeConversationId,
  )

  const headerTitle = activeChannel
    ? activeChannel.name
    : activeConversation
      ? activeConversation.members
          .filter((m) => m.id !== user?.id)
          .map((m) => m.display_name)
          .join(', ') || 'Direct Message'
      : ''

  const headerDescription = activeChannel?.description || ''

  const typingNames = Array.from(typingUsers.values()).map(
    (t) => t.displayName,
  )

  if (!user) return null

  return (
    <div className="flex h-screen bg-background">
      <Sidebar
        channels={channels}
        conversations={conversations}
        unreadCounts={unreadCounts}
        activeChannelId={activeChannelId}
        activeConversationId={activeConversationId}
        currentUser={user}
        onCreateChannel={() => setShowCreateChannel(true)}
        onBrowseChannels={() => setShowBrowseChannels(true)}
        onNewDM={() => setShowNewDM(true)}
        onLogout={logout}
      />

      {/* Main content */}
      <main className="flex flex-1 flex-col overflow-hidden">
        {activeChannelId || activeConversationId ? (
          <>
            {/* Header */}
            <header className="flex h-14 flex-shrink-0 items-center border-b border-border px-5">
              <div className="flex items-baseline gap-2">
                <div className="flex items-center gap-2">
                  {activeChannel ? (
                    activeChannel.is_private ? (
                      <Lock className="h-4 w-4 text-muted-foreground" />
                    ) : (
                      <Hash className="h-4 w-4 text-muted-foreground" />
                    )
                  ) : (
                    <Users className="h-4 w-4 text-muted-foreground" />
                  )}
                  <h2 className="text-sm font-semibold">{headerTitle}</h2>
                </div>
                {headerDescription && (
                  <>
                    <span className="text-muted-foreground/30">|</span>
                    <p className="text-sm text-muted-foreground truncate max-w-md">
                      {headerDescription}
                    </p>
                  </>
                )}
              </div>
            </header>

            {/* Messages */}
            {loadingMessages ? (
              <div className="flex flex-1 items-center justify-center">
                <div className="h-5 w-5 animate-spin rounded-full border-2 border-primary border-t-transparent" />
              </div>
            ) : (
              <MessageList
                messages={messages}
                currentUser={user}
                onEditMessage={handleEditMessage}
                onDeleteMessage={handleDeleteMessage}
                typingNames={typingNames}
              />
            )}

            {/* Input */}
            <MessageInput
              onSend={handleSendMessage}
              onTyping={handleTyping}
              placeholder={
                activeChannel
                  ? `Mensagem no #${activeChannel.name}`
                  : `Mensagem para ${headerTitle}`
              }
            />
          </>
        ) : (
          /* Empty state */
          <div className="flex flex-1 flex-col items-center justify-center text-muted-foreground">
            <div className="mb-4 flex h-20 w-20 items-center justify-center rounded-2xl bg-primary/5">
              <MessageSquare className="h-10 w-10 text-primary/30" />
            </div>
            <h2 className="text-lg font-semibold text-foreground">
              Bem-vindo ao Cofounder Chat
            </h2>
            <p className="mt-1 text-sm">
              Selecione um canal ou conversa para começar
            </p>
          </div>
        )}
      </main>

      {/* Dialogs */}
      <CreateChannelDialog
        open={showCreateChannel}
        onOpenChange={setShowCreateChannel}
        onCreated={handleChannelCreated}
      />

      <BrowseChannelsDialog
        open={showBrowseChannels}
        onOpenChange={setShowBrowseChannels}
        joinedChannelIds={new Set(channels.map((c) => c.id))}
        onJoined={handleChannelJoined}
      />

      <NewDMDialog
        open={showNewDM}
        onOpenChange={setShowNewDM}
        currentUser={user}
        onConversationCreated={handleConversationCreated}
      />
    </div>
  )
}
