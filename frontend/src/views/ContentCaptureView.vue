<script setup>
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { apiRequest } from '@/stores/productStore'

const router = useRouter()
const niches = ref([])
const productTypes = ref([])
const researchNiche = ref('')
const researchCategory = ref('')
const researchKeyword = ref('')
const researchQuery = ref('')
const saving = ref(false)
const message = ref('')
const error = ref('')
const form = ref({ canonical_url: '', original_text: '', author_handle: '', source_query: '', published_at: null, stats: null, media: [], niche_ids: [], product_type_ids: [] })
const threadPostCount = ref(1)

const researchCatalog = {
  'Sukses & Kesuksesan': { Edukasi: ['kebiasaan orang sukses OR mindset sukses', 'skill kerja OR skill masa depan', 'financial freedom OR kebebasan finansial'], 'Tips Praktis': ['cara sukses OR cara berkembang', 'produktivitas OR manajemen waktu', 'mencapai target OR konsisten'], 'Cerita & Pengalaman': ['perjalanan sukses OR proses sukses', 'gagal lalu bangkit OR belajar dari kegagalan', 'pelajaran hidup OR pengalaman hidup'], 'Opini & Debat': ['kerja keras OR kerja cerdas', 'karier OR uang dan karier', 'definisi sukses OR arti sukses'] },
  'Fashion Pria': { 'Outfit & Styling': ['outfit pria OR outfit cowok', 'gaya pria OR style pria', 'padu padan pria OR mix and match pria'], 'Tips Praktis': ['cara berpakaian pria OR tips berpakaian pria', 'kesalahan berpakaian OR fashion mistakes', 'fit baju pria OR ukuran baju pria'], Rekomendasi: ['sepatu pria OR sneakers pria', 'celana pria OR chino pria', 'kaos pria OR kemeja pria'], 'Opini & Tren': ['tren fashion pria OR trend fashion pria', 'fashion pria lokal OR brand lokal pria', 'gaya pria minimalis OR outfit minimalis pria'] },
  'Hubungan / Relasi Pria Wanita': { Edukasi: ['komunikasi dalam hubungan OR komunikasi pasangan', 'attachment style OR gaya keterikatan', 'bahasa cinta OR love language'], 'Tips Praktis': ['cara komunikasi pasangan OR komunikasi yang sehat', 'cara pdkt OR tips pdkt', 'hubungan sehat OR relationship sehat'], 'Masalah & Pain Point': ['red flags hubungan OR tanda hubungan toxic', 'pasangan menjauh OR pasangan berubah', 'susah move on OR cara move on'], 'Opini & Debat': ['kencan dan relasi OR dating dan relasi', 'standar pasangan OR standar dalam hubungan', 'pria wanita zaman sekarang OR hubungan zaman sekarang'] },
  'Gym, Lari & Exercise': { Edukasi: ['progressive overload OR latihan beban', 'protein dan otot OR protein untuk otot', 'recovery olahraga OR pemulihan olahraga'], 'Tips Praktis': ['gym pemula OR latihan gym pemula', 'workout di rumah OR home workout', 'lari untuk pemula OR tips lari pemula'], 'Kesalahan & Cedera': ['kesalahan gym pemula OR kesalahan saat gym', 'cedera gym OR cedera olahraga', 'overtraining OR latihan berlebihan'], 'Cerita & Progress': ['progress gym OR progress latihan', 'transformasi badan OR body transformation', 'lari 5k OR latihan 5k'] },
  Affiliate: { Edukasi: ['affiliate marketing OR pemasaran affiliate', 'cara kerja affiliate OR cara jadi affiliate', 'konten jualan OR konten promosi'], 'Tips Praktis': ['tips jualan online OR cara jualan online', 'cara promosi produk OR strategi promosi', 'copywriting jualan OR caption jualan'], 'Masalah & Pain Point': ['jualan sepi OR toko sepi', 'susah closing OR susah jualan', 'produk tidak laku OR barang tidak laku'], 'Cerita & Studi Kasus': ['penghasilan affiliate OR hasil affiliate', 'pengalaman jualan online OR cerita jualan online', 'jualan dari rumah OR bisnis dari rumah'] },
}

