import type { Report } from '../types'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'

interface Props {
  report: Report
}

const statusMap: Record<string, string> = {
  pending: '待处理',
  running: '执行中',
  completed: '已完成',
  failed: '失败',
}

const statusVariant: Record<
  string,
  'default' | 'secondary' | 'destructive' | 'outline'
> = {
  completed: 'default',
  failed: 'destructive',
  running: 'secondary',
  pending: 'outline',
}

export function ReportCard({ report }: Props) {
  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle>{report.title}</CardTitle>
        <CardDescription>{report.topic}</CardDescription>
      </CardHeader>
      <CardContent>
        <Badge variant={statusVariant[report.status] || 'outline'}>
          {statusMap[report.status] || report.status}
        </Badge>
      </CardContent>
    </Card>
  )
}
