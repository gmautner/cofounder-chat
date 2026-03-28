import { BrowserRouter, Routes, Route } from 'react-router-dom'

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

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/*" element={<ChatLayout />} />
      </Routes>
    </BrowserRouter>
  )
}
