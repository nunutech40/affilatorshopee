<script setup>
import { reactive, watch } from 'vue'

const props = defineProps({ product: { type: Object, required: true }, saving: Boolean })
const emit = defineEmits(['save'])
const fields = ['product_name', 'shopee_link', 'image_url', 'keyword', 'problem', 'cluster', 'content_model', 'capture_angle', 'benefit_1', 'benefit_2', 'benefit_3', 'urgency', 'caption_template', 'notes']
const form = reactive({})

function sync(value) {
  fields.forEach((field) => { form[field] = value[field] ?? '' })
}
watch(() => props.product, sync, { immediate: true })

function submit() {
  const payload = { ...form }
  fields.forEach((field) => { if (payload[field] === '') payload[field] = null })
  emit('save', payload)
}
</script>

<template>
  <section class="panel">
    <h2>Edit structured data</h2>
    <p class="muted">Raw text asli tetap tersimpan dan tidak diubah.</p>
    <div class="form-grid">
      <div class="field"><label>Product name</label><input v-model="form.product_name" class="input" /></div>
      <div class="field"><label>Cluster</label><input v-model="form.cluster" class="input" placeholder="contoh: rumah tangga" /></div>
      <div class="field"><label>Content model</label><select v-model="form.content_model" class="select"><option value="">Pilih model</option><option value="trending">Trending</option><option value="branded">Branded</option><option value="cheap">Murah</option><option value="curated">Curated</option></select></div>
      <div class="field"><label>Capture angle</label><select v-model="form.capture_angle" class="select"><option value="">Tidak ada</option><option value="search">Search</option><option value="reply">Reply</option><option value="trend">Trend</option><option value="problem">Problem</option></select></div>
      <div class="field"><label>Keyword</label><input v-model="form.keyword" class="input" /></div>
      <div class="field"><label>Problem</label><input v-model="form.problem" class="input" /></div>
      <div class="field"><label>Benefit 1</label><input v-model="form.benefit_1" class="input" /></div>
      <div class="field"><label>Benefit 2</label><input v-model="form.benefit_2" class="input" /></div>
      <div class="field"><label>Benefit 3</label><input v-model="form.benefit_3" class="input" /></div>
      <div class="field"><label>Urgency</label><input v-model="form.urgency" class="input" placeholder="Kosongkan jika tidak ada bukti" /></div>
    </div>
    <div class="field"><label>Shopee link</label><input v-model="form.shopee_link" class="input" /></div>
    <div class="field"><label>Image URL</label><input v-model="form.image_url" class="input" /></div>
    <div class="field"><label>Notes</label><textarea v-model="form.notes" class="textarea" /></div>
    <button class="button-primary" :disabled="saving" @click="submit">{{ saving ? 'Menyimpan...' : 'Simpan perubahan' }}</button>
  </section>
</template>

<style scoped>.muted { color: #78867c; font-size: 13px; line-height: 1.5; margin: 8px 0 22px; }</style>
