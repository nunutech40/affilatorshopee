<script setup>
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { computed } from 'vue'

const route = useRoute()
const pageTitle = computed(() => {
  if (route.path === '/') return 'Product library'
  if (route.path === '/products/new') return 'Tambah produk'
  if (route.path === '/post-logs') return 'Riwayat posting'
  return 'Product detail'
})
</script>

<template>
  <div class="app-shell">
    <header class="topbar">
      <RouterLink to="/" class="brand"><span class="brand-mark">AS</span><span>Affiliator<span class="brand-soft">Shopee</span></span></RouterLink>
      <nav class="nav-links">
        <RouterLink to="/" :class="{ active: route.path === '/' }">Library</RouterLink>
        <RouterLink to="/post-logs" :class="{ active: route.path === '/post-logs' }">Riwayat</RouterLink>
        <RouterLink to="/products/new" class="nav-add">+ Produk baru</RouterLink>
      </nav>
    </header>
    <main class="page-wrap">
      <div class="eyebrow">AFFILIATE WORKSPACE / {{ pageTitle }}</div>
      <RouterView />
    </main>
  </div>
</template>
