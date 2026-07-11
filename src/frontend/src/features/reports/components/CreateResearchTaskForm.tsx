import { useForm, useFieldArray, Controller } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Button } from '@/shared/components/Button'

const taskSchema = z
  .object({
    title: z.string().min(2, '标题至少 2 个字符').max(100, '标题最多 100 个字符'),
    topic: z.string().min(1, '请输入研究主题'),
    mode: z.enum(['quick', 'standard', 'deep'], {
      errorMap: () => ({ message: '请选择研究模式' }),
    }),
    hasCustomRange: z.boolean(),
    startDate: z.string().optional(),
    endDate: z.string().optional(),
    researchers: z
      .array(
        z.object({
          name: z.string().min(1, '姓名不能为空'),
          role: z.enum(['lead', 'analyst', 'writer'], {
            errorMap: () => ({ message: '请选择角色' }),
          }),
        })
      )
      .min(1, '至少需要一名研究员'),
    model: z.enum(['gpt-4o', 'claude-3-5', 'kimi-latest'], {
      errorMap: () => ({ message: '请选择模型' }),
    }),
  })
  .refine(
    (data) => {
      if (!data.hasCustomRange) return true
      return !!data.startDate && !!data.endDate
    },
    { message: '请填写完整的时间范围', path: ['startDate'] }
  )
  .refine(
    (data) => {
      if (!data.startDate || !data.endDate) return true
      return new Date(data.startDate) < new Date(data.endDate)
    },
    { message: '结束日期必须晚于开始日期', path: ['endDate'] }
  )

export type CreateTaskFormData = z.infer<typeof taskSchema>

interface CreateResearchTaskFormProps {
  onSubmit: (data: CreateTaskFormData) => void
  isLoading?: boolean
}

export function CreateResearchTaskForm({ onSubmit, isLoading }: CreateResearchTaskFormProps) {
  const {
    register,
    control,
    handleSubmit,
    watch,
    formState: { errors },
  } = useForm<CreateTaskFormData>({
    resolver: zodResolver(taskSchema),
    defaultValues: {
      title: '',
      topic: '',
      mode: 'standard',
      hasCustomRange: false,
      researchers: [{ name: '', role: 'lead' }],
      model: 'kimi-latest',
    },
  })

  const { fields, append, remove } = useFieldArray({
    control,
    name: 'researchers',
  })

  const hasCustomRange = watch('hasCustomRange')

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      <div>
        <label htmlFor="title">任务标题</label>
        <input id="title" {...register('title')} className="w-full rounded border p-2" />
        {errors.title && <p className="text-sm text-red-500">{errors.title.message}</p>}
      </div>

      <div>
        <label htmlFor="topic">研究主题</label>
        <input id="topic" {...register('topic')} className="w-full rounded border p-2" />
        {errors.topic && <p className="text-sm text-red-500">{errors.topic.message}</p>}
      </div>

      <div>
        <label htmlFor="mode">研究模式</label>
        <select id="mode" {...register('mode')} className="w-full rounded border p-2">
          <option value="quick">快速调研</option>
          <option value="standard">标准研究</option>
          <option value="deep">深度研究</option>
        </select>
        {errors.mode && <p className="text-sm text-red-500">{errors.mode.message}</p>}
      </div>

      <div>
        <label className="flex items-center gap-2">
          <input type="checkbox" {...register('hasCustomRange')} />
          自定义时间范围
        </label>
        {hasCustomRange && (
          <div className="mt-2 flex gap-2">
            <input type="date" {...register('startDate')} className="rounded border p-2" />
            <input type="date" {...register('endDate')} className="rounded border p-2" />
          </div>
        )}
        {errors.startDate && <p className="text-sm text-red-500">{errors.startDate.message}</p>}
        {errors.endDate && <p className="text-sm text-red-500">{errors.endDate.message}</p>}
      </div>

      <div>
        <h4 className="mb-2 font-medium">研究员</h4>
        {fields.map((field, index) => (
          <div key={field.id} className="mb-2 flex gap-2">
            <input
              {...register(`researchers.${index}.name`)}
              placeholder="姓名"
              className="rounded border p-2"
            />
            <select {...register(`researchers.${index}.role`)} className="rounded border p-2">
              <option value="lead">负责人</option>
              <option value="analyst">分析师</option>
              <option value="writer">写手</option>
            </select>
            <Button type="button" variant="outline" onClick={() => remove(index)}>
              删除
            </Button>
          </div>
        ))}
        <Button
          type="button"
          variant="secondary"
          onClick={() => append({ name: '', role: 'analyst' })}
        >
          添加研究员
        </Button>
        {errors.researchers?.root && (
          <p className="text-sm text-red-500">{errors.researchers.root.message}</p>
        )}
      </div>

      <div>
        <label htmlFor="model">默认模型</label>
        <Controller
          name="model"
          control={control}
          render={({ field }) => (
            <select
              id="model"
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
        {errors.model && <p className="text-sm text-red-500">{errors.model.message}</p>}
      </div>

      <Button type="submit" disabled={isLoading}>
        {isLoading ? '创建中...' : '创建研究任务'}
      </Button>
    </form>
  )
}