const currentNiche = () => niches.value.find((item) => item.id === researchNiche.value)
const categories = () => Object.keys(researchCatalog[currentNiche()?.name] || {})
const keywords = () => researchCatalog[currentNiche()?.name]?.[researchCategory.value] || []
const customMode = () => researchQuery.value.trim().length > 0
function query() {
  const custom = researchQuery.value.trim()
  const keyword = custom || researchKeyword.value.trim() || Object.values(researchCatalog[currentNiche()?.name] || {}).flat()[0] || ''
  if (!keyword) return ''
  const terms = keyword.split(/\s+OR\s+/i).map((item) => item.trim()).filter(Boolean)
  const base = custom ? keyword : terms.length > 1 ? `(${terms.join(' OR ')})` : keyword
  return `${base}${/\blang:/i.test(base) ? '' : ' lang:in'}${/(?:^|\s)-is:retweet\b/i.test(base) ? '' : ' -is:retweet'}${/(?:^|\s)-is:reply\b/i.test(base) ? '' : ' -is:reply'}`
}
function openResearch(filter = 'top') {
  const q = query()
  if (!q) return
  window.open(`https://x.com/search?${new URLSearchParams({ q, src: 'typed_query', f: filter })}`, '_blank', 'noopener')
  form.value.source_query = q
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
  form.value.media = [...new Set(item.media || [])]
  threadPostCount.value = item.thread_post_count || 1
  message.value = `${threadPostCount.value > 1 ? `Thread ${threadPostCount.value} post` : 'Post'} ditangkap · ${form.value.media.length} media. Review sebelum disimpan.`
  error.value = ''
}
function removeMedia(index) { form.value.media.splice(index, 1) }
async function save() {
  if (!form.value.canonical_url.trim() || !form.value.original_text.trim()) return
  saving.value = true; error.value = ''; message.value = ''
  try {
    const saved = await apiRequest('/api/content-items', { method: 'POST', body: JSON.stringify(form.value) })
    message.value = 'Konten dan media berhasil disimpan ke bank konten.'
    if (saved?.id) router.push(`/content-bank/${saved.id}`)
  } catch (e) { error.value = e.message } finally { saving.value = false }
}
watch(researchNiche, () => { researchCategory.value = ''; researchKeyword.value = ''; researchQuery.value = '' })
onMounted(async () => { niches.value = await apiRequest('/api/content-niches'); productTypes.value = await apiRequest('/api/product-types').catch(() => []) ; window.addEventListener('message', receiveCapture) })
onBeforeUnmount(() => window.removeEventListener('message', receiveCapture))
</script>

