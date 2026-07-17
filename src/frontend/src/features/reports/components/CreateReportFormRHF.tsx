import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { reportSchema, type ReportFormData } from '../schemas/reportSchema'

interface CreateReportFormProps {
  onSubmit: (data: ReportFormData) => void
  isLoading?: boolean
}

export function CreateReportForm({ onSubmit, isLoading }: CreateReportFormProps) {
  const form = useForm<ReportFormData>({
    resolver: zodResolver(reportSchema),
    defaultValues: {
      title: '',
      topic: '',
      depth: 'medium',
      model: 'kimi-latest',
      description: '',
    },
  })

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
        <FormField
          control={form.control}
          name="title"
          render={({ field }) => (
            <FormItem>
              <FormLabel>报告标题</FormLabel>
              <FormControl>
                <Input placeholder="例如：多智能体协作调研" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="topic"
          render={({ field }) => (
            <FormItem>
              <FormLabel>研究主题</FormLabel>
              <FormControl>
                <Input placeholder="输入核心研究问题" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="depth"
          render={({ field }) => (
            <FormItem>
              <FormLabel>研究深度</FormLabel>
              <Select onValueChange={field.onChange} defaultValue={field.value}>
                <FormControl>
                  <SelectTrigger>
                    <SelectValue placeholder="选择研究深度" />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  <SelectItem value="shallow">浅度概览</SelectItem>
                  <SelectItem value="medium">中度分析</SelectItem>
                  <SelectItem value="deep">深度研究</SelectItem>
                </SelectContent>
              </Select>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="model"
          render={({ field }) => (
            <FormItem>
              <FormLabel>默认模型</FormLabel>
              <Select onValueChange={field.onChange} defaultValue={field.value}>
                <FormControl>
                  <SelectTrigger>
                    <SelectValue placeholder="选择默认模型" />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  <SelectItem value="gpt-4o">GPT-4o</SelectItem>
                  <SelectItem value="claude-3-5">Claude 3.5</SelectItem>
                  <SelectItem value="kimi-latest">Kimi Latest</SelectItem>
                </SelectContent>
              </Select>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="description"
          render={({ field }) => (
            <FormItem>
              <FormLabel>补充描述</FormLabel>
              <FormControl>
                <Textarea
                  placeholder="描述报告目标、预期读者、参考资料等"
                  {...field}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <Button type="submit" disabled={isLoading}>
          {isLoading ? '创建中...' : '创建报告'}
        </Button>
      </form>
    </Form>
  )
}

export type { ReportFormData as CreateReportFormValues }
