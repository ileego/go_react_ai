import { useSearchParams } from 'react-router-dom'

export function ReportSearch() {
  const [searchParams, setSearchParams] = useSearchParams()
  const keyword = searchParams.get('keyword') || ''

  const handleChange = (value: string) => {
    setSearchParams((prev) => {
      if (value) {
        prev.set('keyword', value)
      } else {
        prev.delete('keyword')
      }
      return prev
    })
  }

  return (
    <input
      value={keyword}
      onChange={(e) => handleChange(e.target.value)}
      placeholder="搜索报告"
      className="w-full rounded-lg border border-gray-300 p-2 focus:border-brand-500 focus:outline-none focus:ring-2 focus:ring-brand-200 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-100"
    />
  )
}
