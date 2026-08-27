<script setup>
import { onMounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { useProductStore } from '@/stores/productStore'

const store = useProductStore()
const items = ref([])
const total = ref(0)
const page = ref(1)
const limit = ref(20)
const search = ref('')
const loading = ref(false)
const error = ref('')

const money = (v) => v == null ? '-' : new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(v)
const totalPages = () => Math.max(1, Math.ceil(total.value / limit.value))

async function load() {
  loading.value = true
  error.value = ''
  try {
    const data = await store.fetchSoldProducts(page.value, limit.value, search.value)
    items.value = data.items
    total.value = data.total
  } catch (e) { error.value = e.message } finally { loading.value = false }
}
let timer
function onSearch() { clearTimeout(timer); timer = setTimeout(() => { page.value = 1; load() }, 300) }
function prev() { if (page.value > 1) { page.value--; load() } }
function next() { if (page.value < totalPages()) { page.value++; load() } }

onMounted(load)
watch(limit, () => { page.value = 1; load() })
</script>

<template>
  <section class="hero"><div><h1>Produk terjual</h1><p class="hero-copy">Ringkasan dari laporan komisi Shopee. Tidak semua dari link kamu — tapi bisa jadi ide follow-up. Kalau tag-nya ada di library, ada link langsung.</p></div><div class="hero-note">Sumber: CSV komisi. Klik & komisi di-sync via CSV di Library.</div></section>

  <div class="toolbar"><input v-model="search" class="input" placeholder="Cari tag atau nama produk..." @input="onSearch" /><select v-model.number="limit" class="select" style="width:auto"><option :value="10">10 / hal</option><option :value="20">20 / hal</option><option :value="50">50 / hal</option></select><button class="button" @click="load">Refresh</button></div>

  <div v-if="error" class="error-box">{{ error }}</div>
  <div v-else-if="loading && !items.length" class="loading">Memuat produk terjual...</div>
  <div v-else-if="!items.length" class="empty"><h3>Belum ada data terjual</h3><p>Upload CSV komisi di Library → Sync Komisi CSV.</p></div>

  <div v-else class="sold-list">
    <div class="list-header sold-header"><span>Produk / Tag</span><span>Terjual</span><span>Pesanan</span><span>Komisi</span><span>Link</span></div>
    <article v-for="row in items" :key="row.normalized_tag" class="sold-row" :class="{ matched: row.is_in_library }">
      <div class="sold-main">
        <div class="sold-img-wrap">
          <img v-if="row.image_url" :src="row.image_url" alt="" class="sold-img" loading="lazy" />
          <span v-else class="sold-img placeholder">📦</span>
        </div>
        <div class="sold-text">
          <div class="sold-title">{{ row.product_name || row.item_name || 'Produk terjual' }}</div>
          <div class="sold-sub">tag: <code>{{ row.tracking_tag }}</code> · <span :class="{ 'in-lib': row.is_in_library }">{{ row.is_in_library ? 'ada di library' : 'tidak di library' }}</span><span v-if="row.shop_name"> · {{ row.shop_name }}</span><span v-if="row.item_id"> · ID: {{ row.item_id }}</span></div>
          <div v-if="row.last_ordered_at" class="sold-sub muted">terakhir {{ new Date(row.last_ordered_at).toLocaleDateString('id-ID') }}</div>
        </div>
      </div>
      <span class="sold-qty">{{ row.total_quantity }} pcs</span>
      <span class="sold-orders">{{ row.order_count }} order</span>
      <span class="sold-commission">{{ money(row.total_commission) }}</span>
      <div class="sold-actions">
        <RouterLink v-if="row.is_in_library && row.product_id" :to="`/products/${row.product_id}`" class="button-primary small">Buka di library →</RouterLink>
        <a v-else-if="row.item_id" :href="`https://shopee.co.id/search?keyword=${row.item_id}`" target="_blank" rel="noopener" class="button small">Cari di Shopee →</a>
        <a v-else-if="row.shopee_link" :href="row.shopee_link" target="_blank" rel="noopener" class="button small">Buka Shopee →</a>
        <span v-else class="muted small">—</span>
      </div>
    </article>

    <div class="pagination">
      <button class="button" :disabled="page <= 1" @click="prev">‹ Prev</button>
      <span class="page-info">Hal {{ page }} dari {{ totalPages() }} · {{ total }} produk terjual</span>
      <button class="button" :disabled="page >= totalPages()" @click="next">Next ›</button>
    </div>
  </div>
</template>

<style scoped>
.sold-list{display:grid; gap:8px}
.sold-header{display:none; grid-template-columns:1fr 90px 90px 120px 150px; gap:12px; padding:0 16px 6px; font:600 10px 'DM Mono'; letter-spacing:.08em; text-transform:uppercase; color:#8a978d}
.sold-row{display:grid; grid-template-columns:1fr 90px 90px 120px 150px; gap:12px; align-items:center; padding:14px 16px; background:rgba(255,255,255,.62); border:1px solid #d9ded6; border-radius:10px}
.sold-row.matched{border-color:#b8d8c2; background:rgba(255,255,255,.9)}
.sold-main{display:flex; gap:12px; align-items:center; min-width:0}
.sold-img{width:44px; height:44px; border-radius:7px; object-fit:cover; background:#eef4ee; border:1px solid #d9ded6}
.sold-img.placeholder{display:grid; place-items:center; font-size:18px}
.sold-text{min-width:0}
.sold-title{font:600 13px 'Space Grotesk'; color:#1f2721; white-space:nowrap; overflow:hidden; text-overflow:ellipsis}
.sold-sub{font:11px 'DM Mono'; color:#8a978d; margin-top:2px; white-space:nowrap; overflow:hidden; text-overflow:ellipsis}
.sold-sub.muted{color:#b0bdb2}
.in-lib{color:#1f6b4f; font-weight:700}
.sold-qty{font:700 13px 'Space Grotesk'; color:#1f2721; text-align:center}
.sold-orders{font:12px 'DM Mono'; color:#6b7a6e; text-align:center}
.sold-commission{font:700 13px 'Space Grotesk'; color:#1f6b4f; text-align:right}
.sold-actions{justify-self:end}
.button.small{padding:7px 10px; font-size:12px}
.button-primary.small{padding:7px 10px; font-size:12px}
.muted{color:#8a978d}
@media(max-width:900px){ .sold-header{display:none} .sold-row{grid-template-columns:1fr auto} .sold-qty,.sold-orders,.sold-commission{display:none} }
@media(min-width:901px){ .sold-header{display:grid} }
.pagination{ display:flex; align-items:center; gap:10px; justify-content:center; margin-top:18px; flex-wrap:wrap} .page-info{ font:12px 'DM Mono'; color:#6b7a6e}
</style>
