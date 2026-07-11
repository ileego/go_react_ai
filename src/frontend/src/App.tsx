import { BrowserRouter, Routes, Route, Link } from 'react-router-dom'
import { ReportList } from '@/features/reports'

// 注意：本文件是 17.1.2 节的声明式路由教学示例。
// 项目主入口（main.tsx）目前使用数据 API 路由（routes.tsx + RouterProvider）。

function Home() {
  return (
    <div>
      <h1>深度研究与报告平台</h1>
      <p>欢迎使用 Go + React + AI 全栈应用。</p>
    </div>
  )
}

export default function App() {
  return (
    <BrowserRouter>
      <nav style={{ display: 'flex', gap: '1rem', padding: '1rem 0' }}>
        <Link to="/">首页</Link>
        <Link to="/reports">报告列表</Link>
        <Link to="/reports/new">新建报告</Link>
      </nav>

      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/reports" element={<ReportList />} />
      </Routes>
    </BrowserRouter>
  )
}
