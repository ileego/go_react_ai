// Shared API Client
// 所有 feature 的 API 调用都基于这个客户端，统一处理基础 URL、错误、认证令牌注入与刷新。

import { useAuthStore } from '@/features/auth/store/authStore'
import type { TokenResponse } from '@/features/auth/types'

export const BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api'

export class ApiError extends Error {
  constructor(
    message: string,
    public statusCode: number,
    public errCode?: string
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

let refreshPromise: Promise<boolean> | null = null

async function doRefresh(): Promise<boolean> {
  const refreshToken = useAuthStore.getState().refreshToken
  if (!refreshToken) return false

  try {
    const res = await fetch(`${BASE_URL}/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken }),
    })
    const body = await res.json().catch(() => ({}))
    if (!res.ok || body.code !== 0) return false

    useAuthStore.getState().setTokens(body.data as TokenResponse)
    return true
  } catch {
    return false
  }
}

async function refreshWithLock(): Promise<boolean> {
  if (refreshPromise) return refreshPromise

  refreshPromise = doRefresh().finally(() => {
    refreshPromise = null
  })
  return refreshPromise
}

export async function fetchJson<T>(path: string, init?: RequestInit): Promise<T> {
  const makeRequest = (token: string | null) =>
    fetch(`${BASE_URL}${path}`, {
      ...init,
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...init?.headers,
      },
    })

  let res = await makeRequest(useAuthStore.getState().accessToken)

  if (res.status === 401 && useAuthStore.getState().accessToken) {
    const ok = await refreshWithLock()
    if (ok) {
      res = await makeRequest(useAuthStore.getState().accessToken)
    } else {
      useAuthStore.getState().clearAuth()
    }
  }

  const body = await res.json().catch(() => ({}))
  if (!res.ok || body.code !== 0) {
    throw new ApiError(body.message || `HTTP ${res.status}`, res.status, body.err_code)
  }

  return body.data as T
}
