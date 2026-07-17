import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { CreateReportForm } from '../components/CreateReportFormRHF'
import { useReports } from '../hooks/useReports'
import type { CreateReportFormValues } from '../components/CreateReportFormRHF'
import { Alert, AlertDescription } from '@/components/ui/alert'

export function CreateReportPage() {
  const navigate = useNavigate()
  const { createReport } = useReports()
  const [error, setError] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  const handleSubmit = async (data: CreateReportFormValues) => {
    setError(null)
    setIsLoading(true)
    try {
      const report = await createReport({
        title: data.title,
        topic: data.topic,
      })
      navigate(`/reports/${report.id}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : '创建失败')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="mx-auto max-w-2xl space-y-4">
      <h1 className="text-2xl font-bold">新建报告</h1>
      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}
      <CreateReportForm onSubmit={handleSubmit} isLoading={isLoading} />
    </div>
  )
}
