<script setup>
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { apiRequest, useProductStore } from '@/stores/productStore'

const products = useProductStore()
const items = ref([])
const niches = ref([])
const productTypes = ref([])
const selectedNiche = ref('')
const researchNiche = ref('')
const researchCategory = ref('')
const researchKeyword = ref('')
const researchQuery = ref('')
const search = ref('')
const loading = ref(false)
const saving = ref(false)
const message = ref('')
const error = ref('')
const form = ref({ canonical_url: '', original_text: '', author_handle: '', source_query: '', published_at: null, stats: null, media: [], niche_ids: [], product_type_ids: [] })

const researchCatalog = {
  'Sukses & Kesuksesan': {
    Edukasi: ['kebiasaan orang sukses', 'mindset sukses', 'skill kerja'],
    'Tips Praktis': ['cara sukses', 'produktivitas', 'mencapai target'],
    'Cerita & Pengalaman': ['perjalanan sukses', 'gagal lalu bangkit', 'pelajaran hidup'],
    'Opini & Debat': ['kerja keras vs kerja cerdas', 'karier dan uang', 'definisi sukses'],
  },
  'Fashion Pria': {
    'Outfit & Styling': ['outfit pria', 'gaya pria', 'padu padan pria'],
    'Tips Praktis': ['cara berpakaian pria', 'kesalahan berpakaian', 'fit baju pria'],
    Rekomendasi: ['sepatu pria', 'celana pria', 'kaos pria'],
    'Opini & Tren': ['tren fashion pria', 'fashion pria lokal', 'gaya pria minimalis'],
  },
  'Hubungan / Relasi Pria Wanita': {
    Edukasi: ['komunikasi dalam hubungan', 'attachment style', 'bahasa cinta'],
    'Tips Praktis': ['cara komunikasi pasangan', 'cara pdkt', 'hubungan sehat'],
    'Masalah & Pain Point': ['red flags hubungan', 'pasangan menjauh', 'susah move on'],
    'Opini & Debat': ['kencan dan relasi', 'standar pasangan', 'pria wanita zaman sekarang'],
  },
  'Gym, Lari & Exercise': {
    Edukasi: ['progressive overload', 'protein dan otot', 'recovery olahraga'],
    'Tips Praktis': ['gym pemula', 'workout di rumah', 'lari untuk pemula'],
    'Kesalahan & Cedera': ['kesalahan gym pemula', 'cedera gym', 'overtraining'],
    'Cerita & Progress': ['progress gym', 'transformasi badan', 'lari 5k'],
  },
  Affiliate: {
    Edukasi: ['affiliate marketing', 'cara kerja affiliate', 'konten jualan'],
    'Tips Praktis': ['tips jualan online', 'cara promosi produk', 'copywriting jualan'],
    'Masalah & Pain Point': ['jualan sepi', 'susah closing', 'produk tidak laku'],
    'Cerita & Studi Kasus': ['penghasilan affiliate', 'pengalaman jualan online', 'jualan dari rumah'],
  },
}

const currentResearchNiche = () => niches.value.find((item) => item.id === researchNiche.value)
const researchCategories = () => Object.keys(researchCatalog[currentResearchNiche()?.name] || {})
const researchKeywords = () => researchCatalog[currentResearchNiche()?.name]?.[researchCategory.value] || []
const suggestedQueries = () => {
  const niche = niches.value.find((item) => item.id === researchNiche.value)
  return niche ? Object.values(researchCatalog[niche.name] || {}).flat() : []
}

function buildResearchQuery() {
  const keyword = researchKeyword.value.trim() || researchQuery.value.trim() || suggestedQueries()[0] || ''
  if (!keyword) return ''
  const terms = keyword.split(/\s+OR\s+/i).map((term) => term.trim()).filter(Boolean)
  const base = terms.length > 1 ? `(${terms.join(' OR ')})` : (keyword.includes(' ') ? `"${keyword}"` : keyword)
  return `${base} lang:id -is:retweet`
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const params = new URLSearchParams({ page: 1, limit: 100, platform: 'x' })
    if (selectedNiche.value) params.set('content_niche_id', selectedNiche.value)
    if (search.value.trim()) params.set('search', search.value.trim())
    const [content, nicheData, types] = await Promise.all([
      apiRequest(`/api/content-items?${params}`),
      apiRequest('/api/content-niches'),
      products.fetchNiches(),
    ])
    items.value = content.items || []
    niches.value = nicheData || []
    productTypes.value = types || []
  } catch (e) { error.value = e.message } finally { loading.value = false }
}

