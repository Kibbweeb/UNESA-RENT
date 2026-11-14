import { createContext, useCallback, useEffect, useMemo, useState } from 'react'
import { authApi } from '../lib/authApi'
import { clearToken, getToken, saveToken } from './tokenStorage'

export const AuthContext = createContext(null)

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null)
  const [token, setToken] = useState(() => getToken())
  const [status, setStatus] = useState(() => (getToken() ? 'loading' : 'anonymous'))
  const [error, setError] = useState(null)
  const [pending, setPending] = useState(false)

  useEffect(() => {
    if (!getToken()) return undefined

    const controller = new AbortController()

    ;(async () => {
      try {
        const data = await authApi.me(controller.signal)
        setUser(data.user)
        setStatus('authenticated')
      } catch (err) {
        if (controller.signal.aborted) return

        if (err.status === 401) {
          clearToken()
          setToken(null)
        } else {
          setError(err)
        }
        setUser(null)
        setStatus('anonymous')
      }
    })()

    return () => controller.abort()
  }, [])

  const applySession = useCallback((data) => {
    saveToken(data.token)
    setToken(data.token)
    setUser(data.user)
    setStatus('authenticated')
    setError(null)
  }, [])

  const login = useCallback(
    async (credentials) => {
      setPending(true)
      setError(null)
      try {
        const data = await authApi.login(credentials)
        applySession(data)
        return data.user
      } catch (err) {
        setError(err)
        throw err
      } finally {
        setPending(false)
      }
    },
    [applySession],
  )

  const register = useCallback(
    async (payload) => {
      setPending(true)
      setError(null)
      try {
        const data = await authApi.register(payload)
        applySession(data)
        return data.user
      } catch (err) {
        setError(err)
        throw err
      } finally {
        setPending(false)
      }
    },
    [applySession],
  )

  const logout = useCallback(() => {
    clearToken()
    setToken(null)
    setUser(null)
    setError(null)
    setStatus('anonymous')
  }, [])

  const value = useMemo(
    () => ({
      user,
      token,
      status,
      error,
      pending,
      isAuthenticated: status === 'authenticated',
      login,
      register,
      logout,
      clearError: () => setError(null),
    }),
    [user, token, status, error, pending, login, register, logout],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}
