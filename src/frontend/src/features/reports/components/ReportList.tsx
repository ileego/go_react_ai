import { useReports } from '../hooks/useReports'
import { ReportCard } from './ReportCard'

interface Props {
  keyword?: string
}

export function ReportList({ keyword = '' }: Props) {
  const { reports, loading, error } = useReports()

  const filtered = keyword
    ? reports.filter(
        (r) =>
          r.title.toLowerCase().includes(keyword.toLowerCase()) ||
          r.topic.toLowerCase().includes(keyword.toLowerCase())
      )
    : reports

  if (loading) return <div className="text-gray-500">加载中...</div>
  if (error) return <div className="text-red-600">错误: {error}</div>
  if (filtered.length === 0) return <div className="text-gray-500">暂无报告</div>

  return (
    <div className="grid gap-4">
      {filtered.map((r) => (
        <ReportCard key={r.id} report={r} />
      ))}
    </div>
  )
}
