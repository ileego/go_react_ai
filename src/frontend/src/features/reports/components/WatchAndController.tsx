import { useForm, Controller } from 'react-hook-form'

interface Values {
  depth: 'shallow' | 'medium' | 'deep'
  model: 'gpt-4o' | 'claude-3-5' | 'kimi-latest'
}

export function WatchAndController() {
  const { register, watch, control } = useForm<Values>({
    defaultValues: { depth: 'medium', model: 'kimi-latest' },
  })

  const depth = watch('depth')

  return (
    <div className="space-y-2">
      <select {...register('depth')} className="w-full rounded border p-2">
        <option value="shallow">浅度</option>
        <option value="medium">中度</option>
        <option value="deep">深度</option>
      </select>
      {depth === 'deep' && <input placeholder="详细章节数" className="w-full rounded border p-2" />}

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
  )
}
