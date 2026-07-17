import { z } from 'zod'

export const reportSchema = z.object({
  title: z.string().min(2, '标题至少需要 2 个字符').max(100, '标题最多 100 个字符'),
  topic: z.string().min(2, '主题至少需要 2 个字符').max(200, '主题最多 200 个字符'),
  depth: z.enum(['shallow', 'medium', 'deep']),
  model: z.enum(['gpt-4o', 'claude-3-5', 'kimi-latest']),
  description: z.string().max(1000, '描述最多 1000 个字符').optional(),
})

export type ReportFormData = z.infer<typeof reportSchema>
