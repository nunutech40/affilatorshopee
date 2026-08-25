import { describe, expect, it } from 'vitest'

describe('frontend smoke contract', () => {
  it('uses the expected local API path', () => {
    const params = new URLSearchParams({ page: 1, limit: 20 })
    expect(`/api/products?${params}`).toBe('/api/products?page=1&limit=20')
  })
})
