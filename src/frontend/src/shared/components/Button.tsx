import type { ButtonHTMLAttributes, ReactNode } from 'react'

interface Props extends ButtonHTMLAttributes<HTMLButtonElement> {
  children: ReactNode
  variant?: 'primary' | 'secondary' | 'outline' | 'danger'
}

export function Button({ children, variant = 'primary', ...rest }: Props) {
  const base =
    'px-4 py-2 rounded font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed'
  const styles = {
    primary:
      'bg-brand-600 text-white hover:bg-brand-700 focus:ring-brand-300 focus:ring-offset-white dark:focus:ring-offset-slate-900',
    secondary:
      'bg-gray-200 text-gray-800 hover:bg-gray-300 focus:ring-gray-300 focus:ring-offset-white dark:bg-slate-700 dark:text-slate-100 dark:hover:bg-slate-600 dark:focus:ring-offset-slate-900',
    outline:
      'border-2 border-gray-300 text-gray-700 hover:bg-gray-100 focus:ring-gray-300 focus:ring-offset-white dark:border-slate-600 dark:text-slate-200 dark:hover:bg-slate-800 dark:focus:ring-offset-slate-900',
    danger:
      'bg-red-600 text-white hover:bg-red-700 focus:ring-red-300 focus:ring-offset-white dark:focus:ring-offset-slate-900',
  }

  return (
    <button className={`${base} ${styles[variant]}`} {...rest}>
      {children}
    </button>
  )
}
