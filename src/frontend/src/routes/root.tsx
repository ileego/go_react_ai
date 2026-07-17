import { Outlet, Link, ScrollRestoration, useLocation } from 'react-router-dom'
import { useEffect } from 'react'
import { useAuth } from '@/features/auth/hooks/useAuth'
import { ThemeToggle } from '@/shared/components/ThemeToggle'
import { Button } from '@/components/ui/button'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
  SheetClose,
} from '@/components/ui/sheet'
import { Menu } from 'lucide-react'

const navItems = [
  { to: '/', label: '首页' },
  { to: '/reports', label: '报告列表' },
  { to: '/reports/new', label: '新建报告' },
  { to: '/design-system', label: '设计系统' },
]

export function Root() {
  const { pathname } = useLocation()
  const { user, logout } = useAuth()

  useEffect(() => {
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }, [pathname])

  return (
    <div className="min-h-screen bg-background text-foreground">
      <ScrollRestoration />
      <header className="border-b p-4">
        <nav className="flex items-center justify-between gap-6">
          <div className="hidden items-center gap-6 md:flex">
            {navItems.map((item) => (
              <Link
                key={item.to}
                to={item.to}
                className="text-sm font-medium transition-colors hover:text-primary"
              >
                {item.label}
              </Link>
            ))}
          </div>

          <div className="flex items-center gap-4 md:hidden">
            <Sheet>
              <SheetTrigger asChild>
                <Button variant="outline" size="icon" aria-label="打开导航">
                  <Menu className="h-4 w-4" />
                </Button>
              </SheetTrigger>
              <SheetContent side="left" className="w-[240px]">
                <SheetHeader>
                  <SheetTitle>导航</SheetTitle>
                </SheetHeader>
                <div className="mt-6 flex flex-col gap-2">
                  {navItems.map((item) => (
                    <SheetClose key={item.to} asChild>
                      <Link
                        to={item.to}
                        className="rounded-md px-3 py-2 text-sm font-medium transition-colors hover:bg-accent"
                      >
                        {item.label}
                      </Link>
                    </SheetClose>
                  ))}
                </div>
              </SheetContent>
            </Sheet>
          </div>

          <div className="flex items-center gap-4">
            <ThemeToggle />
            {user ? (
              <div className="flex items-center gap-3">
                <span className="hidden text-sm md:inline">{user.nickname}</span>
                <Button variant="outline" onClick={() => logout()}>
                  退出
                </Button>
              </div>
            ) : (
              <Button variant="outline" asChild>
                <Link to="/login">登录</Link>
              </Button>
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
