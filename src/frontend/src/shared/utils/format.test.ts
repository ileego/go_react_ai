import { describe, it, expect } from 'vitest'
import { formatDate, truncate } from './format'

describe('format utilities', () => {
  it('truncate returns original string when within limit', () => {
    expect(truncate('hello', 10)).toBe('hello')
  })

  it('truncate appends ellipsis when exceeded', () => {
    expect(truncate('hello world', 5)).toBe('hello...')
  })

  it('formatDate returns a Chinese locale string', () => {
    const result = formatDate('2026-07-10T00:00:00.000Z')
    expect(result).toContain('2026')
    expect(result).toContain('7')
  })
})
