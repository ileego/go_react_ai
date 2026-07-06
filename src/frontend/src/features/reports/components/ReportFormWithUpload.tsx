import { useForm, Controller } from 'react-hook-form'
import {
  AttachmentUploader,
  type Attachment,
} from './AttachmentUploader'

interface ReportWithAttachmentsForm {
  title: string
  attachments: Attachment[]
}

export function ReportFormWithUpload() {
  const { control, handleSubmit } = useForm<ReportWithAttachmentsForm>({
    defaultValues: { title: '', attachments: [] },
  })

  const onSubmit = (data: ReportWithAttachmentsForm) => {
    console.log('带附件的报告:', data)
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      <Controller
        name="title"
        control={control}
        render={({ field }) => (
          <input
            {...field}
            placeholder="报告标题"
            className="w-full rounded border p-2"
          />
        )}
      />

      <Controller
        name="attachments"
        control={control}
        render={({ field }) => (
          <AttachmentUploader
            value={field.value}
            onChange={field.onChange}
          />
        )}
      />

      <button
        type="submit"
        className="rounded bg-blue-600 px-4 py-2 text-white"
      >
        提交
      </button>
    </form>
  )
}
