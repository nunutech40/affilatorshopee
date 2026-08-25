import { defineStore } from 'pinia'

const API = import.meta.env.VITE_API_URL || ''

async function request(path, options = {}) {
  const response = await fetch(`${API}${path}`, {
    headers: { 'Content-Type': 'application/json', ...(options.headers || {}) },
    ...options,
  })
  if (response.status === 204) return null
  const body = await response.json().catch(() => null)
  if (!response.ok || !body?.success) {
    throw new Error(body?.error?.message || 'Request gagal')
  }
  return body.data
}

export const apiRequest = request

export const useProductStore = defineStore('products', {
  state: () => ({
    items: [],
    total: 0,
    page: 1,
    limit: 20,
    loading: false,
    error: '',
    filters: { search: '', status: '', content_model: '', cluster: '' },
  }),
  actions: {
    async fetchProducts() {
      this.loading = true
      this.error = ''
      try {
        const params = new URLSearchParams({ page: this.page, limit: this.limit })
        Object.entries(this.filters).forEach(([key, value]) => {
          if (value) params.set(key, value)
        })
        const data = await request(`/api/products?${params}`)
        this.items = data.items
        this.total = data.total
      } catch (error) {
        this.error = error.message
      } finally {
        this.loading = false
      }
    },
    async createProduct(payload) {
      return request('/api/products', { method: 'POST', body: JSON.stringify(payload) })
    },
    async getProduct(id) {
      return request(`/api/products/${id}`)
    },
    async updateProduct(id, payload) {
      return request(`/api/products/${id}`, { method: 'PATCH', body: JSON.stringify(payload) })
    },
    async reformat(ids, model) {
      return request('/api/ai/reformat', { method: 'POST', body: JSON.stringify({ product_ids: ids, model }) })
    },
    async fetchModels() {
      return request('/api/ai/models')
    },
    async deleteProduct(id) {
      return request(`/api/products/${id}`, { method: 'DELETE' })
    },
  },
})
