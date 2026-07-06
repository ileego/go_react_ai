export interface FileValidationOptions {
  maxSizeMB?: number
  allowedTypes?: string[]
  maxCount?: number
}

export function validateFiles(
  files: File[],
  options: FileValidationOptions,
): string | null {
  if (options.maxCount && files.length > options.maxCount) {
    return `最多上传 ${options.maxCount} 个文件`
  }

  for (const file of files) {
    if (options.maxSizeMB && file.size > options.maxSizeMB * 1024 * 1024) {
      return `${file.name} 超过 ${options.maxSizeMB}MB 限制`
    }
    if (options.allowedTypes && !options.allowedTypes.includes(file.type)) {
      return `${file.name} 的文件类型不允许`
    }
  }

  return null
}
