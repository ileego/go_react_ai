import { useReports } from '../hooks/useReports'
import { ReportCard } from './ReportCard'
import { Skeleton } from '@/components/ui/skeleton'

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

  if (loading) {
    return (
      <div className="grid gap-4">
        {Array.from({ length: 3 }).map((_, i) => (
          <div key={i} className="rounded-lg border p-4">
            <Skeleton className="h-5 w-1/3" />
            <Skeleton className="mt-2 h-4 w-1/2" />
            <Skeleton className="mt-3 h-6 w-16 rounded-full" />
          </div>
        ))}
      </div>
    )
  }

  if (error) return <div className="text-destructive">错误: {error}</div>
  if (filtered.length === 0) return <div className="text-muted-foreground">暂无报告</div>

  return (
    <div className="grid gap-4">
      {filtered.map((r) => (
        <ReportCard key={r.id} report={r} />
      ))}
    </div>
  )
}
