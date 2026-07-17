import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { RouterProvider } from 'react-router-dom'
import { AuthInitializer } from '@/features/auth/components/AuthInitializer'
import { Toaster } from '@/components/ui/sonner'
import { router } from './routes'
import './index.css'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <AuthInitializer>
      <RouterProvider router={router} />
      <Toaster />
    </AuthInitializer>
  </StrictMode>
)
