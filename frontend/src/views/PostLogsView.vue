<script setup>
import { onMounted, ref } from 'vue'
import { apiRequest } from '@/stores/productStore'
const logs = ref([])
const loading = ref(true)
const error = ref('')
onMounted(async () => { try { const data = await apiRequest('/api/post-logs'); logs.value = data.items } catch (e) { error.value = e.message } finally { loading.value = false } })
</script>

<template><section class="hero"><div><h1>Posting journal.</h1><p class="hero-copy">Catatan sederhana untuk melihat copy yang sudah dipakai. Tidak ada akun X yang disimpan.</p></div></section><div v-if="loading" class="loading">Memuat riwayat...</div><div v-else-if="error" class="error-box">{{ error }}</div><div v-else-if="!logs.length" class="empty">Belum ada riwayat posting.</div><div v-else class="log-list"><article v-for="log in logs" :key="log.id" class="log-item"><div class="log-head"><RouterLink :to="`/products/${log.product_id}`" class="card-link">Buka produk →</RouterLink><span>{{ new Date(log.posted_at).toLocaleString('id-ID') }}</span></div><div class="log-text">{{ log.caption }}</div><div class="cluster">{{ (log.hashtags || []).join(' ') }}</div></article></div></template>
