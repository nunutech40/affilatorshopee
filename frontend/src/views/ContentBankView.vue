<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { apiRequest, useProductStore } from '@/stores/productStore'

const products = useProductStore()
const items = ref([])
const niches = ref([])
const productTypes = ref([])
const selectedNiche = ref('')
const selectedPlatform = ref('')
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
    Edukasi: ['kebiasaan orang sukses OR mindset sukses', 'skill kerja OR skill masa depan', 'financial freedom OR kebebasan finansial'],
    'Tips Praktis': ['cara sukses OR cara berkembang', 'produktivitas OR manajemen waktu', 'mencapai target OR konsisten'],
    'Cerita & Pengalaman': ['perjalanan sukses OR proses sukses', 'gagal lalu bangkit OR belajar dari kegagalan', 'pelajaran hidup OR pengalaman hidup'],
    'Opini & Debat': ['kerja keras OR kerja cerdas', 'karier OR uang dan karier', 'definisi sukses OR arti sukses'],
  },
  'Fashion Pria': {
    'Outfit & Styling': ['outfit pria OR outfit cowok', 'gaya pria OR style pria', 'padu padan pria OR mix and match pria'],
    'Tips Praktis': ['cara berpakaian pria OR tips berpakaian pria', 'kesalahan berpakaian OR fashion mistakes', 'fit baju pria OR ukuran baju pria'],
    Rekomendasi: ['sepatu pria OR sneakers pria', 'celana pria OR chino pria', 'kaos pria OR kemeja pria'],
    'Opini & Tren': ['tren fashion pria OR trend fashion pria', 'fashion pria lokal OR brand lokal pria', 'gaya pria minimalis OR outfit minimalis pria'],
  },
  'Hubungan / Relasi Pria Wanita': {
    Edukasi: ['komunikasi dalam hubungan OR komunikasi pasangan', 'attachment style OR gaya keterikatan', 'bahasa cinta OR love language'],
    'Tips Praktis': ['cara komunikasi pasangan OR komunikasi yang sehat', 'cara pdkt OR tips pdkt', 'hubungan sehat OR relationship sehat'],
    'Masalah & Pain Point': ['red flags hubungan OR tanda hubungan toxic', 'pasangan menjauh OR pasangan berubah', 'susah move on OR cara move on'],
    'Opini & Debat': ['kencan dan relasi OR dating dan relasi', 'standar pasangan OR standar dalam hubungan', 'pria wanita zaman sekarang OR hubungan zaman sekarang'],
  },
  'Gym, Lari & Exercise': {
    Edukasi: ['progressive overload OR latihan beban', 'protein dan otot OR protein untuk otot', 'recovery olahraga OR pemulihan olahraga'],
    'Tips Praktis': ['gym pemula OR latihan gym pemula', 'workout di rumah OR home workout', 'lari untuk pemula OR tips lari pemula'],
    'Kesalahan & Cedera': ['kesalahan gym pemula OR kesalahan saat gym', 'cedera gym OR cedera olahraga', 'overtraining OR latihan berlebihan'],
    'Cerita & Progress': ['progress gym OR progress latihan', 'transformasi badan OR body transformation', 'lari 5k OR latihan 5k'],
  },
  Affiliate: {
    Edukasi: ['affiliate marketing OR pemasaran affiliate', 'cara kerja affiliate OR cara jadi affiliate', 'konten jualan OR konten promosi'],
    'Tips Praktis': ['tips jualan online OR cara jualan online', 'cara promosi produk OR strategi promosi', 'copywriting jualan OR caption jualan'],
    'Masalah & Pain Point': ['jualan sepi OR toko sepi', 'susah closing OR susah jualan', 'produk tidak laku OR barang tidak laku'],
    'Cerita & Studi Kasus': ['penghasilan affiliate OR hasil affiliate', 'pengalaman jualan online OR cerita jualan online', 'jualan dari rumah OR bisnis dari rumah'],
  },
}

