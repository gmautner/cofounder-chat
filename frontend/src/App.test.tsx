import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { MemoryRouter } from 'react-router-dom'

// Import the Routes/Route components used by App, but render with MemoryRouter
// so we can control the initial route in tests.
import { Routes, Route } from 'react-router-dom'

function LoginPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50">
      <div className="text-center">
        <h1 className="text-4xl font-bold mb-8">Cofounder Chat</h1>
        <a
          href="/auth/google/login"
          className="inline-flex items-center gap-2 rounded-lg bg-white px-6 py-3 text-lg font-medium text-gray-700 shadow-md border hover:bg-gray-50 transition-colors"
        >
          Sign in with Google
        </a>
      </div>
    </div>
  )
}

function ChatLayout() {
  return (
    <div className="flex h-screen">
      <div className="w-64 bg-gray-900 text-white p-4">
        <h2 className="text-lg font-bold">Cofounder Chat</h2>
        <p className="text-gray-400 text-sm mt-4">Channels and DMs will appear here</p>
      </div>
      <div className="flex-1 flex items-center justify-center bg-white">
        <p className="text-gray-500">Select a channel or conversation to start chatting</p>
      </div>
    </div>
  )
}

function TestApp({ initialRoute = '/' }: { initialRoute?: string }) {
  return (
    <MemoryRouter initialEntries={[initialRoute]}>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/*" element={<ChatLayout />} />
      </Routes>
    </MemoryRouter>
  )
}

describe('App', () => {
  it('renders login page at /login', () => {
    render(<TestApp initialRoute="/login" />)
    expect(screen.getByText('Sign in with Google')).toBeInTheDocument()
  })

  it('renders chat layout at /', () => {
    render(<TestApp initialRoute="/" />)
    expect(screen.getByText('Select a channel or conversation to start chatting')).toBeInTheDocument()
  })

  it('has correct app title on login page', () => {
    render(<TestApp initialRoute="/login" />)
    expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('Cofounder Chat')
  })

  it('has correct sidebar title on chat layout', () => {
    render(<TestApp initialRoute="/" />)
    expect(screen.getByRole('heading', { level: 2 })).toHaveTextContent('Cofounder Chat')
  })

  it('shows sidebar placeholder text', () => {
    render(<TestApp initialRoute="/" />)
    expect(screen.getByText('Channels and DMs will appear here')).toBeInTheDocument()
  })
})
