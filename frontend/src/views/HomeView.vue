<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { useProductStore } from '@/stores/productStore'
import ProductList from '@/components/ProductList.vue'

const products = useProductStore()
const clickFile = ref(null)
const clickSyncing = ref(false)
const clickMessage = ref('')
const commissionFile = ref(null)
const commissionSyncing = ref(false)
const commissionMessage = ref('')
const niches = ref([])
const totalPages = computed(() => Math.max(1, Math.ceil((products.total || 0) / (products.limit || 20))))
async function updateModel(id, contentModel) {
  try {
    const updated = await products.updateProduct(id, { content_model: contentModel })
    const item = products.items.find((product) => product.id === id)
    if (item) item.content_model = updated.content_model
  } catch (e) { products.error = e.message }
}
async function remove(id) {
  if (!confirm('Hapus produk ini? Tindakan tidak bisa di-undo.')) return
  try { await products.deleteProduct(id); await products.fetchProducts() } catch (e) { products.error = e.message }
}
let timer
watch(() => [products.filters.status, products.filters.content_model, products.filters.source_category, products.filters.cluster, products.filters.niche_id], () => { products.page = 1; products.fetchProducts() })
watch(() => products.page, () => products.fetchProducts())
watch(() => products.limit, () => { products.page = 1; products.fetchProducts() })
function search() { clearTimeout(timer); timer = setTimeout(() => { products.page = 1; products.fetchProducts() }, 250) }
function prev() { if (products.page > 1) products.page-- }
function next() { if (products.page < totalPages.value) products.page++ }
function selectClickFile(event) { clickFile.value = event.target.files?.[0] || null; clickMessage.value = '' }
async function syncClicks() {
  if (!clickFile.value) return
  clickSyncing.value = true; clickMessage.value = ''
  try {
    const result = await products.importClicks(clickFile.value)
    clickMessage.value = `Sync selesai: ${result.imported} klik baru, ${result.duplicate} duplikat, ${result.matched} produk cocok.`
    await products.fetchProducts()
  } catch (error) { clickMessage.value = error.message } finally { clickSyncing.value = false }
}
function selectCommissionFile(event) { commissionFile.value = event.target.files?.[0] || null; commissionMessage.value = '' }
async function syncCommissions() {
  if (!commissionFile.value) return
  commissionSyncing.value = true; commissionMessage.value = ''
  try {
    const result = await products.importCommissions(commissionFile.value)
    commissionMessage.value = `Sync komisi: ${result.imported} baru, ${result.updated} update, ${result.matched} produk cocok.`
    await products.fetchProducts()
  } catch (error) { commissionMessage.value = error.message } finally { commissionSyncing.value = false }
}
onMounted(async () => { niches.value = await products.fetchNiches().catch(() => []); await products.fetchProducts() })
</script>

<template>
  <section class="hero"><div><h1>Turn messy product data into ready-to-post copy.</h1><p class="hero-copy">Satu workspace untuk menyimpan produk affiliate, merapikan detail yang berantakan, dan mengulang angle caption tanpa mulai dari nol.</p></div><div class="hero-note">Posting tetap manual di X. Session browser yang menentukan akun, bukan aplikasi ini.</div></section>
  <div class="toolbar"><input v-model="products.filters.search" class="input" placeholder="Cari produk, keyword, atau raw text..." @input="search" /><select v-model="products.filters.status" class="select"><option value="">Semua status</option><option value="raw">Raw</option><option value="reformatted">Reformatted</option><option value="ready">Ready</option></select><select v-model="products.filters.content_model" class="select"><option value="">Semua model</option><option value="trending">Trending</option><option value="branded">Branded</option><option value="cheap">Murah</option><option value="capture">Captured (legacy)</option></select><select v-model="products.filters.source_category" class="select"><option value="">Semua sumber</option><option value="import_x">X</option><option value="scrape_shopee">Shopee</option><option value="raw_text">Copas</option></select><select v-model="products.filters.niche_id" class="select"><option value="">Semua niche</option><option v-for="niche in niches" :key="niche.id" :value="niche.id">{{ niche.name }}</option></select><input v-model="products.filters.cluster" class="input" placeholder="Filter cluster" /></div>
  <section class="sync-panel"><div><h3>Sync data klik Shopee</h3><p class="muted">Upload CSV report kapan saja. Klik ID yang sama tidak dihitung ulang.</p></div><input type="file" accept=".csv,text/csv" @change="selectClickFile" /><button class="button-primary" :disabled="!clickFile || clickSyncing" @click="syncClicks">{{ clickSyncing ? 'Menyinkronkan...' : 'Sync Klik CSV' }}</button><span v-if="clickMessage" class="sync-message">{{ clickMessage }}</span></section>
  <section class="sync-panel"><div><h3>Sync komisi Shopee</h3><p class="muted">Upload CSV komisi (event_id, order_status). Otomatis hitung sales/pending/komisi per tracking tag.</p></div><input type="file" accept=".csv,text/csv" @change="selectCommissionFile" /><button class="button-primary" :disabled="!commissionFile || commissionSyncing" @click="syncCommissions">{{ commissionSyncing ? 'Menyinkronkan...' : 'Sync Komisi CSV' }}</button><span v-if="commissionMessage" class="sync-message">{{ commissionMessage }}</span><RouterLink to="/sold" class="button" style="margin-left:8px">Lihat produk terjual →</RouterLink></section>
  <div v-if="products.error" class="error-box">{{ products.error }}</div>
  <div v-else-if="products.loading && !products.items.length" class="loading">Memuat product library...</div>
  <template v-else-if="products.items.length">
    <ProductList :items="products.items" @delete="remove" @update-model="updateModel" />
    <div class="pagination">
      <button class="button" :disabled="products.page <= 1" @click="prev">‹ Prev</button>
      <span class="page-info">Hal {{ products.page }} dari {{ totalPages }} · {{ products.total }} produk</span>
      <button class="button" :disabled="products.page >= totalPages" @click="next">Next ›</button>
      <select v-model.number="products.limit" class="select" style="width:auto; margin-left:8px">
        <option :value="10">10 / hal</option>
        <option :value="20">20 / hal</option>
        <option :value="50">50 / hal</option>
      </select>
    </div>
  </template>
  <div v-else class="empty"><h3>Belum ada produk</h3><p>Mulai dari data Shopee yang masih berantakan.</p><RouterLink to="/products/new" class="button-primary">Tambah produk pertama</RouterLink></div>
</template>

<style scoped>.pagination{ display:flex; align-items:center; gap:10px; justify-content:center; margin-top:18px; flex-wrap:wrap} .page-info{ font:12px 'DM Mono'; color:#6b7a6e}</style>