const currentResearchNiche = () => niches.value.find((item) => item.id === researchNiche.value)
const researchCategories = () => Object.keys(researchCatalog[currentResearchNiche()?.name] || {})
const researchKeywords = () => researchCatalog[currentResearchNiche()?.name]?.[researchCategory.value] || []
const customResearchQuery = computed(() => researchQuery.value.trim().length > 0)
const suggestedQueries = () => {
  const niche = niches.value.find((item) => item.id === researchNiche.value)
  return niche ? Object.values(researchCatalog[niche.name] || {}).flat() : []
}

function buildResearchQuery() {
  const custom = researchQuery.value.trim()
  const keyword = custom || researchKeyword.value.trim() || suggestedQueries()[0] || ''
  if (!keyword) return ''
  if (custom) return appendResearchFilters(keyword)
  const terms = keyword.split(/\s+OR\s+/i).map((term) => term.trim()).filter(Boolean)
  const base = terms.length > 1 ? `(${terms.map(formatResearchTerm).join(' OR ')})` : formatResearchTerm(keyword)
  return appendResearchFilters(base)
}

function formatResearchTerm(term) {
  // Preset terms are intentionally broad: exact phrases often return zero
  // results in X when the wording differs slightly. Users can still enter
  // an exact phrase in custom mode by typing quotation marks themselves.
  return term
}

