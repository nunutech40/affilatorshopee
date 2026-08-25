<script setup>
import { ref } from 'vue'

const emit = defineEmits(['save'])
const rawText = ref('')
const shopeeLink = ref('')
const imageURL = ref('')
const notes = ref('')

function submit() {
  emit('save', {
    raw_text: rawText.value,
    shopee_link: shopeeLink.value,
    image_url: imageURL.value || null,
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
    <div class="form-grid">
      <div class="field"><label>Link affiliate Shopee</label><input v-model="shopeeLink" class="input" placeholder="https://shopee.co.id/..." /></div>
      <div class="field"><label>URL gambar eksternal</label><input v-model="imageURL" class="input" placeholder="https://..." /></div>
    </div>
    <div class="field"><label>Catatan</label><input v-model="notes" class="input" placeholder="Opsional" /></div>
    <button class="button-primary" :disabled="!rawText.trim() || !shopeeLink.trim()" @click="submit">Simpan produk raw</button>
  </section>
</template>

<style scoped>.muted { color: #78867c; font-size: 13px; line-height: 1.5; margin: 8px 0 22px; }</style>
