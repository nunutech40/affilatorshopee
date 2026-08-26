<script setup>
defineProps({ count: Number, loading: Boolean })
defineEmits(['run', 'run-variant'])
</script>

<template><div class="bulk-actions"><button class="button" :disabled="loading || !count || count > 10" @click="$emit('run')">{{ loading ? 'AI sedang bekerja...' : `Reformat ${count || 0} produk (max 10)` }}</button><button class="button" :disabled="loading || !count || count > 10" @click="$emit('run-variant')">{{ loading ? 'AI sedang bekerja...' : `Buat varian caption (${count || 0})` }}</button></div></template>

<style scoped>.bulk-actions{display:flex;gap:8px;flex-wrap:wrap}</style>
