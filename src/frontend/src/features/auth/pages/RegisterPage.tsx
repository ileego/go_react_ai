import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'

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
    <Card className="mx-auto my-8 max-w-md">
      <CardHeader>
        <CardTitle>注册</CardTitle>
        <CardDescription>创建新账号以使用全部功能</CardDescription>
      </CardHeader>

      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="nickname">昵称</Label>
            <Input
              id="nickname"
              type="text"
              value={nickname}
              onChange={(e) => setNickname(e.target.value)}
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="email">邮箱</Label>
            <Input
              id="email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="password">密码</Label>
            <Input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
            <p className="text-xs text-muted-foreground">至少 8 位，包含大小写字母和数字</p>
          </div>

          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          <Button type="submit" disabled={isLoading} className="w-full">
            {isLoading ? '注册中...' : '注册'}
          </Button>
        </form>

        <p className="mt-4 text-sm text-muted-foreground">
          已有账号？{' '}
          <Link to="/login" className="text-primary hover:underline">
            去登录
          </Link>
        </p>
      </CardContent>
    </Card>
  )
}
