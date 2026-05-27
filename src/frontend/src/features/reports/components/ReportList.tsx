import { useReports } from '../hooks/useReports'
import { ReportCard } from './ReportCard'

export function ReportList() {
  const { reports, loading, error } = useReports()

  if (loading) return <div>加载中...</div>
  if (error) return <div style={{ color: 'red' }}>错误: {error}</div>
  if (reports.length === 0) return <div>暂无报告</div>

  return (
    <div style={{ display: 'grid', gap: '1rem' }}>
      {reports.map((r) => (
        <ReportCard key={r.id} report={r} />
      ))}
    </div>
  )
}