async function save() {
  if (!form.value.canonical_url.trim() || !form.value.original_text.trim()) return
  saving.value = true; error.value = ''; message.value = ''
  try {
    await apiRequest('/api/content-items', { method: 'POST', body: JSON.stringify(form.value) })
    form.value = { canonical_url: '', original_text: '', author_handle: '', source_query: '', published_at: null, stats: null, media: [], niche_ids: [], product_type_ids: [] }
    message.value = 'Konten berhasil disimpan ke bank.'
    await load()
  } catch (e) { error.value = e.message } finally { saving.value = false }
}

function openResearch(filter = 'top') {
  const query = buildResearchQuery()
  if (!query) return
  const params = new URLSearchParams({ q: query, src: 'typed_query', f: filter })
  window.open(`https://x.com/search?${params.toString()}`, '_blank', 'noopener')
  form.value.source_query = query
  if (researchNiche.value && !form.value.niche_ids.includes(researchNiche.value)) form.value.niche_ids.push(researchNiche.value)
}

function receiveCapture(event) {
  if (event.data?.type !== 'AFFILIATOR_X_RESEARCH_CAPTURE' || !event.data.item) return
  const item = event.data.item
  form.value.canonical_url = item.canonical_url || ''
  form.value.original_text = item.original_text || ''
  form.value.author_handle = item.author_handle ? `@${item.author_handle}` : ''
  form.value.published_at = item.published_at || null
  form.value.stats = item.stats || null
  form.value.media = item.media || []
  message.value = `Post ditangkap${form.value.media.length ? ` · ${form.value.media.length} media` : ''}. Review lalu simpan.`
  error.value = ''
}

watch([selectedNiche, search], load)
watch(researchNiche, () => { researchCategory.value = ''; researchKeyword.value = ''; researchQuery.value = '' })
onMounted(() => { load(); window.addEventListener('message', receiveCapture) })
onBeforeUnmount(() => window.removeEventListener('message', receiveCapture))
</script>

<template>
  <div class="content-page">
    <section class="hero-row"><div><h1>Bank konten</h1><p class="muted">Riset konten X per niche. Simpan sumber asli, lalu kurasi dan reformat nanti.</p></div><span class="roadmap-badge">X / RESEARCH</span></section>
    <section class="panel research-form">
      <div class="section-heading"><div><h2>1. Riset konten X</h2><p class="muted">Mulai dari niche dan demand. Cari posting populer dulu, lalu pilih konten yang layak masuk bank.</p></div></div>
      <div class="form-grid"><select v-model="researchNiche" class="select"><option value="">Pilih niche konten</option><option v-for="niche in niches" :key="niche.id" :value="niche.id">{{ niche.name }}</option></select><select v-model="researchCategory" class="select" :disabled="!researchCategories().length"><option value="">Pilih kategori konten</option><option v-for="category in researchCategories()" :key="category" :value="category">{{ category }}</option></select></div>
      <div v-if="researchKeywords().length" class="query-options"><button v-for="keyword in researchKeywords()" :key="keyword" class="query-chip" type="button" :class="{ selected: researchKeyword === keyword }" @click="researchKeyword = keyword; researchQuery = ''">{{ keyword }}</button></div>
      <input v-model="researchQuery" class="input" placeholder="Keyword/query manual (opsional)" @keyup.enter="openResearch" />
      <p v-if="buildResearchQuery()" class="query-preview">Query: <code>{{ buildResearchQuery() }}</code></p>
      <div class="research-actions"><button class="button-primary" :disabled="!buildResearchQuery()" @click="openResearch('top')">Buka Populer ↗</button><button class="button-secondary" :disabled="!buildResearchQuery()" @click="openResearch('live')">Buka Terbaru ↗</button><button class="button-secondary" :disabled="!buildResearchQuery()" @click="openResearch('images')">Buka Media ↗</button></div>
      <p class="muted">Pilih post yang relevan di X, lalu klik extension <strong>X Research</strong> untuk menangkap URL, teks, media, waktu, dan statistik yang terlihat. Review dulu sebelum disimpan.</p>
    </section>
    <section class="panel content-form">
      <div class="section-heading"><div><h2>2. Simpan hasil riset</h2><p class="muted">Konten asli tidak diubah. Niche konten dan Jenis Barang dapat dipilih lebih dari satu.</p></div></div>
      <input v-model="form.canonical_url" class="input" placeholder="URL post X" />
      <div class="form-grid"><input v-model="form.author_handle" class="input" placeholder="@author (opsional)" /><input v-model="form.source_query" class="input" placeholder="Query riset (opsional)" /></div>
      <textarea v-model="form.original_text" class="textarea" rows="6" placeholder="Paste konten asli di sini..."></textarea>
      <div class="label-group"><strong>Niche konten</strong><div class="tag-options"><label v-for="niche in niches" :key="niche.id" class="tag-option"><input v-model="form.niche_ids" type="checkbox" :value="niche.id" /> {{ niche.name }}</label></div></div>
      <div class="label-group"><strong>Jenis Barang (opsional)</strong><div class="tag-options"><label v-for="type in productTypes" :key="type.id" class="tag-option"><input v-model="form.product_type_ids" type="checkbox" :value="type.id" /> {{ type.name }}</label></div></div>
      <button class="button-primary" :disabled="saving || !form.canonical_url.trim() || !form.original_text.trim()" @click="save">{{ saving ? 'Menyimpan...' : 'Simpan ke bank konten' }}</button>
      <p v-if="message" class="save-notice">✓ {{ message }}</p><p v-if="error" class="error-box">{{ error }}</p>
    </section>
    <section class="panel"><div class="section-heading"><div><h2>Konten tersimpan</h2><p class="muted">{{ items.length }} konten pada tampilan ini</p></div><div class="filter-row"><select v-model="selectedNiche" class="select"><option value="">Semua niche</option><option v-for="niche in niches" :key="niche.id" :value="niche.id">{{ niche.name }}</option></select><input v-model="search" class="input" placeholder="Cari konten..." /></div></div><p v-if="loading" class="muted">Memuat bank konten...</p><div v-else-if="items.length" class="content-list"><article v-for="item in items" :key="item.id" class="content-card"><div class="content-card-head"><span class="status-pill">{{ item.status }}</span><a :href="item.canonical_url" target="_blank" rel="noreferrer">Buka sumber ↗</a></div><p>{{ item.original_text }}</p><small>{{ item.author_handle || 'Author tidak dicatat' }} · {{ new Date(item.created_at).toLocaleString('id-ID') }}</small></article></div><p v-else class="empty-state">Belum ada konten. Simpan post pertama dari hasil riset.</p></section>
  </div>
