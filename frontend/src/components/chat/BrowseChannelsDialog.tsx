import { useState, useEffect } from 'react'
import { api } from '@/lib/api'
import type { Channel } from '@/lib/types'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog'
import { Hash, Loader2, ArrowRight } from 'lucide-react'

interface BrowseChannelsDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  joinedChannelIds: Set<string>
  onJoined: (channel: Channel) => void
}

export function BrowseChannelsDialog({
  open,
  onOpenChange,
  joinedChannelIds,
  onJoined,
}: BrowseChannelsDialogProps) {
  const [channels, setChannels] = useState<Channel[]>([])
  const [loading, setLoading] = useState(false)
  const [joining, setJoining] = useState<string | null>(null)

  useEffect(() => {
    if (!open) return
    setLoading(true)
    api
      .browseChannels()
      .then((all) => {
        // Show channels the user hasn't joined
        setChannels(all.filter((c) => !joinedChannelIds.has(c.id)))
      })
      .catch(() => setChannels([]))
      .finally(() => setLoading(false))
  }, [open, joinedChannelIds])

  const handleJoin = async (channel: Channel) => {
    setJoining(channel.id)
    try {
      await api.joinChannel(channel.id)
      onJoined(channel)
      setChannels((prev) => prev.filter((c) => c.id !== channel.id))
    } catch {
      // Error handling
    } finally {
      setJoining(null)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Browse channels</DialogTitle>
          <DialogDescription>
            Find and join public channels to stay in the loop.
          </DialogDescription>
        </DialogHeader>

        <div className="max-h-72 overflow-y-auto">
          {loading ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
            </div>
          ) : channels.length === 0 ? (
            <p className="py-8 text-center text-sm text-muted-foreground">
              You&apos;re already in all available channels!
            </p>
          ) : (
            <div className="space-y-0.5">
              {channels.map((channel) => (
                <button
                  key={channel.id}
                  onClick={() => handleJoin(channel)}
                  disabled={joining !== null}
                  className="flex w-full items-center gap-3 rounded-lg px-3 py-2.5 text-left transition-colors hover:bg-accent cursor-pointer disabled:opacity-50"
                >
                  <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary/10">
                    <Hash className="h-4 w-4 text-primary" />
                  </div>
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-sm font-medium">
                      {channel.name}
                    </p>
                    {channel.description && (
                      <p className="truncate text-xs text-muted-foreground">
                        {channel.description}
                      </p>
                    )}
                  </div>
                  {joining === channel.id ? (
                    <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
                  ) : (
                    <ArrowRight className="h-4 w-4 text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity" />
                  )}
                </button>
              ))}
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}
