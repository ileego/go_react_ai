import { useLoaderData } from 'react-router-dom'
import type { Report } from '../types'

export function ReportsPageWithLoader() {
  const reports = useLoaderData() as Report[]

  return (
    <div>
      <h1>报告列表（loader 预取）</h1>
      <ul>
        {reports.map((report) => (
          <li key={report.id}>{report.title}</li>
        ))}
      </ul>
    </div>
  )
}
