<script setup>
import { onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { apiRequest, useProductStore } from '@/stores/productStore'
import ShareButton from '@/components/ShareButton.vue'
import PostLogForm from '@/components/PostLogForm.vue'

const route = useRoute()
const router = useRouter()
const products = useProductStore()
const product = ref(null)
const logs = ref([])
const media = ref([])
const loading = ref(true)
const saving = ref(false)
const savingLink = ref(false)
const aiAction = ref('')
const error = ref('')
const editText = ref('')
const selectedContentModel = ref('')
const shopeeLink = ref('')
const imageURL = ref('')
const videoURL = ref('')
const copiedTag = ref(false)
const saveNotice = ref('')
const previewFile = ref(null)
const selectedMedia = ref([])

function getPromo(p) {
  if (!p) return ''
  const text = p.reformatted_text || ''
  if (!p.shopee_link) return text
  return text.replace(/https?:\/\/(?:www\.)?(?:s\.shopee\.co\.id|shopee\.co\.id)\/[^\s)\]]+/gi, p.shopee_link)
}

watch(product, (p) => { if (p) { editText.value = getPromo(p); selectedContentModel.value = p.content_model || 'cheap'; shopeeLink.value = p.shopee_link || '' } })

async function load() { loading.value = true; error.value = ''; try { product.value = await products.getProduct(route.params.id); const logData = await apiRequest(`/api/post-logs?product_id=${route.params.id}`); logs.value = logData.items; media.value = await apiRequest(`/api/products/${route.params.id}/media`); if (route.query.ai === 'failed') error.value = 'Produk tersimpan sebagai raw text. Reformat AI bisa dicoba ulang dari tombol di atas.' } catch (e) { error.value = e.message } finally { loading.value = false } }

async function saveReformat() {
  saving.value = true
  aiAction.value = 'Menyimpan promo text...'
  saveNotice.value = ''
  error.value = ''
  try { product.value = await products.updateProduct(route.params.id, { reformatted_text: editText.value }); editText.value = getPromo(product.value) } catch (e) { error.value = e.message } finally { saving.value = false }
  if (!error.value) saveNotice.value = 'Promo text berhasil disimpan.'
  aiAction.value = ''
}

async function resetReformat() {
  saving.value = true
  aiAction.value = 'Mereset promo text...'
  error.value = ''
  try {
    product.value = await products.updateProduct(route.params.id, { reset_reformatted: true })
    editText.value = ''
  } catch (e) { error.value = e.message } finally { saving.value = false; aiAction.value = '' }
}

async function updateContentModel() {
  try { product.value = await products.updateProduct(route.params.id, { content_model: selectedContentModel.value }) } catch (e) { error.value = e.message }
}

async function saveShopeeLink() {
  savingLink.value = true
  error.value = ''
  try {
    product.value = await products.updateProduct(route.params.id, { shopee_link: shopeeLink.value.trim() })
  } catch (e) { error.value = e.message } finally { savingLink.value = false }
}

async function copyTrackingTag() {
  try {
    await navigator.clipboard.writeText(product.value.tracking_tag)
    copiedTag.value = true
    setTimeout(() => { copiedTag.value = false }, 1800)
  } catch (e) { error.value = 'Tag tidak bisa disalin: ' + e.message }
}

async function addMedia() {
  const images = imageURL.value.trim() ? [imageURL.value.trim()] : []
  const video = videoURL.value.trim() || null
  if (!images.length && !video) return
  saving.value = true
  error.value = ''
  try {
    const result = await apiRequest(`/api/products/${route.params.id}/media`, { method: 'POST', body: JSON.stringify({ image_urls: images, video_url: video }) })
    media.value = [...media.value, ...(result.downloaded || [])]
    imageURL.value = ''
    videoURL.value = ''
    if (result.failed?.length) error.value = result.failed.map((item) => item.message).join('; ')
  } catch (e) { error.value = e.message } finally { saving.value = false }
}

