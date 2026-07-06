import { useRef } from 'react'
import { Button } from '@/shared/components/Button'

export function UncontrolledTopicInput() {
  const inputRef = useRef<HTMLInputElement>(null)

  const handleSubmit = () => {
    const value = inputRef.current?.value ?? ''
    console.log('主题:', value)
  }

  return (
    <div className="flex gap-2">
      <input
        ref={inputRef}
        type="text"
        placeholder="输入研究主题"
        className="w-full rounded border p-2"
      />
      <Button type="button" onClick={handleSubmit}>
        提交
      </Button>
    </div>
  )
}
