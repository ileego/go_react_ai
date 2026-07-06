import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { reportSchema, type ReportFormData } from '../schemas/reportSchema'
import { Button } from '@/shared/components/Button'

export function CreateReportFormValidated() {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<ReportFormData>({
    resolver: zodResolver(reportSchema),
    defaultValues: {
      title: '',
      topic: '',
      depth: 'medium',
      model: 'kimi-latest',
      description: '',
    },
  })

  const onSubmit = (data: ReportFormData) => {
    console.log('校验通过:', data)
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      <div>
        <input
          {...register('title')}
          placeholder="报告标题"
          className="w-full rounded border p-2"
        />
        {errors.title && (
          <p className="text-red-500">{errors.title.message}</p>
        )}
      </div>

      <div>
        <input
          {...register('topic')}
          placeholder="研究主题"
          className="w-full rounded border p-2"
        />
        {errors.topic && (
          <p className="text-red-500">{errors.topic.message}</p>
        )}
      </div>

      <Button type="submit">提交</Button>
    </form>
  )
}
