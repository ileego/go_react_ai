import { useCallback, useEffect, useState } from 'react'
import { fetchJson } from '@/shared/api/client'
import type { CreateReportRequest, Report } from '../types'

interface ListResponse {
  list: Report[]
  total: number
  page: number
  page_size: number
}

export function useReports() {
  const [reports, setReports] = useState<Report[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const fetchReports = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const data = await fetchJson<ListResponse>('/reports')
      setReports(data.list || [])
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      setLoading(false)
    }
  }, [])

  const createReport = useCallback(async (req: CreateReportRequest) => {
    const report = await fetchJson<Report>('/reports', {
      method: 'POST',
      body: JSON.stringify(req),
    })
    setReports((prev) => [report, ...prev])
    return report
  }, [])

  useEffect(() => {
    fetchReports()
  }, [fetchReports])

  return { reports, loading, error, fetchReports, createReport }
}
