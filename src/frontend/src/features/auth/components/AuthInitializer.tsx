import { useEffect } from 'react'
import { useAuthStore } from '../store/authStore'
import { fetchJson } from '@/shared/api/client'
import type { User } from '../types'

interface Props {
  children: React.ReactNode
}

export function AuthInitializer({ children }: Props) {
  const hydrated = useAuthStore((s) => s.hydrated)
  const setUser = useAuthStore((s) => s.setUser)
  const clearAuth = useAuthStore((s) => s.clearAuth)

  useEffect(() => {
    if (!hydrated) return
    if (!useAuthStore.getState().accessToken) return

    fetchJson<User>('/auth/me')
      .then((user) => setUser(user))
      .catch(() => clearAuth())
  }, [hydrated, setUser, clearAuth])

  return <>{children}</>
}
