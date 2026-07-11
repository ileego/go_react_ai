import { useForm, useFieldArray } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Button } from '@/shared/components/Button'

const schema = z.object({
  title: z.string().min(2),
  researchers: z
    .array(
      z.object({
        name: z.string().min(1, '请输入姓名'),
        role: z.enum(['lead', 'analyst', 'writer'], {
          errorMap: () => ({ message: '请选择角色' }),
        }),
      })
    )
    .min(1, '至少需要一名研究员'),
})

type Values = z.infer<typeof schema>

export function ResearcherFieldArray() {
  const { register, control, formState, handleSubmit } = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: {
      title: '',
      researchers: [{ name: '', role: 'lead' }],
    },
  })

  const { fields, append, remove } = useFieldArray({
    control,
    name: 'researchers',
  })

  const onSubmit = (data: Values) => console.log(data)

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      <input {...register('title')} placeholder="报告标题" className="w-full rounded border p-2" />

      <div className="space-y-2">
        <h4>研究员列表</h4>
        {fields.map((field, index) => (
          <div key={field.id} className="flex gap-2">
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
      </div>

      {formState.errors.researchers?.root && (
        <p className="text-red-500">{formState.errors.researchers.root.message}</p>
      )}

      <Button type="submit">提交</Button>
    </form>
  )
}
