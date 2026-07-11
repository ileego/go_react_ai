import { useActionState } from 'react'
import { Button } from '@/shared/components/Button'

interface ActionState {
  ok: boolean
  message: string
}

async function createReportAction(
  _prevState: ActionState | null,
  formData: FormData
): Promise<ActionState> {
  const topic = formData.get('topic') as string
  const depth = formData.get('depth') as string
  console.log('创建报告:', { topic, depth })
  return { ok: true, message: `已创建 ${topic}` }
}

export function ActionForm() {
  const [state, submitAction, isPending] = useActionState(createReportAction, null)

  return (
    <form action={submitAction} className="space-y-4">
      <input
        name="topic"
        type="text"
        placeholder="研究主题"
        required
        className="w-full rounded border p-2"
      />
      <select name="depth" defaultValue="medium" className="w-full rounded border p-2">
        <option value="shallow">浅度</option>
        <option value="medium">中度</option>
        <option value="deep">深度</option>
      </select>
      <Button type="submit" disabled={isPending}>
        {isPending ? '提交中...' : '创建报告'}
      </Button>
      {state?.message && <p className="text-sm text-green-600">{state.message}</p>}
    </form>
  )
}
