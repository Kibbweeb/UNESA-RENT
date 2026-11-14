import { useAuth } from './useAuth'

export function RequireAuth({ children, fallback = null }) {
  const { status } = useAuth()

  if (status === 'loading') {
    return <p>Memeriksa sesi...</p>
  }

  if (status !== 'authenticated') {
    return fallback
  }

  return children
}