<template>
  <div class="capture-page">
    <section class="hero-row"><div><p class="eyebrow">AFFILIATE WORKSPACE / CAPTURE</p><h1>Tangkap konten</h1><p class="muted">Riset di X, tangkap posting beserta gambar/video, review, lalu simpan ke bank konten.</p></div><RouterLink class="button-secondary" to="/content-bank">Lihat bank konten →</RouterLink></section>
    <section class="panel research-form"><div><h2>1. Riset konten X</h2><p class="muted">Pilih niche dan kategori untuk membuat query, lalu buka hasil populer di X.</p></div><div class="form-grid"><select v-model="researchNiche" class="select" :disabled="customMode()"><option value="">Pilih niche konten</option><option v-for="niche in niches" :key="niche.id" :value="niche.id">{{ niche.name }}</option></select><select v-model="researchCategory" class="select" :disabled="customMode() || !categories().length"><option value="">Pilih kategori konten</option><option v-for="category in categories()" :key="category" :value="category">{{ category }}</option></select></div><div v-if="keywords().length" class="query-options"><button v-for="keyword in keywords()" :key="keyword" class="query-chip" type="button" :disabled="customMode()" :class="{ selected: researchKeyword === keyword }" @click="researchKeyword = keyword; researchQuery = ''">{{ keyword }}</button></div><div class="custom-query-row"><input v-model="researchQuery" class="input" placeholder="Keyword/query custom (opsional — mengambil alih preset)" /><button v-if="customMode()" class="button-secondary" type="button" @click="researchQuery = ''">Pakai preset</button></div><p v-if="query()" class="query-preview">Query: <code>{{ query() }}</code></p><div class="research-actions"><button class="button-primary" :disabled="!query()" @click="openResearch('top')">Buka Populer ↗</button><button class="button-secondary" :disabled="!query()" @click="openResearch('live')">Buka Terbaru ↗</button><button class="button-secondary" :disabled="!query()" @click="openResearch('media')">Buka Media ↗</button></div></section>
    <section class="panel content-form"><div><h2>2. Review & simpan tangkapan</h2><p class="muted">Extension X Research mengisi form ini. Data belum masuk DB sampai tombol simpan ditekan.</p></div><input v-model="form.canonical_url" class="input" placeholder="URL post X" /><div class="form-grid"><input v-model="form.author_handle" class="input" placeholder="@author (opsional)" /><input v-model="form.source_query" class="input" placeholder="Query riset (opsional)" /></div><textarea v-model="form.original_text" class="textarea" rows="8" placeholder="Konten asli hasil tangkapan..."></textarea><div class="label-group"><strong>Niche konten</strong><div class="tag-options"><label v-for="niche in niches" :key="niche.id" class="tag-option"><input v-model="form.niche_ids" type="checkbox" :value="niche.id" /> {{ niche.name }}</label></div></div><div class="label-group"><strong>Jenis Barang (opsional)</strong><div class="tag-options"><label v-for="type in productTypes" :key="type.id" class="tag-option"><input v-model="form.product_type_ids" type="checkbox" :value="type.id" /> {{ type.name }}</label></div></div><div v-if="form.media.length" class="media-review"><strong>Media tertangkap ({{ form.media.length }})</strong><div class="media-grid"><figure v-for="(src, index) in form.media" :key="src"><video v-if="/\.(mp4|webm|mov)(?:\?|$)/i.test(src)" :src="src" controls></video><img v-else :src="src" alt="Media post X" /><button type="button" @click="removeMedia(index)">Hapus</button></figure></div></div><button class="button-primary" :disabled="saving || !form.canonical_url.trim() || !form.original_text.trim()" @click="save">{{ saving ? 'Menyimpan...' : 'Simpan ke bank konten' }}</button><p v-if="message" class="save-notice">✓ {{ message }}</p><p v-if="error" class="error-box">{{ error }}</p></section>
  </div>
</template>

<style scoped>
.capture-page{display:grid;gap:18px}.hero-row{display:flex;justify-content:space-between;align-items:end;gap:16px}.hero-row h1{font-size:48px;margin:0}.eyebrow{font:12px 'DM Mono';color:#78867c;letter-spacing:2px}.content-form,.research-form{display:grid;gap:12px}.form-grid{display:grid;grid-template-columns:1fr 1fr;gap:10px}.custom-query-row{display:flex;gap:8px}.custom-query-row .input{flex:1}.textarea{width:100%;resize:vertical;border:1px solid #cedbd0;border-radius:8px;padding:12px;background:#fffdf9;font:14px 'DM Mono';box-sizing:border-box}.label-group{display:grid;gap:8px;color:#52655a}.tag-options,.query-options,.research-actions{display:flex;flex-wrap:wrap;gap:8px}.tag-option,.query-chip{padding:8px 10px;border:1px solid #d9e2d8;border-radius:8px;background:#f8fbf7}.tag-option input{accent-color:#1f6b4f}.query-chip{cursor:pointer;color:#1f6b4f}.query-chip.selected{border-color:#1f6b4f;background:#e7f1e8}.query-chip:disabled{opacity:.55}.query-preview{margin:0;padding:10px 12px;border-radius:8px;background:#eef5ed;overflow:auto}.query-preview code{font-family:'DM Mono';color:#1f6b4f}.media-review{display:grid;gap:8px}.media-grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(150px,1fr));gap:10px}.media-grid figure{position:relative;margin:0;border:1px solid #d9e2d8;border-radius:8px;overflow:hidden;background:#eef5ed}.media-grid img,.media-grid video{display:block;width:100%;height:150px;object-fit:cover}.media-grid button{width:100%;padding:7px;border:0;background:#fff0ed;color:#a84f43;cursor:pointer}.save-notice{color:#176b4f;font-weight:600}.error-box{padding:10px;background:#fff0ed;color:#a84f43}@media(max-width:700px){.form-grid,.hero-row{display:grid;grid-template-columns:1fr}.custom-query-row{flex-direction:column}}
</style>
