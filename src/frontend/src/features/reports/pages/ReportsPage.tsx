import { useSearchParams } from 'react-router-dom'
import { ReportSearch } from '../components/ReportSearch'
import { ReportList } from '../components/ReportList'

export function ReportsPage() {
  const [searchParams] = useSearchParams()
  const keyword = searchParams.get('keyword') || ''

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-bold">报告列表</h1>
      <ReportSearch />
      <ReportList keyword={keyword} />
    </div>
  )
}
