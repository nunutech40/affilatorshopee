import { defineStore } from 'pinia'
import { apiRequest } from './productStore'

export const useCaptionStore = defineStore('captions', {
  state: () => ({
    current: null,
    variations: [],
    loading: false,
    error: '',
  }),
  actions: {
    async generate(productId, template, hashtags) {
      this.loading = true
      this.error = ''
      try {
        this.current = await apiRequest('/api/captions/generate', {
          method: 'POST',
          body: JSON.stringify({ product_id: productId, template, hashtags }),
        })
        return this.current
      } catch (error) {
        this.error = error.message
        throw error
      } finally {
        this.loading = false
      }
    },
    async generateVariations(productId, template, hashtags) {
      this.loading = true
      this.error = ''
      try {
        this.variations = await apiRequest('/api/captions/variations', {
          method: 'POST',
          body: JSON.stringify({ product_id: productId, template, count: 3, hashtags }),
        })
        return this.variations
      } catch (error) {
        this.error = error.message
        throw error
      } finally {
        this.loading = false
      }
    },
    async fetchVariations(productId) {
      this.variations = await apiRequest(`/api/products/${productId}/caption-variations`)
      return this.variations
    },
  },
})
