import { Navigate, useLocation } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'

interface Props {
  allowedRoles: string[]
  children: React.ReactNode
}

export function RequireRole({ allowedRoles, children }: Props) {
  const { user } = useAuth()
  const location = useLocation()

  if (!user) {
    return <Navigate to="/login" state={{ from: location.pathname }} replace />
  }

  if (!allowedRoles.includes(user.role)) {
    return <div>权限不足</div>
  }

  return <>{children}</>
}
