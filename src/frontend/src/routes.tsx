import { createBrowserRouter } from 'react-router-dom'
import { Root } from './routes/root'
import { HomePage } from './routes/HomePage'
import { DesignSystemPage } from './routes/DesignSystemPage'
import { RouteError } from './routes/RouteError'
import { LoginPage } from '@/features/auth/pages/LoginPage'
import { RegisterPage } from '@/features/auth/pages/RegisterPage'
import { RequireAuth } from '@/features/auth/components/RequireAuth'
import { RequireRole } from '@/features/auth/components/RequireRole'
import { ReportCenterLayout } from '@/features/reports/pages/ReportCenterLayout'
import { ReportsPage } from '@/features/reports/pages/ReportsPage'
import { CreateReportPage } from '@/features/reports/pages/CreateReportPage'
import { ReportDetailPage } from '@/features/reports/pages/ReportDetailPage'

export const router = createBrowserRouter([
  {
    path: '/',
    element: <Root />,
    errorElement: <RouteError />,
    children: [
      { index: true, element: <HomePage /> },
      {
        path: 'design-system',
        element: <DesignSystemPage />,
      },
      {
        path: 'login',
        element: <LoginPage />,
      },
      {
        path: 'register',
        element: <RegisterPage />,
      },
      {
        path: 'reports',
        element: <ReportCenterLayout />,
        children: [
          { index: true, element: <ReportsPage /> },
          {
            path: 'new',
            element: (
              <RequireAuth>
                <CreateReportPage />
              </RequireAuth>
            ),
          },
          {
            path: ':id',
            element: <ReportDetailPage />,
          },
        ],
      },
      {
        path: 'settings',
        element: (
          <RequireRole allowedRoles={['admin']}>
            <div>
              <h1>系统设置</h1>
              <p>只有管理员可见。</p>
            </div>
          </RequireRole>
        ),
      },
    ],
  },
])
