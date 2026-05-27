import type { Report } from '../types'

interface Props {
  report: Report
}

const statusMap: Record<string, string> = {
  pending: '待处理',
  running: '执行中',
  completed: '已完成',
  failed: '失败',
}

export function ReportCard({ report }: Props) {
  return (
    <div
      style={{
        padding: '1rem',
        border: '1px solid #e5e7eb',
        borderRadius: '8px',
      }}
    >
      <h3 style={{ margin: '0 0 0.5rem' }}>{report.title}</h3>
      <p style={{ margin: '0 0 0.5rem', color: '#6b7280' }}>{report.topic}</p>
      <span
        style={{
          display: 'inline-block',
          padding: '0.25rem 0.75rem',
          borderRadius: '9999px',
          fontSize: '0.875rem',
          background: report.status === 'completed' ? '#dcfce7' : '#f3f4f6',
          color: report.status === 'completed' ? '#166534' : '#374151',
        }}
      >
        {statusMap[report.status] || report.status}
      </span>
    </div>
  )
}
