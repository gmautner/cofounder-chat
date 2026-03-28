import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi } from 'vitest'
import { MessageInput } from './MessageInput'

describe('MessageInput', () => {
  it('renders the input with placeholder', () => {
    render(
      <MessageInput
        onSend={vi.fn()}
        onTyping={vi.fn()}
        placeholder="Message #general"
      />,
    )
    expect(screen.getByPlaceholderText('Message #general')).toBeInTheDocument()
  })

  it('send button is disabled when input is empty', () => {
    render(<MessageInput onSend={vi.fn()} onTyping={vi.fn()} />)
    const sendButton = screen.getByRole('button')
    expect(sendButton).toBeDisabled()
  })

  it('send button is enabled when input has text', async () => {
    const user = userEvent.setup()
    render(<MessageInput onSend={vi.fn()} onTyping={vi.fn()} />)

    const textarea = screen.getByPlaceholderText('Type a message...')
    await user.type(textarea, 'Hello world')

    const sendButton = screen.getByRole('button')
    expect(sendButton).not.toBeDisabled()
  })

  it('calls onSend when send button is clicked', async () => {
    const user = userEvent.setup()
    const onSend = vi.fn().mockResolvedValue(undefined)
    render(<MessageInput onSend={onSend} onTyping={vi.fn()} />)

    const textarea = screen.getByPlaceholderText('Type a message...')
    await user.type(textarea, 'Hello world')
    await user.click(screen.getByRole('button'))

    expect(onSend).toHaveBeenCalledWith('Hello world')
  })

  it('clears input after sending', async () => {
    const user = userEvent.setup()
    const onSend = vi.fn().mockResolvedValue(undefined)
    render(<MessageInput onSend={onSend} onTyping={vi.fn()} />)

    const textarea = screen.getByPlaceholderText(
      'Type a message...',
    ) as HTMLTextAreaElement
    await user.type(textarea, 'Hello')
    await user.click(screen.getByRole('button'))

    expect(textarea.value).toBe('')
  })

  it('calls onTyping when user types', async () => {
    const user = userEvent.setup()
    const onTyping = vi.fn()
    render(
      <MessageInput onSend={vi.fn()} onTyping={onTyping} />,
    )

    const textarea = screen.getByPlaceholderText('Type a message...')
    await user.type(textarea, 'H')

    expect(onTyping).toHaveBeenCalled()
  })

  it('shows keyboard shortcuts hint', () => {
    render(<MessageInput onSend={vi.fn()} onTyping={vi.fn()} />)
    expect(screen.getByText(/to send/)).toBeInTheDocument()
    expect(screen.getByText(/for new line/)).toBeInTheDocument()
  })
})
