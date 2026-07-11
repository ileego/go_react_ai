import { BASE_URL } from '@/shared/api/client'
import type { TokenResponse, User } from '../types'

export interface AuthApiError extends Error {
  statusCode?: number
  errCode?: string
}

async function authFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...init?.headers,
    },
  })

  const body = await res.json().catch(() => ({}))
  if (!res.ok || body.code !== 0) {
    const err = new Error(body.message || `HTTP ${res.status}`) as AuthApiError
    err.statusCode = res.status
    err.errCode = body.err_code
    throw err
  }

  return body.data as T
}

export function login(email: string, password: string) {
  return authFetch<TokenResponse>('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  })
}

export function registerAccount(email: string, password: string, nickname: string) {
  return authFetch<User>('/auth/register', {
    method: 'POST',
    body: JSON.stringify({ email, password, nickname }),
  })
}

export function logout(accessToken: string | null, refreshToken: string | null) {
  return authFetch<void>('/auth/logout', {
    method: 'POST',
    body: JSON.stringify({
      access_token: accessToken,
      refresh_token: refreshToken,
    }),
  })
}
