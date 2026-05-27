export type ReportStatus = 'pending' | 'running' | 'completed' | 'failed'

export interface Report {
  id: number
  title: string
  topic: string
  status: ReportStatus
  content?: string
  sources: string[]
  createdBy: number
  createdAt: string
  updatedAt: string
}

export interface CreateReportRequest {
  title: string
  topic: string
}
