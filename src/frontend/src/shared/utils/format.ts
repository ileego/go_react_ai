// Shared Utilities
// 全局通用的纯函数工具，不依赖任何框架或业务逻辑。

export function formatDate(iso: string): string {
  return new Date(iso).toLocaleString('zh-CN')
}

export function truncate(str: string, maxLen: number): string {
  if (str.length <= maxLen) return str
  return str.slice(0, maxLen) + '...'
}
