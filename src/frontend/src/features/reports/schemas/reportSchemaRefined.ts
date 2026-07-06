import { z } from 'zod'

export const reportSchemaRefined = z
  .object({
    title: z.string().min(2, '标题至少 2 个字符'),
    topic: z.string().min(1, '请输入研究主题'),
    hasCustomRange: z.boolean(),
    startDate: z.string().optional(),
    endDate: z.string().optional(),
  })
  .refine(
    (data) => {
      if (!data.hasCustomRange) return true
      return data.startDate && data.endDate
    },
    {
      message: '选择自定义时间范围时，必须填写开始和结束日期',
      path: ['startDate'],
    },
  )
  .refine(
    (data) => {
      if (!data.startDate || !data.endDate) return true
      return new Date(data.startDate) < new Date(data.endDate)
    },
    {
      message: '结束日期必须晚于开始日期',
      path: ['endDate'],
    },
  )

export type RefinedReportFormData = z.infer<typeof reportSchemaRefined>
