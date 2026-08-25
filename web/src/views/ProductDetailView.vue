<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { apiRequest, useProductStore } from '@/stores/productStore'
import { useCaptionStore } from '@/stores/captionStore'
import CaptionGenerator from '@/components/CaptionGenerator.vue'
import PostLogForm from '@/components/PostLogForm.vue'
import ModelSelector from '@/components/ModelSelector.vue'

const route = useRoute()
const router = useRouter()
const products = useProductStore()
const captions = useCaptionStore()
const product = ref(null)
const logs = ref([])
const media = ref([])
const loading = ref(true)
const saving = ref(false)
const working = ref(false)
const selectedModel = ref(localStorage.getItem('ai_model') || 'opencode/muse-spark-1.2-contributor-free')
const error = ref('')
const jsonError = ref('')
const editText = ref('')
const hasCaption = computed(() => captions.current?.caption || '')

function formatReformat(p) {
  if (!p) return ''
  const obj = {
    product_name: p.product_name || '',
    shopee_link: p.shopee_link || '',
    image_url: p.image_url || '',
    image_urls: p.image_urls || [],
    video_url: p.video_url || '',
    keyword: p.keyword || '',
    problem: p.problem || '',
    cluster: p.cluster || '',
    content_model: p.content_model || '',
    capture_angle: p.capture_angle || '',
    benefit_1: p.benefit_1 || '',
    benefit_2: p.benefit_2 || '',
    benefit_3: p.benefit_3 || '',
    urgency: p.urgency || '',
    caption_template: p.caption_template || 'direct_product',
    hashtag_pool: p.hashtag_pool || [],
    normal_price: p.normal_price,
    sale_price: p.sale_price,
    discount_percent: p.discount_percent,
    rating: p.rating,
    sold_count: p.sold_count || '',
    review_count: p.review_count || '',
    notes: p.notes || ''
  }
  return JSON.stringify(obj, null, 2)
}

watch(product, (p) => { if (p) editText.value = formatReformat(p) })

async function load() { loading.value = true; error.value = ''; try { product.value = await products.getProduct(route.params.id); const logData = await apiRequest(`/api/post-logs?product_id=${route.params.id}`); logs.value = logData.items; media.value = await apiRequest(`/api/products/${route.params.id}/media`); await captions.fetchVariations(route.params.id) } catch (e) { error.value = e.message } finally { loading.value = false } }

async function saveReformat() {
  jsonError.value = ''
  let parsed
  try { parsed = JSON.parse(editText.value) } catch (e) { jsonError.value = 'JSON tidak valid: ' + e.message; return }
  // kosong -> null agar bisa dihapus
  Object.keys(parsed).forEach(k => { if (parsed[k] === '') parsed[k] = null })
  saving.value = true
  try { product.value = await products.updateProduct(route.params.id, parsed); editText.value = formatReformat(product.value) } catch (e) { jsonError.value = e.message } finally { saving.value = false }
}

async function reformat() { working.value = true; error.value = ''; try { const result = await products.reformat([route.params.id], selectedModel.value); if (result.processed?.length) { product.value = result.processed[0]; editText.value = formatReformat(product.value) } if (result.failed?.length) error.value = result.failed[0].code + ': ' + result.failed[0].message } catch (e) { error.value = e.message } finally { working.value = false } }
async function markReady() { try { const updated = await products.updateProduct(route.params.id, { status: 'ready' }); product.value = updated } catch (e) { error.value = e.message } }
async function remove() {
  if (!confirm('Hapus produk ini permanen?')) return
  try { await products.deleteProduct(route.params.id); router.push('/') } catch (e) { error.value = e.message }
}
async function refreshLogs() { const logData = await apiRequest(`/api/post-logs?product_id=${route.params.id}`); logs.value = logData.items; product.value = await products.getProduct(route.params.id); editText.value = formatReformat(product.value) }
onMounted(load)
</script>

