import { useState, useEffect, useCallback } from 'react'
import { api } from '@/lib/api'
import type { User, AuthUser } from '@/lib/types'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Loader2 } from 'lucide-react'

interface NewDMDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  currentUser: AuthUser
  onConversationCreated: (convId: string) => void
}

export function NewDMDialog({
  open,
  onOpenChange,
  currentUser,
  onConversationCreated,
}: NewDMDialogProps) {
  const [search, setSearch] = useState('')
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(false)
  const [creating, setCreating] = useState<string | null>(null)

  const loadUsers = useCallback(async () => {
    setLoading(true)
    try {
      if (search.trim()) {
        const results = await api.searchUsers(search.trim())
        setUsers(results.filter((u) => u.id !== currentUser.id))
      } else {
        const all = await api.listUsers()
        setUsers(all.filter((u) => u.id !== currentUser.id))
      }
    } catch {
      setUsers([])
    } finally {
      setLoading(false)
    }
  }, [search, currentUser.id])

  useEffect(() => {
    if (!open) return
    const timer = setTimeout(loadUsers, 300)
    return () => clearTimeout(timer)
  }, [open, loadUsers])

  const handleSelectUser = async (userId: string) => {
    setCreating(userId)
    try {
      const result = await api.createConversation(userId)
      onConversationCreated(result.id)
      onOpenChange(false)
      setSearch('')
    } catch {
      // Error handling
    } finally {
      setCreating(null)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Nova mensagem</DialogTitle>
          <DialogDescription>
            Inicie uma conversa direta com alguém do seu time.
          </DialogDescription>
        </DialogHeader>

        <Input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Buscar por nome ou email..."
          autoFocus
        />

        <div className="mt-2 max-h-64 overflow-y-auto">
          {loading ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
            </div>
          ) : users.length === 0 ? (
            <p className="py-8 text-center text-sm text-muted-foreground">
              {search ? 'Nenhum usuário encontrado' : 'Nenhum outro usuário ainda'}
            </p>
          ) : (
            <div className="space-y-0.5">
              {users.map((user) => {
                const initials = user.display_name
                  ?.split(' ')
                  .map((n) => n[0])
                  .join('')
                  .slice(0, 2)
                  .toUpperCase() || '?'

                return (
                  <button
                    key={user.id}
                    onClick={() => handleSelectUser(user.id)}
                    disabled={creating !== null}
                    className="flex w-full items-center gap-3 rounded-lg px-3 py-2.5 text-left transition-colors hover:bg-accent cursor-pointer disabled:opacity-50"
                  >
                    {user.avatar_url ? (
                      <img
                        src={user.avatar_url}
                        alt={user.display_name}
                        className="h-9 w-9 rounded-lg object-cover"
                      />
                    ) : (
                      <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary/10 text-xs font-semibold text-primary">
                        {initials}
                      </div>
                    )}
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-sm font-medium">
                        {user.display_name}
                      </p>
                      <p className="truncate text-xs text-muted-foreground">
                        {user.email}
                      </p>
                    </div>
                    {creating === user.id && (
                      <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
                    )}
                  </button>
                )
              })}
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}
