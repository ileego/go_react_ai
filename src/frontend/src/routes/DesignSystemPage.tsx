import { LayoutDemo } from '@/shared/components/design-system/LayoutDemo'
import { TokenCard } from '@/shared/components/design-system/TokenCard'

export function DesignSystemPage() {
  return (
    <div className="space-y-8">
      <section>
        <h1 className="text-3xl font-bold">设计系统展示</h1>
        <p className="mt-2 text-gray-600 dark:text-gray-300">
          本页用于直观展示 Tailwind CSS、设计令牌、响应式布局与暗色模式的效果。
        </p>
      </section>

      <section className="space-y-4">
        <h2 className="text-xl font-semibold">设计令牌</h2>
        <TokenCard />
      </section>

      <section className="space-y-4">
        <h2 className="text-xl font-semibold">布局系统</h2>
        <LayoutDemo />
      </section>
    </div>
  )
}
