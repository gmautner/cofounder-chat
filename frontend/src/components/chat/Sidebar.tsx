import { cn } from '@/lib/utils'
import type { Channel, Conversation, UnreadCount, AuthUser } from '@/lib/types'
import { useNavigate } from 'react-router-dom'
import {
  Hash,
  Lock,
  Plus,
  MessageSquarePlus,
  Search,
  LogOut,
  ChevronDown,
} from 'lucide-react'
import { useState } from 'react'

interface SidebarProps {
  channels: Channel[]
  conversations: Conversation[]
  unreadCounts: UnreadCount[]
  activeChannelId: string | null
  activeConversationId: string | null
  currentUser: AuthUser
  onCreateChannel: () => void
  onBrowseChannels: () => void
  onNewDM: () => void
  onLogout: () => void
}

export function Sidebar({
  channels,
  conversations,
  unreadCounts,
  activeChannelId,
  activeConversationId,
  currentUser,
  onCreateChannel,
  onBrowseChannels,
  onNewDM,
  onLogout,
}: SidebarProps) {
  const navigate = useNavigate()
  const [channelsOpen, setChannelsOpen] = useState(true)
  const [dmsOpen, setDmsOpen] = useState(true)

  const getUnreadCount = (id: string, type: 'channel' | 'conversation') => {
    return unreadCounts.find((u) => u.id === id && u.type === type)?.count ?? 0
  }

  const getConversationDisplayName = (conv: Conversation) => {
    const others = conv.members.filter((m) => m.id !== currentUser.id)
    if (others.length === 0) return 'You'
    return others.map((m) => m.display_name).join(', ')
  }

  const getConversationAvatar = (conv: Conversation) => {
    const other = conv.members.find((m) => m.id !== currentUser.id)
    return other?.avatar_url || ''
  }

  const getConversationInitials = (conv: Conversation) => {
    const other = conv.members.find((m) => m.id !== currentUser.id)
    if (!other) return 'Y'
    return (
      other.display_name
        ?.split(' ')
        .map((n) => n[0])
        .join('')
        .slice(0, 2)
        .toUpperCase() || '?'
    )
  }

  const userInitials =
    currentUser.display_name
      ?.split(' ')
      .map((n) => n[0])
      .join('')
      .slice(0, 2)
      .toUpperCase() || '?'

  return (
    <div className="flex h-full w-64 flex-shrink-0 flex-col bg-sidebar text-sidebar-foreground">
      {/* Header */}
      <div className="flex h-14 items-center border-b border-sidebar-border px-4">
        <h1 className="text-base font-bold tracking-tight text-sidebar-primary-foreground">
          Cofounder Chat
        </h1>
      </div>

      {/* Scrollable content */}
      <div className="flex-1 overflow-y-auto px-2 py-3 space-y-1">
        {/* Channels section */}
        <div>
          <button
            onClick={() => setChannelsOpen(!channelsOpen)}
            className="flex w-full items-center gap-1 rounded-md px-2 py-1 text-xs font-semibold uppercase tracking-wider text-sidebar-foreground/60 hover:text-sidebar-foreground/80 cursor-pointer"
          >
            <ChevronDown
              className={cn(
                'h-3 w-3 transition-transform',
                !channelsOpen && '-rotate-90',
              )}
            />
            Channels
          </button>

          {channelsOpen && (
            <div className="mt-0.5 space-y-px">
              {channels.map((ch) => {
                const isActive = ch.id === activeChannelId
                const unread = getUnreadCount(ch.id, 'channel')
                return (
                  <button
                    key={ch.id}
                    onClick={() => navigate(`/channels/${ch.id}`)}
                    className={cn(
                      'flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm transition-colors cursor-pointer',
                      isActive
                        ? 'bg-sidebar-accent text-sidebar-accent-foreground font-medium'
                        : 'text-sidebar-foreground/70 hover:bg-sidebar-accent/50 hover:text-sidebar-foreground',
                      unread > 0 && !isActive && 'font-semibold text-sidebar-foreground',
                    )}
                  >
                    {ch.is_private ? (
                      <Lock className="h-3.5 w-3.5 flex-shrink-0 opacity-60" />
                    ) : (
                      <Hash className="h-3.5 w-3.5 flex-shrink-0 opacity-60" />
                    )}
                    <span className="truncate">{ch.name}</span>
                    {unread > 0 && (
                      <span className="ml-auto flex h-5 min-w-5 items-center justify-center rounded-full bg-sidebar-primary px-1.5 text-[10px] font-bold text-sidebar-primary-foreground">
                        {unread > 99 ? '99+' : unread}
                      </span>
                    )}
                  </button>
                )
              })}

              <div className="flex gap-px mt-1">
                <button
                  onClick={onCreateChannel}
                  className="flex flex-1 items-center gap-2 rounded-md px-2 py-1.5 text-sm text-sidebar-foreground/50 hover:bg-sidebar-accent/50 hover:text-sidebar-foreground/80 cursor-pointer"
                >
                  <Plus className="h-3.5 w-3.5" />
                  Create
                </button>
                <button
                  onClick={onBrowseChannels}
                  className="flex flex-1 items-center gap-2 rounded-md px-2 py-1.5 text-sm text-sidebar-foreground/50 hover:bg-sidebar-accent/50 hover:text-sidebar-foreground/80 cursor-pointer"
                >
                  <Search className="h-3.5 w-3.5" />
                  Browse
                </button>
              </div>
            </div>
          )}
        </div>

        {/* DMs section */}
        <div className="pt-2">
          <button
            onClick={() => setDmsOpen(!dmsOpen)}
            className="flex w-full items-center gap-1 rounded-md px-2 py-1 text-xs font-semibold uppercase tracking-wider text-sidebar-foreground/60 hover:text-sidebar-foreground/80 cursor-pointer"
          >
            <ChevronDown
              className={cn(
                'h-3 w-3 transition-transform',
                !dmsOpen && '-rotate-90',
              )}
            />
            Direct Messages
          </button>

          {dmsOpen && (
            <div className="mt-0.5 space-y-px">
              {conversations.map((conv) => {
                const isActive = conv.id === activeConversationId
                const unread = getUnreadCount(conv.id, 'conversation')
                const displayName = getConversationDisplayName(conv)
                const avatar = getConversationAvatar(conv)
                const initials = getConversationInitials(conv)

                return (
                  <button
                    key={conv.id}
                    onClick={() => navigate(`/dm/${conv.id}`)}
                    className={cn(
                      'flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm transition-colors cursor-pointer',
                      isActive
                        ? 'bg-sidebar-accent text-sidebar-accent-foreground font-medium'
                        : 'text-sidebar-foreground/70 hover:bg-sidebar-accent/50 hover:text-sidebar-foreground',
                      unread > 0 && !isActive && 'font-semibold text-sidebar-foreground',
                    )}
                  >
                    {avatar ? (
                      <img
                        src={avatar}
                        alt={displayName}
                        className="h-5 w-5 rounded flex-shrink-0 object-cover"
                      />
                    ) : (
                      <div className="flex h-5 w-5 items-center justify-center rounded bg-sidebar-primary/20 text-[9px] font-bold text-sidebar-primary flex-shrink-0">
                        {initials}
                      </div>
                    )}
                    <span className="truncate">{displayName}</span>
                    {unread > 0 && (
                      <span className="ml-auto flex h-5 min-w-5 items-center justify-center rounded-full bg-sidebar-primary px-1.5 text-[10px] font-bold text-sidebar-primary-foreground">
                        {unread > 99 ? '99+' : unread}
                      </span>
                    )}
                  </button>
                )
              })}

              <button
                onClick={onNewDM}
                className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm text-sidebar-foreground/50 hover:bg-sidebar-accent/50 hover:text-sidebar-foreground/80 cursor-pointer"
              >
                <MessageSquarePlus className="h-3.5 w-3.5" />
                New message
              </button>
            </div>
          )}
        </div>
      </div>

      {/* User footer */}
      <div className="border-t border-sidebar-border px-3 py-3">
        <div className="flex items-center gap-2.5">
          {currentUser.avatar_url ? (
            <img
              src={currentUser.avatar_url}
              alt={currentUser.display_name}
              className="h-8 w-8 rounded-lg object-cover"
            />
          ) : (
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-sidebar-primary/20 text-xs font-bold text-sidebar-primary">
              {userInitials}
            </div>
          )}
          <div className="min-w-0 flex-1">
            <p className="truncate text-sm font-medium text-sidebar-foreground">
              {currentUser.display_name}
            </p>
            <p className="truncate text-[11px] text-sidebar-foreground/50">
              {currentUser.email}
            </p>
          </div>
          <button
            onClick={onLogout}
            className="rounded-md p-1.5 text-sidebar-foreground/40 hover:bg-sidebar-accent hover:text-sidebar-foreground/80 cursor-pointer"
            title="Log out"
          >
            <LogOut className="h-4 w-4" />
          </button>
        </div>
      </div>
    </div>
  )
}
