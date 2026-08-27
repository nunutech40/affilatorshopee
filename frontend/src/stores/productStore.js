import { defineStore } from 'pinia'

const API = import.meta.env.VITE_API_URL || ''

async function request(path, options = {}) {
  const isFormData = options.body instanceof FormData
  const response = await fetch(`${API}${path}`, {
    headers: { ...(isFormData ? {} : { 'Content-Type': 'application/json' }), ...(options.headers || {}) },
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
    filters: { search: '', status: '', content_model: '', source_category: '', cluster: '' },
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
    async reformat(ids, model, variant = false) {
      this.loading = true
      this.error = ''
      try {
        return await request('/api/ai/reformat', { method: 'POST', body: JSON.stringify({ product_ids: ids, model, variant }) })
      } finally {
        this.loading = false
      }
    },
    async fetchModels() {
      return request('/api/ai/models')
    },
    async deleteProduct(id) {
      return request(`/api/products/${id}`, { method: 'DELETE' })
    },
    async importClicks(file) {
      const body = new FormData()
      body.append('file', file)
      return request('/api/analytics/clicks/import', { method: 'POST', body })
    },
    async importCommissions(file) {
      const body = new FormData()
      body.append('file', file)
      return request('/api/analytics/commissions/import', { method: 'POST', body })
    },
    async fetchSoldProducts(page = 1, limit = 20, search = '', filters = {}) {
      const params = new URLSearchParams({ page, limit })
      if (search) params.set('search', search)
      if (filters.month) params.set('month', filters.month)
      if (filters.start_date) params.set('start_date', filters.start_date)
      if (filters.end_date) params.set('end_date', filters.end_date)
      return request(`/api/analytics/commissions/sold?${params}`)
    },
    async fetchCommissionEvents(page = 1, limit = 50, search = '', filters = {}) {
      const params = new URLSearchParams({ page, limit })
      if (search) params.set('search', search)
      if (filters.month) params.set('month', filters.month)
      if (filters.start_date) params.set('start_date', filters.start_date)
      if (filters.end_date) params.set('end_date', filters.end_date)
      return request(`/api/analytics/commissions/events?${params}`)
    },
    async fetchCommissionSummary(search = '', filters = {}) {
      const params = new URLSearchParams()
      if (search) params.set('search', search)
      if (filters.month) params.set('month', filters.month)
      if (filters.start_date) params.set('start_date', filters.start_date)
      if (filters.end_date) params.set('end_date', filters.end_date)
      return request(`/api/analytics/commissions/summary?${params}`)
    },
  },
})
