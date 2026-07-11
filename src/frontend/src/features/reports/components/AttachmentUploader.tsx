import { useState, useCallback, useEffect, type ChangeEvent } from 'react'
import { Button } from '@/shared/components/Button'

export interface Attachment {
  id: string
  name: string
  url: string
}

interface UploadFile {
  id: string
  file: File
  status: 'pending' | 'uploading' | 'done' | 'error'
  progress: number
  previewUrl?: string
  attachment?: Attachment
}

interface AttachmentUploaderProps {
  value?: Attachment[]
  onChange?: (attachments: Attachment[]) => void
}

function generateId() {
  return Math.random().toString(36).slice(2)
}

export function AttachmentUploader({ value = [], onChange }: AttachmentUploaderProps) {
  const [queue, setQueue] = useState<UploadFile[]>([])
  const [isDragging, setIsDragging] = useState(false)

  // 受控模式下，把外部 value 同步到内部队列（仅初始化/增删时）
  useEffect(() => {
    setQueue((prev) => {
      const existingIds = new Set(prev.map((f) => f.id))
      const merged = [...prev]
      for (const attachment of value) {
        if (existingIds.has(attachment.id)) continue
        merged.push({
          id: attachment.id,
          file: new File([], attachment.name),
          status: 'done',
          progress: 100,
          attachment,
        })
      }
      return merged
    })
  }, [value])

  const notifyChange = useCallback(
    (nextQueue: UploadFile[]) => {
      if (!onChange) return
      const attachments = nextQueue
        .filter((f) => f.status === 'done' && f.attachment)
        .map((f) => f.attachment!)
      onChange(attachments)
    },
    [onChange]
  )

  const addFiles = useCallback(
    (files: FileList | null) => {
      if (!files) return
      const newFiles: UploadFile[] = Array.from(files).map((file) => ({
        id: generateId(),
        file,
        status: 'pending',
        progress: 0,
        previewUrl: file.type.startsWith('image/') ? URL.createObjectURL(file) : undefined,
      }))
      setQueue((prev) => [...prev, ...newFiles])
    },
    [setQueue]
  )

  const startUpload = useCallback(async () => {
    for (const item of queue) {
      if (item.status !== 'pending') continue
      setQueue((prev) => prev.map((f) => (f.id === item.id ? { ...f, status: 'uploading' } : f)))

      for (let progress = 0; progress <= 100; progress += 20) {
        await new Promise((resolve) => setTimeout(resolve, 200))
        setQueue((prev) => prev.map((f) => (f.id === item.id ? { ...f, progress } : f)))
      }

      const attachment: Attachment = {
        id: item.id,
        name: item.file.name,
        url: `/uploads/${item.file.name}`,
      }

      setQueue((prev): UploadFile[] => {
        const next = prev.map(
          (f): UploadFile =>
            f.id === item.id ? { ...f, status: 'done', progress: 100, attachment } : f
        )
        notifyChange(next)
        return next
      })
    }
  }, [queue, notifyChange])

  const removeFile = useCallback(
    (id: string) => {
      setQueue((prev) => {
        const item = prev.find((f) => f.id === id)
        if (item?.previewUrl) URL.revokeObjectURL(item.previewUrl)
        const next = prev.filter((f) => f.id !== id)
        notifyChange(next)
        return next
      })
    },
    [notifyChange]
  )

  return (
    <div className="space-y-4">
      <div
        onDragOver={(e) => {
          e.preventDefault()
          setIsDragging(true)
        }}
        onDragLeave={(e) => {
          e.preventDefault()
          setIsDragging(false)
        }}
        onDrop={(e) => {
          e.preventDefault()
          setIsDragging(false)
          addFiles(e.dataTransfer.files)
        }}
        className={`border-2 border-dashed p-6 text-center ${
          isDragging ? 'border-blue-500 bg-blue-50' : 'border-gray-300'
        }`}
      >
        <input
          type="file"
          multiple
          onChange={(e: ChangeEvent<HTMLInputElement>) => addFiles(e.target.files)}
        />
        <p className="text-sm text-gray-500">支持拖拽上传多个附件</p>
      </div>

      <ul className="space-y-2">
        {queue.map((item) => (
          <li key={item.id} className="flex items-center gap-3 rounded border p-2">
            {item.previewUrl && (
              <img
                src={item.previewUrl}
                alt={item.file.name}
                className="h-12 w-12 rounded object-cover"
              />
            )}
            <div className="flex-1">
              <p className="text-sm font-medium">{item.file.name}</p>
              <div className="h-2 w-full rounded bg-gray-200">
                <div className="h-2 rounded bg-blue-500" style={{ width: `${item.progress}%` }} />
              </div>
              <p className="text-xs text-gray-500">
                {item.status === 'done'
                  ? '上传完成'
                  : item.status === 'uploading'
                    ? `上传中 ${item.progress}%`
                    : '等待上传'}
              </p>
            </div>
            <Button type="button" variant="outline" onClick={() => removeFile(item.id)}>
              删除
            </Button>
          </li>
        ))}
      </ul>

      {queue.some((f) => f.status === 'pending') && (
        <Button type="button" onClick={startUpload}>
          开始上传
        </Button>
      )}
    </div>
  )
}
