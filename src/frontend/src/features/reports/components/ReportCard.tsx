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

const statusClasses: Record<string, string> = {
  completed: 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-200',
  failed: 'bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-200',
  running: 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-200',
  pending: 'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300',
}

export function ReportCard({ report }: Props) {
  return (
    <div className="rounded-2xl border border-gray-200 bg-white p-4 shadow-card dark:border-slate-700 dark:bg-slate-800">
      <h3 className="text-lg font-semibold">{report.title}</h3>
      <p className="mt-1 text-gray-600 dark:text-gray-300">{report.topic}</p>
      <span
        className={`mt-2 inline-block rounded-full px-3 py-1 text-sm ${
          statusClasses[report.status] || statusClasses.pending
        }`}
      >
        {statusMap[report.status] || report.status}
      </span>
    </div>
  )
}