async function removeMedia(file) {
  if (!confirm(`Hapus ${file.filename}?`)) return
  saving.value = true
  try {
    await apiRequest(`/api/products/${route.params.id}/media/${file.id}`, { method: 'DELETE' })
    media.value = media.value.filter((item) => item.id !== file.id)
    selectedMedia.value = selectedMedia.value.filter((id) => id !== file.id)
  } catch (e) { error.value = e.message } finally { saving.value = false }
}
function toggleMedia(id) { selectedMedia.value = selectedMedia.value.includes(id) ? selectedMedia.value.filter((item) => item !== id) : [...selectedMedia.value, id] }
function selectAllMedia() { selectedMedia.value = selectedMedia.value.length === media.value.length ? [] : media.value.map((file) => file.id) }
async function removeSelectedMedia() {
  if (!selectedMedia.value.length || !confirm(`Hapus ${selectedMedia.value.length} media terpilih?`)) return
  saving.value = true
  aiAction.value = `Menghapus ${selectedMedia.value.length} media...`
  error.value = ''
  try {
    for (const id of selectedMedia.value) await apiRequest(`/api/products/${route.params.id}/media/${id}`, { method: 'DELETE' })
    media.value = media.value.filter((file) => !selectedMedia.value.includes(file.id))
    selectedMedia.value = []
  } catch (e) { error.value = e.message } finally { saving.value = false; aiAction.value = '' }
}
function mediaURL(file) { return `${window.location.origin}/api/products/${product.value.id}/media/${file.id}/file` }
function openPreview(file) { previewFile.value = file }
function closePreview() { previewFile.value = null }

async function reformatVariant() {
  saving.value = true
  aiAction.value = 'AI sedang membuat varian caption...'
  error.value = ''
  try {
    await products.updateProduct(route.params.id, { content_model: selectedContentModel.value })
    await products.reformat([route.params.id], localStorage.getItem('ai_model') || 'stealth/ox-alpha', true)
    await load()
  } catch (e) { error.value = e.message } finally { saving.value = false; aiAction.value = '' }
}

async function reformatAI() {
  saving.value = true
  aiAction.value = 'AI sedang membuat promo text dari raw text...'
  error.value = ''
  try {
    await products.updateProduct(route.params.id, { content_model: selectedContentModel.value })
    await products.reformat([route.params.id], localStorage.getItem('ai_model') || 'stealth/ox-alpha', false)
    await load()
  } catch (e) { error.value = `Reformat AI gagal. Produk tetap tersimpan sebagai raw text: ${e.message}` } finally { saving.value = false; aiAction.value = '' }
}

