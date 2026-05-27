// Shared API Client
// 所有 feature 的 API 调用都基于这个客户端，统一处理基础 URL、错误、超时等。

const BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api'

export async function fetchJson<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...init?.headers,
    },
  })

  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.message || `HTTP ${res.status}`)
  }

  return res.json() as Promise<T>
}
