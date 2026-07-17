import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Skeleton } from '@/components/ui/skeleton'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Separator } from '@/components/ui/separator'
import { Terminal } from 'lucide-react'

export function ShadcnShowcase() {
  return (
    <div className="space-y-8">
      <section className="space-y-4">
        <h3 className="text-lg font-semibold">基础组件</h3>
        <div className="flex flex-wrap gap-3">
          <Button>默认按钮</Button>
          <Button variant="secondary">次要</Button>
          <Button variant="outline">边框</Button>
          <Button variant="destructive">危险</Button>
          <Button variant="ghost">幽灵</Button>
          <Button variant="link">链接</Button>
        </div>
      </section>

      <Separator />

      <section className="space-y-4">
        <h3 className="text-lg font-semibold">表单组件</h3>
        <div className="grid max-w-md gap-4">
          <div className="space-y-2">
            <Label htmlFor="demo-email">邮箱</Label>
            <Input id="demo-email" placeholder="name@example.com" />
          </div>
          <div className="space-y-2">
            <Label htmlFor="demo-bio">简介</Label>
            <Textarea id="demo-bio" placeholder="写点什么..." />
          </div>
          <div className="space-y-2">
            <Label>模型</Label>
            <Select>
              <SelectTrigger>
                <SelectValue placeholder="选择模型" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="gpt-4o">GPT-4o</SelectItem>
                <SelectItem value="claude">Claude 3.5</SelectItem>
                <SelectItem value="kimi">Kimi Latest</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
      </section>

      <Separator />

      <section className="space-y-4">
        <h3 className="text-lg font-semibold">卡片与徽章</h3>
        <Card className="max-w-md">
          <CardHeader>
            <CardTitle>Shadcn Card</CardTitle>
            <CardDescription>这是一个基于 Tailwind CSS 变量的容器组件。</CardDescription>
          </CardHeader>
          <CardContent className="space-y-2">
            <div className="flex gap-2">
              <Badge>默认</Badge>
              <Badge variant="secondary">次要</Badge>
              <Badge variant="outline">边框</Badge>
              <Badge variant="destructive">危险</Badge>
            </div>
          </CardContent>
        </Card>
      </section>

      <Separator />

      <section className="space-y-4">
        <h3 className="text-lg font-semibold">反馈与占位</h3>
        <Alert className="max-w-md">
          <Terminal className="h-4 w-4" />
          <AlertTitle>提示</AlertTitle>
          <AlertDescription>Shadcn Alert 可用于展示重要提示信息。</AlertDescription>
        </Alert>

        <div className="flex items-center gap-4">
          <Skeleton className="h-12 w-12 rounded-full" />
          <div className="space-y-2">
            <Skeleton className="h-4 w-[250px]" />
            <Skeleton className="h-4 w-[200px]" />
          </div>
        </div>
      </section>

      <Separator />

      <section className="space-y-4">
        <h3 className="text-lg font-semibold">头像</h3>
        <div className="flex gap-2">
          <Avatar>
            <AvatarImage src="https://github.com/shadcn.png" alt="@shadcn" />
            <AvatarFallback>CN</AvatarFallback>
          </Avatar>
          <Avatar>
            <AvatarFallback>AI</AvatarFallback>
          </Avatar>
        </div>
      </section>
    </div>
  )
}
