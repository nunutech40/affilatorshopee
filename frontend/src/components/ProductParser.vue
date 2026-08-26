<script setup>
import { ref, computed } from 'vue'
import { apiRequest } from '@/stores/productStore'

const props = defineProps({ saving: Boolean })
const emit = defineEmits(['save', 'imported'])
const rawText = ref('')
const shopeeLink = ref('')
const imageURLs = ref([''])
const videoURL = ref('')
const contentModel = ref('')
const notes = ref('')
const xUrl = ref('')
const importing = ref(false)
const importError = ref('')

const isXMode = computed(() => xUrl.value.trim() !== '')

function addImageURL() { imageURLs.value.push('') }
function removeImageURL(index) { if (imageURLs.value.length > 1) imageURLs.value.splice(index, 1) }
function submit() {
  const images = imageURLs.value.map((value) => value.trim()).filter(Boolean)
  emit('save', {
    raw_text: rawText.value,
    shopee_link: shopeeLink.value,
    image_url: images[0] || null,
    image_urls: images,
    video_url: videoURL.value.trim() || null,
    source_category: 'raw_text',
    content_model: contentModel.value || null,
    notes: notes.value || null,
  })
}
async function importFromX() {
  importError.value = ''
  if (!xUrl.value.trim()) { importError.value = 'Link X wajib diisi'; return }
  importing.value = true
  try {
    const data = await apiRequest('/api/products/import/x', {
      method: 'POST',
      body: JSON.stringify({ x_url: xUrl.value.trim(), content_model: contentModel.value || null })
    })
    emit('imported', data)
  } catch (e) { importError.value = e.message } finally { importing.value = false }
}
</script>

<template>
  <section class="panel">
    <h2>Input data produk</h2>
    <p class="muted">Paste manual Shopee atau import langsung dari link X. Kalau link X diisi, input lain otomatis inactive kecuali pilihan content model.</p>
    <div class="field" style="background:#eef4ee; padding:14px; border-radius:8px; border:1px solid #d6e2d6">
      <label>Link X (opsional) — https://x.com/user/status/ID</label>
      <input v-model="xUrl" class="input" placeholder="https://x.com/DerryNassu/status/2090691934100902338" />
      <p class="muted" style="margin:6px 0 0">Jika diisi, sistem akan scrap caption + download image/video dari postingan tersebut. Status langsung <b>ready</b> (tidak perlu raw).</p>
      <div v-if="isXMode" style="margin-top:10px; display:flex; gap:8px; align-items:center">
        <button class="button-primary" :disabled="importing || !xUrl.trim()" @click="importFromX">{{ importing ? 'Meng-import...' : 'Import dari X' }}</button>
        <span v-if="importError" class="error-text">{{ importError }}</span>
        <span v-else class="muted">Input lain di bawah akan inactive</span>
      </div>
    </div>
    <div class="field">
      <label>Data Shopee mentah</label>
      <textarea v-model="rawText" :disabled="isXMode" class="textarea" :class="{ disabled: isXMode }" placeholder="Paste nama, harga, rating, jumlah terjual, dan data lain di sini..." />
    </div>
    <div class="form-grid"><div class="field"><label>Link affiliate Shopee</label><input v-model="shopeeLink" :disabled="isXMode" class="input" :class="{ disabled: isXMode }" placeholder="https://shopee.co.id/..." /></div><div class="field"><label>Video URL (optional)</label><input v-model="videoURL" :disabled="isXMode" class="input" :class="{ disabled: isXMode }" placeholder="https://.../video.mp4" /></div></div>
    <div class="field"><label>URL gambar eksternal</label><div v-for="(imageURL, index) in imageURLs" :key="index" class="url-row"><input v-model="imageURLs[index]" :disabled="isXMode" class="input" :class="{ disabled: isXMode }" placeholder="https://.../image.jpg" /><button v-if="imageURLs.length > 1" class="button button-danger" type="button" :disabled="isXMode" @click="removeImageURL(index)">×</button></div><button class="button add-url" type="button" :disabled="isXMode" @click="addImageURL">+ Add image URL</button></div>
    <div class="field"><label>Content model awal</label><select v-model="contentModel" class="select"><option value="">Belum ditentukan (bisa diisi AI)</option><option value="trending">Trending</option><option value="branded">Branded</option><option value="cheap">Murah</option><option value="capture">Captured (legacy)</option></select></div>
    <div class="field"><label>Catatan</label><input v-model="notes" :disabled="isXMode" class="input" :class="{ disabled: isXMode }" placeholder="Opsional" /></div>
    <button v-if="!isXMode" class="button-primary" :disabled="props.saving || !rawText.trim() || !shopeeLink.trim()" @click="submit">{{ props.saving ? 'Menyimpan + Reformat AI...' : 'Simpan + Reformat AI' }}</button>
    <p v-else class="muted">Link X terisi — gunakan tombol <b>Import dari X</b> di atas.</p>
  </section>
</template>

<style scoped>.muted { color: #78867c; font-size: 13px; line-height: 1.5; margin: 8px 0 22px; }.url-row { display: flex; gap: 7px; margin-bottom: 8px; }.url-row .input { flex: 1; }.add-url { margin-top: 3px; color: #1f6b4f; } .disabled{ opacity:.45; pointer-events:none; background:#f0f4f0 !important} .error-text{ color:#a24c41; font-size:13px}</style>
