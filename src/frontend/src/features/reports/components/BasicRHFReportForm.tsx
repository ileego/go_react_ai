import { useForm } from 'react-hook-form'
import { Button } from '@/shared/components/Button'

interface ReportFormValues {
  title: string
  topic: string
}

export function BasicRHFReportForm() {
  const { register, handleSubmit, reset } = useForm<ReportFormValues>({
    defaultValues: { title: '未命名报告', topic: '' },
  })

  const onSubmit = (data: ReportFormValues) => {
    console.log('表单数据:', data)
    reset()
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-2">
      <input {...register('title')} placeholder="报告标题" className="w-full rounded border p-2" />
      <input {...register('topic')} placeholder="研究主题" className="w-full rounded border p-2" />
      <Button type="submit">提交</Button>
    </form>
  )
}
