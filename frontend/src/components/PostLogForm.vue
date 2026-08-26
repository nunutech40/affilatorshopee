<script setup>
import { ref } from 'vue'
import { apiRequest } from '@/stores/productStore'
const props = defineProps({ productId: { type: String, required: true }, caption: { type: String, required: true }, hashtags: { type: Array, default: () => [] } })
const emit = defineEmits(['saved'])
const notes = ref('')
const saving = ref(false)
const error = ref('')
async function save() { saving.value = true; error.value = ''; try { await apiRequest('/api/post-logs', { method: 'POST', body: JSON.stringify({ product_id: props.productId, platform: 'x', caption: props.caption, hashtags: props.hashtags, notes: notes.value || null }) }); notes.value = ''; emit('saved') } catch (e) { error.value = e.message } finally { saving.value = false } }
</script>

<template><section class="panel"><h2>Catat posting</h2><p class="muted">Klik setelah posting berhasil di X. Tidak ada akun yang disimpan.</p><div class="field"><label>Catatan opsional</label><input v-model="notes" class="input" placeholder="Contoh: repost angle problem" /></div><button class="button-primary" :disabled="saving" @click="save">{{ saving ? 'Menyimpan...' : 'Catat posting sekarang' }}</button><p v-if="error" class="inline-error">{{ error }}</p></section></template>

<style scoped>.muted { color: #78867c; font-size: 13px; line-height: 1.5; }.inline-error { color: #a24c41; font-size: 13px; }</style>
