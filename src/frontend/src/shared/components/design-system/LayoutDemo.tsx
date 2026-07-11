export function LayoutDemo() {
  return (
    <div className="space-y-6">
      <section>
        <h3 className="mb-2 font-medium">Flex 布局</h3>
        <div className="flex flex-col gap-4 rounded-2xl border border-gray-200 p-4 md:flex-row md:items-center md:justify-between dark:border-slate-700">
          <div className="rounded bg-brand-100 px-4 py-2 text-brand-700 dark:bg-brand-900 dark:text-brand-200">
            左侧
          </div>
          <div className="rounded bg-brand-100 px-4 py-2 text-brand-700 dark:bg-brand-900 dark:text-brand-200">
            中间
          </div>
          <div className="rounded bg-brand-100 px-4 py-2 text-brand-700 dark:bg-brand-900 dark:text-brand-200">
            右侧
          </div>
        </div>
      </section>

      <section>
        <h3 className="mb-2 font-medium">响应式 Grid 布局</h3>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <div
              key={i}
              className="rounded-2xl border border-gray-200 bg-white p-4 shadow-card dark:border-slate-700 dark:bg-slate-800"
            >
              卡片 {i + 1}
            </div>
          ))}
        </div>
      </section>
    </div>
  )
}
