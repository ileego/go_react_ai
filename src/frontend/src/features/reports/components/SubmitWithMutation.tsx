import { useForm } from 'react-hook-form'
import { useMutation } from '@tanstack/react-query'
import { useDebouncedSubmit } from '../hooks/useDebouncedSubmit'
import type { ReportFormData } from '../schemas/reportSchema'

async function createReport(data: ReportFormData) {
  const res = await fetch('/api/reports', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  if (!res.ok) throw new Error('创建失败')
  return res.json()
}

export function SubmitWithMutation() {
  const { register, handleSubmit } = useForm<ReportFormData>()
  const mutation = useMutation({ mutationFn: createReport })
  const onSubmit = useDebouncedSubmit(
    async (data: ReportFormData) => mutation.mutateAsync(data),
    800,
  )

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-2">
      <input
        {...register('title')}
        placeholder="报告标题"
        className="w-full rounded border p-2"
      />
      <button
        type="submit"
        disabled={mutation.isPending}
        className="rounded bg-blue-600 px-4 py-2 text-white disabled:opacity-50"
      >
        {mutation.isPending ? '提交中...' : '提交'}
      </button>
      {mutation.isError && (
        <p className="text-red-500">创建失败，请重试</p>
      )}
    </form>
  )
}
