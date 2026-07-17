import { LayoutDemo } from '@/shared/components/design-system/LayoutDemo'
import { TokenCard } from '@/shared/components/design-system/TokenCard'
import { ShadcnShowcase } from '@/shared/components/design-system/ShadcnShowcase'

export function DesignSystemPage() {
  return (
    <div className="space-y-8">
      <section>
        <h1 className="text-3xl font-bold">设计系统展示</h1>
        <p className="mt-2 text-muted-foreground">
          本页用于直观展示 Tailwind CSS、设计令牌、响应式布局、暗色模式与 Shadcn UI 组件的效果。
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

      <section className="space-y-4">
        <h2 className="text-xl font-semibold">Shadcn UI 组件</h2>
        <ShadcnShowcase />
      </section>
    </div>
  )
}
