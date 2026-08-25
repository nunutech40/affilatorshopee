<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { apiRequest, useProductStore } from '@/stores/productStore'
import { useCaptionStore } from '@/stores/captionStore'
import ProductForm from '@/components/ProductForm.vue'
import CaptionGenerator from '@/components/CaptionGenerator.vue'
import PostLogForm from '@/components/PostLogForm.vue'

const route = useRoute()
const products = useProductStore()
const captions = useCaptionStore()
const product = ref(null)
const logs = ref([])
const loading = ref(true)
const saving = ref(false)
const working = ref(false)
const error = ref('')
const hasCaption = computed(() => captions.current?.caption || '')

async function load() { loading.value = true; error.value = ''; try { product.value = await products.getProduct(route.params.id); const logData = await apiRequest(`/api/post-logs?product_id=${route.params.id}`); logs.value = logData.items; await captions.fetchVariations(route.params.id) } catch (e) { error.value = e.message } finally { loading.value = false } }
async function save(payload) { saving.value = true; try { product.value = await products.updateProduct(route.params.id, payload) } catch (e) { error.value = e.message } finally { saving.value = false } }
async function reformat() { working.value = true; error.value = ''; try { const result = await products.reformat([route.params.id]); if (result.processed?.length) product.value = result.processed[0]; if (result.failed?.length) error.value = result.failed[0].message } catch (e) { error.value = e.message } finally { working.value = false } }
async function markReady() { await save({ status: 'ready' }) }
async function refreshLogs() { const logData = await apiRequest(`/api/post-logs?product_id=${route.params.id}`); logs.value = logData.items; product.value = await products.getProduct(route.params.id) }
onMounted(load)
</script>

<template>
  <div v-if="loading" class="loading">Memuat product...</div>
  <div v-else-if="error && !product" class="error-box">{{ error }}</div>
  <template v-else-if="product">
    <RouterLink to="/" class="back-link">← Kembali ke library</RouterLink>
    <header class="detail-header"><div class="detail-title"><span class="status" :class="product.status">{{ product.status }}</span><h1>{{ product.product_name || 'Raw product' }}</h1><p class="cluster">{{ product.cluster || 'uncategorized' }} · {{ product.post_count || 0 }} posting</p></div><div class="detail-actions"><button class="button" :disabled="working || product.status !== 'raw'" @click="reformat">{{ working ? 'AI...' : 'Reformat AI' }}</button><button v-if="product.status !== 'ready'" class="button-primary" @click="markReady">Tandai ready</button></div></header>
    <div v-if="error" class="error-box">{{ error }}</div>
    <div class="form-layout"><div><ProductForm :product="product" :saving="saving" @save="save" /><div class="panel raw-panel"><h3>Raw text asli</h3><pre>{{ product.raw_text }}</pre></div></div><div><CaptionGenerator :product="product" /><PostLogForm v-if="hasCaption" :product-id="product.id" :caption="hasCaption" :hashtags="captions.current.hashtags" @saved="refreshLogs" /><section class="panel"><h2>Riwayat posting</h2><div v-if="logs.length" class="log-list"><article v-for="log in logs" :key="log.id" class="log-item"><div class="log-head"><span>{{ log.platform }}</span><span>{{ new Date(log.posted_at).toLocaleString('id-ID') }}</span></div><div class="log-text">{{ log.caption }}</div></article></div><p v-else class="muted">Belum ada catatan posting.</p></section></div></div>
  </template>
</template>

<style scoped>.raw-panel { margin-top: 14px; }.raw-panel pre { white-space: pre-wrap; max-height: 350px; overflow: auto; color: #52655a; font: 13px/1.6 'DM Mono'; }.muted { color: #78867c; font-size: 13px; }</style>
