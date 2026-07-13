export function TokenCard() {
  return (
    <div className="rounded-2xl border border-gray-200 bg-white p-6 shadow-card dark:border-slate-700 dark:bg-slate-800">
      <h3 className="text-lg font-semibold text-gray-900 dark:text-slate-100">TokenCard</h3>
      <p className="mt-2 text-gray-600 dark:text-gray-300">
        这张卡片使用了品牌色、自定义圆角、阴影与间距令牌。
      </p>
      <div className="mt-4 flex flex-wrap gap-3">
        <span className="rounded bg-brand-100 px-3 py-1 text-sm text-brand-700 dark:bg-brand-900 dark:text-brand-200">
          brand-100 / brand-700
        </span>
        <span className="rounded bg-brand-600 px-3 py-1 text-sm text-white">brand-600</span>
        <span className="rounded border border-gray-300 px-3 py-1 text-sm text-gray-700 dark:border-slate-600 dark:text-slate-200">
          outline
        </span>
      </div>
    </div>
  )
}
