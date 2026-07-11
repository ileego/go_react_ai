// Feature: Reports
// 所有与报告相关的组件、钩子、API 调用、类型定义都集中在这里。
// 避免在全局 components/ hooks/ 目录中分散管理。

export { ReportList } from './components/ReportList'
export { ReportCard } from './components/ReportCard'
export { useReports } from './hooks/useReports'
export type { Report, ReportStatus } from './types'

// 第17章路由与导航教学示例
export { ReportsPageWithLoader } from './pages/ReportsPageWithLoader'
export { ReportCenterLayout } from './pages/ReportCenterLayout'
export { ReportsPage } from './pages/ReportsPage'
export { CreateReportPage } from './pages/CreateReportPage'
export { ReportDetailPage } from './pages/ReportDetailPage'
export { ReportDetail } from './components/ReportDetail'
export { ReportSearch } from './components/ReportSearch'

// 第16章表单与校验教学示例
export { UncontrolledTopicInput } from './components/UncontrolledTopicInput'
export { ActionForm } from './components/ActionForm'
export { BasicRHFReportForm } from './components/BasicRHFReportForm'
export { WatchAndController } from './components/WatchAndController'
export { CreateReportForm as CreateReportFormRHF } from './components/CreateReportFormRHF'
export { CreateReportFormValidated } from './components/CreateReportFormValidated'
export { ResearcherFieldArray } from './components/ResearcherFieldArray'
export { CreateResearchTaskForm } from './components/CreateResearchTaskForm'
export { AttachmentUploader } from './components/AttachmentUploader'
export { ReportFormWithUpload } from './components/ReportFormWithUpload'
export { FieldErrorMessage } from './components/FieldError'
export { AutoSaveDraftForm } from './components/AutoSaveDraftForm'
export { SubmitWithMutation } from './components/SubmitWithMutation'
export { useDebouncedSubmit } from './hooks/useDebouncedSubmit'
export { reportSchema } from './schemas/reportSchema'
export { reportSchemaRefined } from './schemas/reportSchemaRefined'
export { validateFiles } from './utils/fileValidation'
export { mapServerErrors } from './utils/mapServerErrors'
export type { ReportFormData } from './schemas/reportSchema'
export type { RefinedReportFormData } from './schemas/reportSchemaRefined'
