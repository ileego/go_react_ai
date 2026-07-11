import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { debounce } from '@/shared/utils/debounce'

interface DraftValues {
  title: string
  topic: string
  description: string
}

const STORAGE_KEY = 'report-draft'

const saveDraft = debounce((data: DraftValues) => {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(data))
  console.log('草稿已保存', data)
}, 1000)

export function AutoSaveDraftForm() {
  const { register, watch, reset } = useForm<DraftValues>({
    defaultValues: { title: '', topic: '', description: '' },
  })

  useEffect(() => {
    const saved = localStorage.getItem(STORAGE_KEY)
    if (saved) {
      try {
        reset(JSON.parse(saved))
      } catch {
        // 草稿损坏，忽略
      }
    }
  }, [reset])

  useEffect(() => {
    const subscription = watch((value) => {
      saveDraft(value as DraftValues)
    })
    return () => subscription.unsubscribe()
  }, [watch])

  return (
    <form className="space-y-4">
      <input {...register('title')} placeholder="标题" className="w-full rounded border p-2" />
      <input {...register('topic')} placeholder="主题" className="w-full rounded border p-2" />
      <textarea
        {...register('description')}
        placeholder="描述"
        className="w-full rounded border p-2"
        rows={4}
      />
    </form>
  )
}
