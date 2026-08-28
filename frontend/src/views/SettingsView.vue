<script setup>
import { onMounted, ref } from 'vue'
import ModelSelector from '@/components/ModelSelector.vue'
import { useProductStore } from '@/stores/productStore'

const products = useProductStore()
const niches = ref([])
const newNiche = ref('')
const nicheMessage = ref('')
const nicheError = ref('')
const nicheSaving = ref(false)

async function loadNiches() { niches.value = await products.fetchNiches() }
async function addNiche() {
  if (!newNiche.value.trim()) return
  nicheSaving.value = true; nicheError.value = ''; nicheMessage.value = ''
  try { const niche = await products.createNiche(newNiche.value); niches.value = [...niches.value, niche].sort((a, b) => a.name.localeCompare(b.name)); newNiche.value = ''; nicheMessage.value = 'Jenis barang berhasil ditambahkan.' } catch (e) { nicheError.value = e.message } finally { nicheSaving.value = false }
}
async function removeNiche(niche) {
  if (!confirm(`Hapus jenis barang "${niche.name}"? Relasi dari produk juga akan dihapus.`)) return
  try { await products.deleteNiche(niche.id); niches.value = niches.value.filter((item) => item.id !== niche.id) } catch (e) { nicheError.value = e.message }
}
onMounted(loadNiches)
</script>

<template>
  <RouterLink to="/" class="back-link">← Kembali ke library</RouterLink>
  <section class="panel settings-panel">
    <h1>Settings</h1>
    <p class="muted">Pengaturan aplikasi tersimpan di browser ini dan dipakai saat produk raw disimpan maupun saat membuat varian caption.</p>
    <ModelSelector />
    <div class="setting-note"><b>Alur AI:</b> produk raw baru otomatis direformat sekali setelah disimpan. Jika gagal, produk tetap tersimpan sebagai raw dan bisa diulang dari detail. Reformat varian caption tetap terpisah.</div>
  </section>
  <section class="panel niche-panel">
    <h2>Master jenis barang</h2>
    <p class="muted">Jenis barang bisa dipasang lebih dari satu pada produk, lalu dipakai untuk filter di dashboard.</p>
    <div class="niche-add"><input v-model="newNiche" class="input" placeholder="Contoh: Kecantikan" @keyup.enter="addNiche" /><button class="button-primary" :disabled="nicheSaving || !newNiche.trim()" @click="addNiche">{{ nicheSaving ? 'Menambah...' : 'Tambah jenis barang' }}</button></div>
    <p v-if="nicheMessage" class="save-notice">✓ {{ nicheMessage }}</p><p v-if="nicheError" class="error-box">{{ nicheError }}</p>
    <div class="niche-master-list"><div v-for="niche in niches" :key="niche.id" class="niche-master-item"><span>{{ niche.name }}</span><button class="button button-danger" @click="removeNiche(niche)">Hapus</button></div><p v-if="!niches.length" class="muted">Belum ada master jenis barang.</p></div>
  </section>
</template>

<style scoped>.settings-panel{max-width:680px}.niche-panel{max-width:680px;margin-top:18px}.muted{color:#78867c;line-height:1.6}.setting-note{margin-top:22px;padding:14px;border-radius:8px;background:#eef4ee;color:#53665a;font-size:13px;line-height:1.6}.niche-add{display:flex;gap:8px;margin-top:16px}.niche-add .input{flex:1}.niche-master-list{display:grid;gap:8px;margin-top:16px}.niche-master-item{display:flex;align-items:center;justify-content:space-between;padding:10px 12px;border:1px solid #d9e2d8;border-radius:8px;background:#f8fbf7;color:#52655a}.niche-master-item .button{padding:5px 9px;font-size:11px}.save-notice{margin-top:12px;padding:10px 12px;border-radius:8px;background:#e7f3eb;color:#176b4f;font-weight:600}.error-box{margin-top:12px}@media(max-width:600px){.niche-add{flex-direction:column}}
</style>
