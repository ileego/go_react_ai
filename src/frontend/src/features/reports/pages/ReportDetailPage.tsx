import { useParams, Link } from 'react-router-dom'
import { ReportDetail } from '../components/ReportDetail'

export function ReportDetailPage() {
  const { id } = useParams<{ id: string }>()

  if (!id) {
    return <div>缺少报告 ID</div>
  }

  return (
    <div>
      <Link to="/reports">← 返回列表</Link>
      <h1 style={{ marginTop: '1rem' }}>报告详情</h1>
      <ReportDetail reportId={id} />
    </div>
  )
}
