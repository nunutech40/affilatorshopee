<script setup>
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { computed } from 'vue'

const route = useRoute()
const pageTitle = computed(() => {
  if (route.path === '/') return 'Product library'
  if (route.path === '/products/new') return 'Tambah produk'
  if (route.path === '/post-logs') return 'Riwayat posting'
  if (route.path === '/sold') return 'Produk terjual'
  if (route.path === '/settings') return 'Settings'
  if (route.path === '/content-bank') return 'Bank konten'
  return 'Product detail'
})
</script>

<template>
  <div class="app-shell">
    <header class="topbar">
      <RouterLink to="/" class="brand"><span class="brand-mark">AS</span><span>Affiliator<span class="brand-soft">Shopee</span></span></RouterLink>
      <nav class="nav-links">
        <RouterLink to="/" :class="{ active: route.path === '/' }">Library</RouterLink>
        <RouterLink to="/sold" :class="{ active: route.path === '/sold' }">Terjual</RouterLink>
        <RouterLink to="/post-logs" :class="{ active: route.path === '/post-logs' }">Riwayat</RouterLink>
        <RouterLink to="/settings" :class="{ active: route.path === '/settings' }">Settings</RouterLink>
        <RouterLink to="/content-bank" :class="{ active: route.path === '/content-bank' }">Bank konten</RouterLink>
        <RouterLink to="/products/new" class="nav-add">+ Produk baru</RouterLink>
      </nav>
    </header>
    <main class="page-wrap">
      <div class="eyebrow">AFFILIATE WORKSPACE / {{ pageTitle }}</div>
      <RouterView />
    </main>
  </div>
</template>
