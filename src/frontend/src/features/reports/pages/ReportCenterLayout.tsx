import { Outlet, Link } from 'react-router-dom'

export function ReportCenterLayout() {
  return (
    <div className="grid grid-cols-1 gap-6 md:grid-cols-[200px_1fr]">
      <aside className="border-b border-gray-200 pb-4 md:border-b-0 md:border-r md:pr-4 dark:border-slate-700">
        <nav className="flex flex-row gap-3 md:flex-col">
          <Link to="/reports" className="hover:text-brand-600 dark:hover:text-brand-400">
            全部报告
          </Link>
          <Link to="/reports/new" className="hover:text-brand-600 dark:hover:text-brand-400">
            新建报告
          </Link>
        </nav>
      </aside>

      <section>
        <Outlet />
      </section>
    </div>
  )
}
