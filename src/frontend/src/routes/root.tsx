import { Outlet, Link, ScrollRestoration, useLocation } from 'react-router-dom'
import { useEffect } from 'react'
import { useAuth } from '@/features/auth/hooks/useAuth'
import { ThemeToggle } from '@/shared/components/ThemeToggle'
import { Button } from '@/shared/components/Button'

export function Root() {
  const { pathname } = useLocation()
  const { user, logout } = useAuth()

  useEffect(() => {
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }, [pathname])

  return (
    <div className="min-h-screen bg-white text-gray-900 dark:bg-slate-900 dark:text-slate-100">
      <ScrollRestoration />
      <header className="border-b border-gray-200 p-4 dark:border-slate-700">
        <nav className="flex items-center justify-between gap-6">
          <div className="flex gap-6">
            <Link to="/" className="hover:text-brand-600 dark:hover:text-brand-400">
              首页
            </Link>
            <Link to="/reports" className="hover:text-brand-600 dark:hover:text-brand-400">
              报告列表
            </Link>
            <Link to="/design-system" className="hover:text-brand-600 dark:hover:text-brand-400">
              设计系统
            </Link>
            <Link to="/reports/new" className="hover:text-brand-600 dark:hover:text-brand-400">
              新建报告
            </Link>
          </div>

          <div className="flex items-center gap-4">
            <ThemeToggle />
            {user ? (
              <div className="flex items-center gap-3">
                <span className="text-sm">{user.nickname}</span>
                <Button variant="outline" onClick={() => logout()}>
                  退出
                </Button>
              </div>
            ) : (
              <Link to="/login" className="text-sm hover:text-brand-600">
                登录
              </Link>
            )}
          </div>
        </nav>
      </header>

      <main key={pathname} className="animate-fadeIn p-6">
        <Outlet />
      </main>
    </div>
  )
}
