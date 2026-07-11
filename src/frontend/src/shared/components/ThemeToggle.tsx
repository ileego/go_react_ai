import { useDarkMode } from '@/shared/hooks/useDarkMode'
import { Button } from './Button'

export function ThemeToggle() {
  const { isDark, toggle } = useDarkMode()

  return (
    <Button variant="outline" onClick={toggle} aria-label="切换主题">
      {isDark ? '深色' : '浅色'}
    </Button>
  )
}
