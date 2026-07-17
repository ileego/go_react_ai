import type { ButtonHTMLAttributes, ReactNode } from 'react'
import { Button as ShadcnButton, buttonVariants } from '@/components/ui/button'
import type { VariantProps } from 'class-variance-authority'

interface Props extends ButtonHTMLAttributes<HTMLButtonElement> {
  children: ReactNode
  variant?: 'primary' | 'secondary' | 'outline' | 'danger' | VariantProps<typeof buttonVariants>['variant']
  size?: VariantProps<typeof buttonVariants>['size']
}

const variantMap: Record<string, VariantProps<typeof buttonVariants>['variant']> = {
  primary: 'default',
  secondary: 'secondary',
  outline: 'outline',
  danger: 'destructive',
}

export function Button({ children, variant = 'primary', size, ...rest }: Props) {
  const shadcnVariant = variant && variantMap[variant] ? variantMap[variant] : (variant as VariantProps<typeof buttonVariants>['variant'])

  return (
    <ShadcnButton variant={shadcnVariant} size={size} {...rest}>
      {children}
    </ShadcnButton>
  )
}
