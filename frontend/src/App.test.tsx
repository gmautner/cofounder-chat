import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { MemoryRouter } from 'react-router-dom'
import { Routes, Route } from 'react-router-dom'
import { LoginPage } from '@/pages/LoginPage'

// Minimal test wrapper — renders just the login page without auth checks
// (since auth requires a running backend and SSE connection)
function TestLoginPage({ initialRoute = '/login' }: { initialRoute?: string }) {
  return (
    <MemoryRouter initialEntries={[initialRoute]}>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
      </Routes>
    </MemoryRouter>
  )
}

describe('LoginPage', () => {
  it('renders the app title', () => {
    render(<TestLoginPage />)
    expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent(
      'Cofounder Chat',
    )
  })

  it('renders Entrar com Google link', () => {
    render(<TestLoginPage />)
    expect(screen.getByText('Entrar com Google')).toBeInTheDocument()
  })

  it('links to Google OAuth endpoint', () => {
    render(<TestLoginPage />)
    const link = screen.getByText('Entrar com Google').closest('a')
    expect(link).toHaveAttribute('href', '/auth/google/login')
  })

  it('renders the subtitle text', () => {
    render(<TestLoginPage />)
    expect(
      screen.getByText('Mensagens em tempo real para seu time'),
    ).toBeInTheDocument()
  })

  it('shows dev login button in development mode', () => {
    render(<TestLoginPage />)
    // In test env, import.meta.env.DEV is true
    expect(screen.getByText('Login de teste')).toBeInTheDocument()
  })
})
