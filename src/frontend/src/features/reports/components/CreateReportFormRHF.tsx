import { useForm, Controller } from 'react-hook-form'
import { Button } from '@/shared/components/Button'

export interface CreateReportFormValues {
  title: string
  topic: string
  depth: 'shallow' | 'medium' | 'deep'
  model: 'gpt-4o' | 'claude-3-5' | 'kimi-latest'
  description: string
}

interface CreateReportFormProps {
  onSubmit: (data: CreateReportFormValues) => void
  isLoading?: boolean
}

export function CreateReportForm({ onSubmit, isLoading }: CreateReportFormProps) {
  const { register, handleSubmit, control, formState } = useForm<CreateReportFormValues>({
    defaultValues: {
      title: '',
      topic: '',
      depth: 'medium',
      model: 'kimi-latest',
      description: '',
    },
  })

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      <div>
        <label htmlFor="title">报告标题</label>
        <input
          id="title"
          {...register('title')}
          placeholder="例如：多智能体协作调研"
          className="w-full rounded border p-2"
        />
        {formState.errors.title && (
          <p className="text-sm text-red-500">{formState.errors.title.message}</p>
        )}
      </div>

      <div>
        <label htmlFor="topic">研究主题</label>
        <input
          id="topic"
          {...register('topic')}
          placeholder="输入核心研究问题"
          className="w-full rounded border p-2"
        />
      </div>

      <div>
        <label htmlFor="depth">研究深度</label>
        <select id="depth" {...register('depth')} className="w-full rounded border p-2">
          <option value="shallow">浅度概览</option>
          <option value="medium">中度分析</option>
          <option value="deep">深度研究</option>
        </select>
      </div>

      <div>
        <label>默认模型</label>
        <Controller
          name="model"
          control={control}
          render={({ field }) => (
            <select
              value={field.value}
              onChange={field.onChange}
              className="w-full rounded border p-2"
            >
              <option value="gpt-4o">GPT-4o</option>
              <option value="claude-3-5">Claude 3.5</option>
              <option value="kimi-latest">Kimi Latest</option>
            </select>
          )}
        />
      </div>

      <div>
        <label htmlFor="description">补充描述</label>
        <textarea
          id="description"
          {...register('description')}
          placeholder="描述报告目标、预期读者、参考资料等"
          className="w-full rounded border p-2"
          rows={4}
        />
      </div>

      <Button type="submit" disabled={isLoading}>
        {isLoading ? '创建中...' : '创建报告'}
      </Button>
    </form>
  )
}
