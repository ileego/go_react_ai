import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'
import { Button } from '@/shared/components/Button'

export function RegisterPage() {
  const navigate = useNavigate()
  const { registerAccount, isLoading } = useAuth()

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [nickname, setNickname] = useState('')
  const [error, setError] = useState<string | null>(null)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)

    try {
      await registerAccount(email, password, nickname)
      navigate('/login', { state: { message: '注册成功，请登录' } })
    } catch (err) {
      setError(err instanceof Error ? err.message : '注册失败')
    }
  }

  return (
    <div className="mx-auto my-8 max-w-md rounded-2xl border border-gray-200 bg-white p-6 shadow-card dark:border-slate-700 dark:bg-slate-800">
      <h1 className="text-xl font-bold">注册</h1>

      <form onSubmit={handleSubmit} className="mt-6 space-y-4">
        <div>
          <label htmlFor="nickname">昵称</label>
          <input
            id="nickname"
            type="text"
            value={nickname}
            onChange={(e) => setNickname(e.target.value)}
            className="w-full rounded border border-gray-300 p-2 dark:border-slate-600 dark:bg-slate-800"
            required
          />
        </div>

        <div>
          <label htmlFor="email">邮箱</label>
          <input
            id="email"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="w-full rounded border border-gray-300 p-2 dark:border-slate-600 dark:bg-slate-800"
            required
          />
        </div>

        <div>
          <label htmlFor="password">密码</label>
          <input
            id="password"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="w-full rounded border border-gray-300 p-2 dark:border-slate-600 dark:bg-slate-800"
            required
          />
          <p className="mt-1 text-xs text-gray-500">至少 8 位，包含大小写字母和数字</p>
        </div>

        {error && <p className="text-sm text-red-600">{error}</p>}

        <Button type="submit" disabled={isLoading}>
          {isLoading ? '注册中...' : '注册'}
        </Button>
      </form>

      <p className="mt-4 text-sm">
        已有账号？{' '}
        <Link to="/login" className="text-brand-600 hover:underline">
          去登录
        </Link>
      </p>
    </div>
  )
}
