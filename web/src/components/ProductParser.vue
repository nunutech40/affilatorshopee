<script setup>
import { ref } from 'vue'

const emit = defineEmits(['save'])
const rawText = ref('')
const shopeeLink = ref('')
const imageURLs = ref([''])
const videoURL = ref('')
const notes = ref('')

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
    notes: notes.value || null,
  })
}
</script>

<template>
  <section class="panel">
    <h2>Save raw product</h2>
    <p class="muted">Paste apa adanya. AI akan merapikan data setelah produk tersimpan.</p>
    <div class="field">
      <label>Data Shopee mentah</label>
      <textarea v-model="rawText" class="textarea" placeholder="Paste nama, harga, rating, jumlah terjual, dan data lain di sini..." />
    </div>
    <div class="form-grid"><div class="field"><label>Link affiliate Shopee</label><input v-model="shopeeLink" class="input" placeholder="https://shopee.co.id/..." /></div><div class="field"><label>Video URL (optional)</label><input v-model="videoURL" class="input" placeholder="https://.../video.mp4" /></div></div>
    <div class="field"><label>URL gambar eksternal</label><div v-for="(imageURL, index) in imageURLs" :key="index" class="url-row"><input v-model="imageURLs[index]" class="input" placeholder="https://.../image.jpg" /><button v-if="imageURLs.length > 1" class="button button-danger" type="button" @click="removeImageURL(index)">×</button></div><button class="button add-url" type="button" @click="addImageURL">+ Add image URL</button></div>
    <div class="field"><label>Catatan</label><input v-model="notes" class="input" placeholder="Opsional" /></div>
    <button class="button-primary" :disabled="!rawText.trim() || !shopeeLink.trim()" @click="submit">Simpan produk raw</button>
  </section>
</template>

<style scoped>.muted { color: #78867c; font-size: 13px; line-height: 1.5; margin: 8px 0 22px; }.url-row { display: flex; gap: 7px; margin-bottom: 8px; }.url-row .input { flex: 1; }.add-url { margin-top: 3px; color: #1f6b4f; }</style>
