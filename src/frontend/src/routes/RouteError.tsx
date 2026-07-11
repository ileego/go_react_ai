import { useRouteError, isRouteErrorResponse } from 'react-router-dom'

export function RouteError() {
  const error = useRouteError()

  if (isRouteErrorResponse(error)) {
    return (
      <div style={{ padding: '2rem', color: '#dc2626' }}>
        <h1>{error.status}</h1>
        <p>{error.statusText}</p>
      </div>
    )
  }

  return (
    <div style={{ padding: '2rem', color: '#dc2626' }}>
      <h1>出错了</h1>
      <p>{error instanceof Error ? error.message : '未知错误'}</p>
    </div>
  )
}