function appendResearchFilters(query) {
  const filters = []
  if (!/\blang:/i.test(query)) filters.push('lang:in')
  if (!/(?:^|\s)-is:retweet\b/i.test(query)) filters.push('-is:retweet')
  if (!/(?:^|\s)-is:reply\b/i.test(query)) filters.push('-is:reply')
  return [query, ...filters].join(' ')
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const params = new URLSearchParams({ page: 1, limit: 100 })
    if (selectedPlatform.value) params.set('platform', selectedPlatform.value)
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
watch(selectedPlatform, load)

function formatCount(value) { if (value === undefined || value === null) return '—'; return new Intl.NumberFormat('id-ID', { notation: 'compact', maximumFractionDigits: 1 }).format(value) }
function popularity(item) { const s = item.latest_stats; if (!s) return 'Belum ada statistik'; return `♥ ${formatCount(s.like_count)} · repost ${formatCount(s.repost_count)} · reply ${formatCount(s.reply_count)}${s.view_count ? ` · views ${formatCount(s.view_count)}` : ''}` }
watch(researchNiche, () => { researchCategory.value = ''; researchKeyword.value = ''; researchQuery.value = '' })
onMounted(() => { load(); window.addEventListener('message', receiveCapture) })
onBeforeUnmount(() => window.removeEventListener('message', receiveCapture))
</script>

<template>
  <div class="content-page">
    <section class="hero-row"><div><h1>Bank konten</h1><p class="muted">Semua konten riset tersimpan di sini. Filter berdasarkan niche atau cari teks asli.</p></div><RouterLink class="button-primary" to="/content-bank/capture">+ Tangkap konten</RouterLink></section>
    <section v-if="false" class="panel research-form">
      <div class="section-heading"><div><h2>1. Riset konten X</h2><p class="muted">Mulai dari niche dan demand. Cari posting populer dulu, lalu pilih konten yang layak masuk bank.</p></div></div>
      <div class="form-grid"><select v-model="researchNiche" class="select" :disabled="customResearchQuery"><option value="">Pilih niche konten</option><option v-for="niche in niches" :key="niche.id" :value="niche.id">{{ niche.name }}</option></select><select v-model="researchCategory" class="select" :disabled="customResearchQuery || !researchCategories().length"><option value="">Pilih kategori konten</option><option v-for="category in researchCategories()" :key="category" :value="category">{{ category }}</option></select></div>
      <div v-if="researchKeywords().length" class="query-options"><button v-for="keyword in researchKeywords()" :key="keyword" class="query-chip" type="button" :disabled="customResearchQuery" :class="{ selected: researchKeyword === keyword }" @click="researchKeyword = keyword; researchQuery = ''">{{ keyword }}</button></div>
      <div class="custom-query-row"><input v-model="researchQuery" class="input" placeholder="Keyword/query custom (opsional — mengambil alih preset)" @keyup.enter="openResearch" /><button v-if="customResearchQuery" class="button-secondary" type="button" @click="researchQuery = ''">Pakai preset</button></div>
      <p v-if="customResearchQuery" class="muted custom-query-note">Mode custom aktif: niche, kategori, dan keyword bawaan dinonaktifkan. Query custom akan dipakai sebagai dasar.</p>
      <p v-if="buildResearchQuery()" class="query-preview">Query: <code>{{ buildResearchQuery() }}</code></p>
      <div class="research-actions"><button class="button-primary" :disabled="!buildResearchQuery()" @click="openResearch('top')">Buka Populer ↗</button><button class="button-secondary" :disabled="!buildResearchQuery()" @click="openResearch('live')">Buka Terbaru ↗</button><button class="button-secondary" :disabled="!buildResearchQuery()" @click="openResearch('media')">Buka Media ↗</button></div>
      <p class="muted">Pilih post yang relevan di X, lalu klik extension <strong>X Research</strong> untuk menangkap URL, teks, media, waktu, dan statistik yang terlihat. Review dulu sebelum disimpan.</p>
    </section>
    <section v-if="false" class="panel content-form">
      <div class="section-heading"><div><h2>2. Simpan hasil riset</h2><p class="muted">Konten asli tidak diubah. Niche konten dan Jenis Barang dapat dipilih lebih dari satu.</p></div></div>
      <input v-model="form.canonical_url" class="input" placeholder="URL post X" />
      <div class="form-grid"><input v-model="form.author_handle" class="input" placeholder="@author (opsional)" /><input v-model="form.source_query" class="input" placeholder="Query riset (opsional)" /></div>
      <textarea v-model="form.original_text" class="textarea" rows="6" placeholder="Paste konten asli di sini..."></textarea>
      <div class="label-group"><strong>Niche konten</strong><div class="tag-options"><label v-for="niche in niches" :key="niche.id" class="tag-option"><input v-model="form.niche_ids" type="checkbox" :value="niche.id" /> {{ niche.name }}</label></div></div>
      <div class="label-group"><strong>Jenis Barang (opsional)</strong><div class="tag-options"><label v-for="type in productTypes" :key="type.id" class="tag-option"><input v-model="form.product_type_ids" type="checkbox" :value="type.id" /> {{ type.name }}</label></div></div>
      <button class="button-primary" :disabled="saving || !form.canonical_url.trim() || !form.original_text.trim()" @click="save">{{ saving ? 'Menyimpan...' : 'Simpan ke bank konten' }}</button>
      <p v-if="message" class="save-notice">✓ {{ message }}</p><p v-if="error" class="error-box">{{ error }}</p>
    </section>
    <section class="panel"><div class="section-heading"><div><h2>Konten tersimpan</h2><p class="muted">{{ items.length }} konten pada tampilan ini · urutkan dan kurasi berdasarkan sinyal popularitas</p></div><div class="filter-row"><select v-model="selectedPlatform" class="select"><option value="">Semua sumber</option><option value="x">X</option><option value="threads">Threads</option><option value="facebook">Facebook</option></select><select v-model="selectedNiche" class="select"><option value="">Semua niche</option><option v-for="niche in niches" :key="niche.id" :value="niche.id">{{ niche.name }}</option></select><input v-model="search" class="input" placeholder="Cari konten..." /></div></div><p v-if="loading" class="muted">Memuat bank konten...</p><p v-if="error" class="error-box">{{ error }}</p><div v-else-if="items.length" class="content-list"><RouterLink v-for="item in items" :key="item.id" :to="`/content-bank/${item.id}`" class="content-card"><div class="content-card-head"><span><span class="status-pill">{{ item.status }}</span><span class="source-pill">{{ item.platform }}</span></span><span class="open-link">Detail →</span></div><h3>{{ (item.original_text || 'Konten tanpa teks').split('\\n')[0].slice(0, 120) }}</h3><p>{{ item.original_text }}</p><div class="popularity">{{ popularity(item) }}</div><small>{{ item.author_handle || 'Author tidak dicatat' }} · {{ item.published_at ? new Date(item.published_at).toLocaleString('id-ID') : 'tanggal tidak dicatat' }}<span v-if="item.media?.length"> · {{ item.media.length }} media</span><span v-if="item.variants?.length"> · {{ item.variants.length }} varian</span></small></RouterLink></div><p v-else-if="!loading" class="empty-state">Belum ada konten. Mulai dari halaman tangkap konten.</p></section>
  </div>
</template>

<style scoped>
.content-page{display:grid;gap:18px}.hero-row{display:flex;justify-content:space-between;align-items:end;gap:16px}.hero-row h1{font-size:48px;margin:0}.roadmap-badge,.status-pill,.source-pill{font:12px 'DM Mono';color:#1f6b4f;background:#e7f1e8;border-radius:999px;padding:7px 10px}.source-pill{margin-left:8px;background:#f2eee1;color:#7d6a3e}.content-form,.research-form{display:grid;gap:12px}.form-grid{display:grid;grid-template-columns:1fr 1fr;gap:10px}.custom-query-row{display:flex;gap:8px;align-items:center}.custom-query-row .input{flex:1}.custom-query-note{margin:0}.textarea{width:100%;resize:vertical;border:1px solid #cedbd0;border-radius:8px;padding:12px;background:#fffdf9;font:14px 'DM Mono';box-sizing:border-box}.label-group{display:grid;gap:8px;color:#52655a}.tag-options,.query-options,.research-actions{display:flex;flex-wrap:wrap;gap:8px}.tag-option,.query-chip{padding:8px 10px;border:1px solid #d9e2d8;border-radius:8px;background:#f8fbf7}.tag-option input{accent-color:#1f6b4f}.query-chip{cursor:pointer;color:#1f6b4f}.query-chip.selected{border-color:#1f6b4f;background:#e7f1e8}.query-chip:disabled{cursor:not-allowed;opacity:.55}.query-preview{margin:0;padding:10px 12px;border-radius:8px;background:#eef5ed;color:#52655a;overflow:auto}.query-preview code{font-family:'DM Mono';color:#1f6b4f}.filter-row{display:flex;gap:8px}.filter-row .input{width:180px}.content-list{display:grid;gap:10px}.content-card{display:block;border:1px solid #d9e2d8;border-radius:10px;padding:16px;background:#fbfcf8;color:inherit;text-decoration:none;transition:transform .15s,border-color .15s}.content-card:hover{transform:translateY(-2px);border-color:#1f6b4f}.content-card-head{display:flex;justify-content:space-between;align-items:center}.open-link{color:#1f6b4f;font-weight:600}.content-card h3{margin:14px 0 8px;font-size:20px}.content-card p{white-space:pre-wrap;line-height:1.5;max-height:110px;overflow:hidden}.content-card small,.popularity{color:#78867c}.popularity{font-weight:600;margin:10px 0}.empty-state{padding:30px;text-align:center;color:#78867c}.error-box{padding:10px;background:#fff0ed;color:#a84f43}.save-notice{color:#176b4f;font-weight:600}@media(max-width:700px){.form-grid,.hero-row{grid-template-columns:1fr;display:grid}.filter-row,.custom-query-row{flex-direction:column;align-items:stretch}.filter-row .input{width:auto}}
</style>
