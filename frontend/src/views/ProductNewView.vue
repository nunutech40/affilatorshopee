<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import ProductParser from '@/components/ProductParser.vue'
import { useProductStore } from '@/stores/productStore'

const router = useRouter()
const products = useProductStore()
const error = ref('')
const saving = ref(false)
async function save(payload) {
  error.value = ''
  saving.value = true
  let created = null
  try {
    created = await products.createProduct(payload)
    const model = localStorage.getItem('ai_model') || ''
    try {
      await products.reformat([created.product.id], model, false)
    } catch (e) {
      // Saving the raw product is still successful. Leave it raw so detail can retry.
      router.push({ path: `/products/${created.product.id}`, query: { ai: 'failed' } })
      return
    }
    router.push(`/products/${created.product.id}`)
  } catch (e) {
    error.value = created
      ? `Produk tersimpan sebagai raw text. Reformat AI gagal: ${e.message}`
      : e.message
  } finally { saving.value = false }
}
async function onImported() { router.push('/') }
</script>

<template><RouterLink to="/" class="back-link">← Kembali ke library</RouterLink><div v-if="error" class="error-box">{{ error }}</div><div v-if="saving" class="ai-loading"><span class="spinner"></span>AI sedang membuat caption promo dari raw text. Jangan tutup halaman.</div><ProductParser :saving="saving" @save="save" @imported="onImported" /></template>

<style scoped>.ai-loading{display:flex;align-items:center;gap:10px;margin:14px 0;padding:13px 16px;border-radius:8px;background:#e7f3eb;color:#176b4f;font-weight:600}.spinner{width:16px;height:16px;border:2px solid #b8d9c6;border-top-color:#176b4f;border-radius:50%;animation:spin .8s linear infinite}@keyframes spin{to{transform:rotate(360deg)}}</style>
