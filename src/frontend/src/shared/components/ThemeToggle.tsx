import { useDarkMode } from '@/shared/hooks/useDarkMode'
import { Button } from '@/components/ui/button'
import { Moon, Sun } from 'lucide-react'

export function ThemeToggle() {
  const { isDark, toggle } = useDarkMode()

  return (
    <Button variant="outline" size="icon" onClick={toggle} aria-label="切换主题">
      {isDark ? <Moon className="h-4 w-4" /> : <Sun className="h-4 w-4" />}
    </Button>
  )
}