async function remove() {
  if (!confirm('Hapus produk ini permanen?')) return
  try { await products.deleteProduct(route.params.id); router.push('/') } catch (e) { error.value = e.message }
}
async function refreshLogs() { const logData = await apiRequest(`/api/post-logs?product_id=${route.params.id}`); logs.value = logData.items; product.value = await products.getProduct(route.params.id); editText.value = getPromo(product.value) }
function mediaURLs() { return [...new Set(media.value.map((file) => `${window.location.origin}/api/products/${product.value.id}/media/${file.id}/file`))] }
function hashtags() { return (editText.value.match(/(^|\s)#\w+/g) || []).map((tag) => tag.trim()) }
function displayProductName(p) {
  if (p?.product_name) return p.product_name
  const lines = String(p?.reformatted_text || '').split(/\r?\n/).map((line) => line.trim()).filter(Boolean)
  return lines.find((line, index) => index > 0 && !/^(?:[💸⚡️👇🔥⭐️✅]|harga|mulai|cek|https?:\/\/|#)/iu.test(line)) || 'Raw product'
}
function modelLabel(value) { return ({ trending: 'Trending', branded: 'Branded', cheap: 'Murah', capture: 'Captured (legacy)' }[value] || 'Content model belum dipilih') }
function sourceLabel(value) { return ({ import_x: 'Import X', scrape_shopee: 'Scrape Shopee', raw_text: 'Raw text' }[value] || 'Raw text') }
onMounted(load)
</script>

<template>
  <div v-if="loading" class="loading">Memuat product...</div>
  <div v-else-if="error && !product" class="error-box">{{ error }}</div>
  <template v-else-if="product">
    <RouterLink to="/" class="back-link">← Kembali ke library</RouterLink>
    <header class="detail-header"><div class="detail-title"><span class="status" :class="product.status">{{ product.status }}</span><h1>{{ displayProductName(product) }}</h1><p class="cluster">Source: {{ sourceLabel(product.source_category) }} · Cluster: {{ product.cluster || 'uncategorized' }} · Content model: {{ modelLabel(product.content_model) }} · {{ product.post_count || 0 }} posting</p><div class="tracking-tag"><span>Tracking tag: <b>{{ product.tracking_tag }}</b></span><button class="button" @click="copyTrackingTag">{{ copiedTag ? 'Copied' : 'Copy tag' }}</button></div></div><div class="detail-actions detail-ai-actions"><select v-model="selectedContentModel" class="select" @change="updateContentModel"><option value="trending">Trending</option><option value="branded">Branded</option><option value="cheap">Murah</option></select><button v-if="!product.reformatted_text || product.status === 'raw'" class="button-primary" :disabled="saving" @click="reformatAI">Reformat AI</button><button class="button" :disabled="saving" @click="reformatVariant">Reformat varian caption</button><button class="button button-danger" @click="remove">Hapus</button></div></header>
    <div v-if="aiAction" class="ai-loading" role="status"><span class="spinner"></span>{{ aiAction }} Jangan klik ulang sampai selesai.</div><div v-if="saveNotice" class="save-notice" role="status">✓ {{ saveNotice }}</div><div v-if="error" class="error-box">{{ error }}</div>
    <div class="stats"><div class="stat"><strong>{{ product.click_count || 0 }}</strong><span>Klik affiliate</span></div><div class="stat"><strong>{{ product.post_count || 0 }}</strong><span>Posting</span></div></div>
    <div class="form-layout">
      <div>
        <section class="panel"><h2>Raw text (asli, read-only)</h2><pre class="raw-pre">{{ product.raw_text }}</pre></section>
        <section class="panel"><h2>Data share</h2><div class="field"><label>Link affiliate Shopee</label><input v-model="shopeeLink" class="input" placeholder="https://s.shopee.co.id/..." /><button class="button-primary" style="margin-top:8px" :disabled="savingLink || !shopeeLink.trim()" @click="saveShopeeLink">{{ savingLink ? 'Menyimpan link...' : 'Simpan link' }}</button></div><div class="field"><label>Tambah image URL</label><input v-model="imageURL" class="input" placeholder="https://.../image.jpg" /></div><div class="field"><label>Tambah video URL</label><input v-model="videoURL" class="input" placeholder="https://.../video.mp4" /></div><button class="button" :disabled="saving || (!imageURL.trim() && !videoURL.trim())" @click="addMedia">Download & tambah media</button></section>
        <section class="panel reformat-panel">
          <div class="section-heading"><div><h2>Promo text (editable)</h2><p class="muted">{{ modelLabel(product.content_model) }} menentukan prompt yang dipakai.</p></div><span class="counter">{{ editText.length }} karakter</span></div>
          <p class="muted">Teks ini akan dibagikan utuh ke X, termasuk hashtag dan link.</p>
          <textarea v-model="editText" class="textarea mono" rows="18" spellcheck="false" placeholder="Isi promo text"></textarea>
          <div style="display:flex; gap:8px; margin-top:12px"><button class="button-primary" :disabled="saving || !editText.trim()" @click="saveReformat">{{ saving ? 'Menyimpan...' : 'Save promo text' }}</button><button class="button" :disabled="saving" @click="resetReformat">{{ saving ? 'Memproses...' : 'Reset' }}</button></div><p v-if="!editText.trim()" class="muted action-hint">Promo text masih kosong. Klik <b>Reformat AI</b> di bagian atas terlebih dahulu.</p>
        </section>
        <section class="panel"><div class="section-heading"><div><h2>Media lokal</h2><p class="muted">Centang file yang ingin dihapus, atau klik gambar/video untuk preview.</p></div><div style="display:flex;gap:8px;align-items:center"><a v-if="media.length" class="card-link" :href="`/api/products/${product.id}/media/download`">Download ZIP →</a><button v-if="media.length" class="button" :disabled="saving" @click="selectAllMedia">{{ selectedMedia.length === media.length ? 'Batal pilih' : 'Pilih semua' }}</button><button v-if="selectedMedia.length" class="button button-danger" :disabled="saving" @click="removeSelectedMedia">Hapus terpilih ({{ selectedMedia.length }})</button></div></div><div v-if="media.length" class="media-grid"><article v-for="file in media" :key="file.id" class="media-card" :class="{ selected: selectedMedia.includes(file.id) }"><label class="media-check"><input type="checkbox" :checked="selectedMedia.includes(file.id)" @change="toggleMedia(file.id)" /><span></span></label><button class="media-preview" type="button" @click="openPreview(file)"><img v-if="file.media_type === 'image'" :src="mediaURL(file)" :alt="file.filename" loading="lazy" /><video v-else :src="mediaURL(file)" muted preload="metadata"></video></button><div class="media-card-footer"><span>{{ file.filename }}</span><button class="button button-danger" :disabled="saving" @click="removeMedia(file)">Hapus</button></div></article></div><p v-else class="muted">Belum ada media yang berhasil di-download.</p></section>
      </div>
      <div><section class="panel"><div class="section-heading"><div><h2>Share ke X</h2><p class="muted">Promo text dan {{ media.length }} media lokal akan dikirim ke extension. Link Shopee tetap dikirim dalam bentuk asli.</p></div></div><div class="caption-box">{{ editText || 'Promo text belum tersedia.' }}</div><div style="margin-top:16px"><ShareButton v-if="editText" :caption="editText" :media="mediaURLs()" /></div></section><PostLogForm v-if="editText" :product-id="product.id" :caption="editText" :hashtags="hashtags()" @saved="refreshLogs" /><section class="panel"><h2>Riwayat posting</h2><div v-if="logs.length" class="log-list"><article v-for="log in logs" :key="log.id" class="log-item"><div class="log-head"><span>{{ log.platform }}</span><span>{{ new Date(log.posted_at).toLocaleString('id-ID') }}</span></div><div class="log-text">{{ log.caption }}</div></article></div><p v-else class="muted">Belum ada catatan posting.</p></section></div>
    </div>
    <div v-if="previewFile" class="media-modal" role="dialog" aria-modal="true" @click.self="closePreview"><button class="modal-close" type="button" @click="closePreview">×</button><img v-if="previewFile.media_type === 'image'" :src="mediaURL(previewFile)" :alt="previewFile.filename" /><video v-else :src="mediaURL(previewFile)" controls autoplay /></div>
  </template>
</template>

<style scoped>.raw-pre{white-space:pre-wrap; max-height:320px; overflow:auto; color:#52655a; font:13px/1.6 'DM Mono'; background:#f9fbf8; padding:14px; border-radius:8px; border:1px solid #e1e8e0} .tracking-tag{display:flex;align-items:center;gap:8px;margin-top:10px;color:#6e7d72;font:12px 'DM Mono';flex-wrap:wrap}.tracking-tag b{color:#1f6b4f} .reformat-panel .mono{font:12px/1.5 'DM Mono'; min-height:420px} .muted{color:#78867c; font-size:13px} .action-hint{margin:10px 0 0}.section-heading{display:flex; justify-content:space-between; gap:10px; align-items:center} .reformat-actions{display:flex; gap:8px; align-items:end; margin:12px 0} .ai-loading{display:flex; align-items:center; gap:10px; margin:14px 0; padding:13px 16px; border-radius:8px; background:#e7f3eb; color:#176b4f; font-weight:600}.save-notice{margin:14px 0;padding:12px 16px;border-radius:8px;background:#e7f3eb;color:#176b4f;font-weight:600} .spinner{width:16px; height:16px; border:2px solid #b8d9c6;border-top-color:#176b4f;border-radius:50%;animation:spin .8s linear infinite} @keyframes spin{to{transform:rotate(360deg)}} .media-grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(150px,1fr));gap:12px;margin-top:16px}.media-card{position:relative;overflow:hidden;border:1px solid #d9e2d8;border-radius:10px;background:#f8fbf7}.media-card.selected{border-color:#1f6b4f;box-shadow:0 0 0 2px #c6dfcf}.media-check{position:absolute;top:8px;left:8px;z-index:2;display:flex;width:25px;height:25px;align-items:center;justify-content:center;border-radius:6px;background:rgba(255,255,255,.9);cursor:pointer}.media-check input{width:17px;height:17px;accent-color:#1f6b4f}.media-preview{display:block;width:100%;height:150px;padding:0;border:0;background:#e8efe7;cursor:zoom-in}.media-preview img,.media-preview video{display:block;width:100%;height:100%;object-fit:cover}.media-card-footer{display:flex;align-items:center;justify-content:space-between;gap:6px;padding:8px;font-size:11px}.media-card-footer span{overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.media-card-footer .button{padding:5px 8px;font-size:11px}.media-modal{position:fixed;inset:0;z-index:1000;display:flex;align-items:center;justify-content:center;padding:40px;background:rgba(15,24,18,.8)}.media-modal img,.media-modal video{max-width:min(92vw,1100px);max-height:88vh;object-fit:contain;border-radius:10px;background:#111}.modal-close{position:absolute;top:20px;right:24px;width:40px;height:40px;border:0;border-radius:50%;background:#fff;color:#1e2721;font-size:28px;cursor:pointer}</style>