</template>

<style scoped>
.content-page{display:grid;gap:18px}.hero-row{display:flex;justify-content:space-between;align-items:end;gap:16px}.hero-row h1{font-size:48px;margin:0}.roadmap-badge,.status-pill{font:12px 'DM Mono';color:#1f6b4f;background:#e7f1e8;border-radius:999px;padding:7px 10px}.content-form,.research-form{display:grid;gap:12px}.form-grid{display:grid;grid-template-columns:1fr 1fr;gap:10px}.textarea{width:100%;resize:vertical;border:1px solid #cedbd0;border-radius:8px;padding:12px;background:#fffdf9;font:14px 'DM Mono';box-sizing:border-box}.label-group{display:grid;gap:8px;color:#52655a}.tag-options,.query-options,.research-actions{display:flex;flex-wrap:wrap;gap:8px}.tag-option,.query-chip{padding:8px 10px;border:1px solid #d9e2d8;border-radius:8px;background:#f8fbf7}.tag-option input{accent-color:#1f6b4f}.query-chip{cursor:pointer;color:#1f6b4f}.query-chip.selected{border-color:#1f6b4f;background:#e7f1e8}.query-preview{margin:0;padding:10px 12px;border-radius:8px;background:#eef5ed;color:#52655a;overflow:auto}.query-preview code{font-family:'DM Mono';color:#1f6b4f}.filter-row{display:flex;gap:8px}.filter-row .input{width:180px}.content-list{display:grid;gap:10px}.content-card{border:1px solid #d9e2d8;border-radius:10px;padding:14px;background:#fbfcf8}.content-card-head{display:flex;justify-content:space-between;align-items:center}.content-card a{color:#1f6b4f}.content-card p{white-space:pre-wrap;line-height:1.5;max-height:130px;overflow:hidden}.content-card small{color:#78867c}.empty-state{padding:30px;text-align:center;color:#78867c}.error-box{padding:10px;background:#fff0ed;color:#a84f43}.save-notice{color:#176b4f;font-weight:600}@media(max-width:700px){.form-grid,.hero-row{grid-template-columns:1fr;display:grid}.filter-row{flex-direction:column}.filter-row .input{width:auto}}
</style>
