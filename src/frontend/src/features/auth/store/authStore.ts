import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { TokenResponse, User } from '../types'

interface AuthState {
  accessToken: string | null
  refreshToken: string | null
  user: User | null
  isLoading: boolean
  hydrated: boolean
  setTokens: (t: TokenResponse) => void
  setUser: (u: User | null) => void
  setLoading: (v: boolean) => void
  setHydrated: () => void
  clearAuth: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      accessToken: null,
      refreshToken: null,
      user: null,
      isLoading: false,
      hydrated: false,
      setTokens: (t) =>
        set({
          accessToken: t.access_token,
          refreshToken: t.refresh_token,
        }),
      setUser: (u) => set({ user: u }),
      setLoading: (v) => set({ isLoading: v }),
      setHydrated: () => set({ hydrated: true }),
      clearAuth: () =>
        set({
          accessToken: null,
          refreshToken: null,
          user: null,
        }),
    }),
    {
      name: 'auth-storage',
      partialize: (s) => ({
        accessToken: s.accessToken,
        refreshToken: s.refreshToken,
      }),
      onRehydrateStorage: () => (state, error) => {
        if (!error) state?.setHydrated()
      },
    }
  )
)
