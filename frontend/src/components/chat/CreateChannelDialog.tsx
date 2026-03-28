import { useState } from 'react'
import { api } from '@/lib/api'
import type { Channel } from '@/lib/types'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Hash, Lock, Loader2 } from 'lucide-react'

interface CreateChannelDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onCreated: (channel: Channel) => void
}

export function CreateChannelDialog({
  open,
  onOpenChange,
  onCreated,
}: CreateChannelDialogProps) {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [isPrivate, setIsPrivate] = useState(false)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    const channelName = name.trim().toLowerCase().replace(/\s+/g, '-')
    if (!channelName) return

    setLoading(true)
    setError('')
    try {
      const channel = await api.createChannel(
        channelName,
        description.trim(),
        isPrivate,
      )
      onCreated(channel)
      onOpenChange(false)
      setName('')
      setDescription('')
      setIsPrivate(false)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create channel')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Criar um canal</DialogTitle>
          <DialogDescription>
            Canais são onde seu time se comunica. Organize-os por assunto.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="text-sm font-medium text-foreground">Nome</label>
            <div className="relative mt-1.5">
              <span className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground">
                {isPrivate ? (
                  <Lock className="h-4 w-4" />
                ) : (
                  <Hash className="h-4 w-4" />
                )}
              </span>
              <Input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="ex: marketing"
                className="pl-9"
                autoFocus
              />
            </div>
          </div>

          <div>
            <label className="text-sm font-medium text-foreground">
              Descrição{' '}
              <span className="font-normal text-muted-foreground">
                (opcional)
              </span>
            </label>
            <Input
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Sobre o que é este canal?"
              className="mt-1.5"
            />
          </div>

          <div className="flex items-center gap-3">
            <button
              type="button"
              onClick={() => setIsPrivate(!isPrivate)}
              className={`relative h-6 w-11 rounded-full transition-colors cursor-pointer ${
                isPrivate ? 'bg-primary' : 'bg-muted'
              }`}
            >
              <span
                className={`absolute top-0.5 left-0.5 h-5 w-5 rounded-full bg-white shadow-sm transition-transform ${
                  isPrivate ? 'translate-x-5' : ''
                }`}
              />
            </button>
            <div>
              <p className="text-sm font-medium">Canal privado</p>
              <p className="text-xs text-muted-foreground">
                Apenas membros convidados podem ver e entrar
              </p>
            </div>
          </div>

          {error && (
            <p className="text-sm text-destructive">{error}</p>
          )}

          <div className="flex justify-end gap-2 pt-2">
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              Cancelar
            </Button>
            <Button type="submit" disabled={!name.trim() || loading}>
              {loading ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : null}
              Criar
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}
