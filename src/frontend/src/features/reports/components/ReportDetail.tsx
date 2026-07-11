import { useEffect, useState } from 'react'
import { fetchJson } from '@/shared/api/client'
import type { Report } from '../types'

interface Props {
  reportId: string
}

export function ReportDetail({ reportId }: Props) {
  const [report, setReport] = useState<Report | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    fetchJson<Report>(`/reports/${reportId}`)
      .then((data) => setReport(data))
      .catch((err) => setError(err instanceof Error ? err.message : '未知错误'))
      .finally(() => setLoading(false))
  }, [reportId])

  if (loading) return <div className="text-gray-500">加载中...</div>
  if (error) return <div className="text-red-600">错误: {error}</div>
  if (!report) return <div>报告不存在</div>

  return (
    <div className="space-y-2">
      <h2 className="text-2xl font-bold">{report.title}</h2>
      <p className="text-gray-600 dark:text-gray-300">主题: {report.topic}</p>
      <p className="text-gray-600 dark:text-gray-300">状态: {report.status}</p>
      <p className="text-sm text-gray-500">创建时间: {report.createdAt}</p>
    </div>
  )
}
