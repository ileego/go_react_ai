import { z } from 'zod'

export const reportSchema = z.object({
  title: z
    .string({ required_error: '报告标题不能为空' })
    .min(2, '标题至少 2 个字符')
    .max(100, '标题最多 100 个字符'),
  topic: z.string({ required_error: '研究主题不能为空' }).min(1, '请输入研究主题'),
  depth: z.enum(['shallow', 'medium', 'deep'], {
    errorMap: () => ({ message: '请选择有效的研究深度' }),
  }),
  model: z.enum(['gpt-4o', 'claude-3-5', 'kimi-latest'], {
    errorMap: () => ({ message: '请选择有效的模型' }),
  }),
  description: z.string().max(2000, '描述最多 2000 个字符').optional(),
})

export type ReportFormData = z.infer<typeof reportSchema>
