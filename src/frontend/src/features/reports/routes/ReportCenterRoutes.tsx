import { Routes, Route } from 'react-router-dom'
import { ReportCenterLayout } from '../pages/ReportCenterLayout'
import { ReportsPage } from '../pages/ReportsPage'
import { CreateReportPage } from '../pages/CreateReportPage'
import { ReportDetailPage } from '../pages/ReportDetailPage'

export default function ReportCenterRoutes() {
  return (
    <Routes>
      <Route element={<ReportCenterLayout />}>
        <Route index element={<ReportsPage />} />
        <Route path="new" element={<CreateReportPage />} />
        <Route path=":id" element={<ReportDetailPage />} />
      </Route>
    </Routes>
  )
}
