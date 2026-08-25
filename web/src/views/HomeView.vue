<script setup>
import { computed, onMounted, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { useProductStore } from '@/stores/productStore'
import ProductList from '@/components/ProductList.vue'
import BulkReformat from '@/components/BulkReformat.vue'

const products = useProductStore()
const selected = computed(() => products.items.filter((item) => item.selected).map((item) => item.id))
function toggle(id) { const item = products.items.find((product) => product.id === id); if (item) item.selected = !item.selected }
async function reformat() { const ids = selected.value; if (!ids.length || ids.length > 20) return; await products.reformat(ids); await products.fetchProducts(); products.items.forEach((item) => { item.selected = false }) }
let timer
watch(() => [products.filters.status, products.filters.content_model, products.filters.cluster], () => { products.fetchProducts() })
function search() { clearTimeout(timer); timer = setTimeout(() => products.fetchProducts(), 250) }
onMounted(() => products.fetchProducts())
</script>

<template>
  <section class="hero"><div><h1>Turn messy product data into ready-to-post copy.</h1><p class="hero-copy">Satu workspace untuk menyimpan produk affiliate, merapikan detail yang berantakan, dan mengulang angle caption tanpa mulai dari nol.</p></div><div class="hero-note">Posting tetap manual di X. Session browser yang menentukan akun, bukan aplikasi ini.</div></section>
  <div class="toolbar"><input v-model="products.filters.search" class="input" placeholder="Cari produk, keyword, atau raw text..." @input="search" /><select v-model="products.filters.status" class="select"><option value="">Semua status</option><option value="raw">Raw</option><option value="reformatted">Reformatted</option><option value="ready">Ready</option></select><select v-model="products.filters.content_model" class="select"><option value="">Semua model</option><option value="capture">Captured</option><option value="cheap">Cheap</option><option value="trending">Trending</option></select><input v-model="products.filters.cluster" class="input" placeholder="Filter cluster" /><BulkReformat :count="selected.length" :loading="products.loading" @run="reformat" /></div>
  <div v-if="products.error" class="error-box">{{ products.error }}</div>
  <div v-else-if="products.loading && !products.items.length" class="loading">Memuat product library...</div>
  <ProductList v-else-if="products.items.length" :items="products.items" :selected="selected" @toggle="toggle" />
  <div v-else class="empty"><h3>Belum ada produk</h3><p>Mulai dari data Shopee yang masih berantakan.</p><RouterLink to="/products/new" class="button-primary">Tambah produk pertama</RouterLink></div>
</template>
