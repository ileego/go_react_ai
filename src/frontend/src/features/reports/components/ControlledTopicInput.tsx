import { useState } from 'react'

export function ControlledTopicInput() {
  const [topic, setTopic] = useState('')

  return (
    <input
      type="text"
      value={topic}
      onChange={(e) => setTopic(e.target.value)}
      placeholder="输入研究主题"
      className="w-full rounded border p-2"
    />
  )
}
