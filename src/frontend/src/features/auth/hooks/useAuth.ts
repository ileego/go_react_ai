import { useAuthStore } from '../store/authStore'
import { login as loginApi, registerAccount, logout as logoutApi } from '../api/authApi'
import { fetchJson } from '@/shared/api/client'
import type { User } from '../types'

export function useAuth() {
  const user = useAuthStore((s) => s.user)
  const isLoading = useAuthStore((s) => s.isLoading)
  const setTokens = useAuthStore((s) => s.setTokens)
  const setUser = useAuthStore((s) => s.setUser)
  const setLoading = useAuthStore((s) => s.setLoading)
  const clearAuth = useAuthStore((s) => s.clearAuth)

  const login = async (email: string, password: string) => {
    setLoading(true)
    try {
      const tokens = await loginApi(email, password)
      setTokens(tokens)
      const me = await fetchJson<User>('/auth/me')
      setUser(me)
    } finally {
      setLoading(false)
    }
  }

  const logout = async () => {
    const { accessToken, refreshToken } = useAuthStore.getState()
    await logoutApi(accessToken, refreshToken).catch(() => {})
    clearAuth()
  }

  return {
    user,
    isAuthenticated: !!user,
    isLoading,
    login,
    registerAccount,
    logout,
  }
}
