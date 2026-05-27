import { useEffect, useState } from 'react'

function App() {
  const [health, setHealth] = useState<string>('checking...')

  useEffect(() => {
    fetch('/api/health')
      .then((res) => res.json())
      .then((data) => setHealth(JSON.stringify(data, null, 2)))
      .catch((err) => setHealth(`error: ${err.message}`))
  }, [])

  return (
    <div style={{ padding: '2rem', fontFamily: 'system-ui, sans-serif' }}>
      <h1>Go + React + AI</h1>
      <p>前端运行正常。后端健康检查:</p>
      <pre
        style={{
          background: '#f5f5f5',
          padding: '1rem',
          borderRadius: '8px',
        }}
      >
        {health}
      </pre>
    </div>
  )
}

export default App
