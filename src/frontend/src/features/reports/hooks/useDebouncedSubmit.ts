import { useCallback, useRef } from 'react'

export function useDebouncedSubmit<T>(onSubmit: (data: T) => Promise<unknown>, delay = 500) {
  const isSubmittingRef = useRef(false)

  return useCallback(
    async (data: T) => {
      if (isSubmittingRef.current) return
      isSubmittingRef.current = true
      try {
        return await onSubmit(data)
      } finally {
        setTimeout(() => {
          isSubmittingRef.current = false
        }, delay)
      }
    },
    [onSubmit, delay]
  )
}
