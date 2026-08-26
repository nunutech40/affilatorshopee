<script setup>
import ModelSelector from '@/components/ModelSelector.vue'
</script>

<template>
  <RouterLink to="/" class="back-link">← Kembali ke library</RouterLink>
  <section class="panel settings-panel">
    <h1>Settings</h1>
    <p class="muted">Pengaturan aplikasi tersimpan di browser ini dan dipakai saat produk raw disimpan maupun saat membuat varian caption.</p>
    <ModelSelector />
    <div class="setting-note"><b>Alur AI:</b> produk raw baru otomatis direformat sekali setelah disimpan. Jika gagal, produk tetap tersimpan sebagai raw dan bisa diulang dari detail. Reformat varian caption tetap terpisah.</div>
  </section>
</template>

<style scoped>.settings-panel{max-width:680px}.muted{color:#78867c;line-height:1.6}.setting-note{margin-top:22px;padding:14px;border-radius:8px;background:#eef4ee;color:#53665a;font-size:13px;line-height:1.6}</style>