<template>
  <div v-if="loading" class="loading">Memuat product...</div>
  <div v-else-if="error && !product" class="error-box">{{ error }}</div>
  <template v-else-if="product">
    <RouterLink to="/" class="back-link">← Kembali ke library</RouterLink>
    <header class="detail-header"><div class="detail-title"><span class="status" :class="product.status">{{ product.status }}</span><h1>{{ product.product_name || 'Raw product' }}</h1><p class="cluster">{{ product.cluster || 'uncategorized' }} · {{ product.post_count || 0 }} posting</p></div><div class="detail-actions" style="flex-direction:column; align-items:end"><ModelSelector v-model="selectedModel" /><div style="display:flex; gap:8px; flex-wrap:wrap; justify-content:end"><button class="button" :disabled="working || product.status !== 'raw'" @click="reformat">{{ working ? 'AI...' : 'Reformat AI' }}</button><button v-if="product.status !== 'ready'" class="button-primary" @click="markReady">Tandai ready</button><button class="button button-danger" @click="remove">Hapus</button></div></div></header>
    <div v-if="error" class="error-box">{{ error }}</div>
    <div class="form-layout">
      <div>
        <section class="panel"><h2>Raw text (asli, read-only)</h2><pre class="raw-pre">{{ product.raw_text }}</pre></section>
        <section class="panel reformat-panel">
          <div class="section-heading"><h2>Reformat (editable, 1 textarea)</h2><span class="counter">{{ product.status }}</span></div>
          <p class="muted">Edit JSON di bawah lalu Save. Semua field terstruktur ada di satu tempat — tidak perlu form terpisah.</p>
          <textarea v-model="editText" class="textarea mono" rows="22" spellcheck="false"></textarea>
          <div v-if="jsonError" class="error-box" style="margin-top:10px">{{ jsonError }}</div>
          <div style="display:flex; gap:8px; margin-top:12px"><button class="button-primary" :disabled="saving" @click="saveReformat">{{ saving ? 'Menyimpan...' : 'Save reformat' }}</button><button class="button" @click="editText = formatReformat(product)">Reset</button></div>
        </section>
        <section class="panel"><div class="section-heading"><h2>Media lokal</h2><a v-if="media.length" class="card-link" :href="`/api/products/${product.id}/media/download`">Download ZIP →</a></div><div v-if="media.length" class="media-list"><div v-for="file in media" :key="file.id" class="media-item"><span>{{ file.filename }}</span><small>{{ file.media_type }} · {{ Math.ceil(file.size_bytes / 1024) }} KB</small></div></div><p v-else class="muted">Belum ada media yang berhasil di-download.</p></section>
      </div>
      <div><CaptionGenerator :product="product" /><PostLogForm v-if="hasCaption" :product-id="product.id" :caption="hasCaption" :hashtags="captions.current.hashtags" @saved="refreshLogs" /><section class="panel"><h2>Riwayat posting</h2><div v-if="logs.length" class="log-list"><article v-for="log in logs" :key="log.id" class="log-item"><div class="log-head"><span>{{ log.platform }}</span><span>{{ new Date(log.posted_at).toLocaleString('id-ID') }}</span></div><div class="log-text">{{ log.caption }}</div></article></div><p v-else class="muted">Belum ada catatan posting.</p></section></div>
    </div>
  </template>
</template>

<style scoped>.raw-pre{white-space:pre-wrap; max-height:320px; overflow:auto; color:#52655a; font:13px/1.6 'DM Mono'; background:#f9fbf8; padding:14px; border-radius:8px; border:1px solid #e1e8e0} .reformat-panel .mono{font:12px/1.5 'DM Mono'; min-height:420px} .muted{color:#78867c; font-size:13px} .section-heading{display:flex; justify-content:space-between; gap:10px; align-items:center} .media-list{display:grid; gap:8px; margin-top:16px} .media-item{display:flex; justify-content:space-between; gap:12px; padding:10px; border-radius:7px; background:#e9f0e8; font-size:12px} .media-item small{color:#758379}</style>
