export interface TokenResponse {
  access_token: string
  refresh_token: string
  expires_in: number
}

export interface User {
  id: number
  email: string
  nickname: string
  avatar_url: string
  role: string
}
