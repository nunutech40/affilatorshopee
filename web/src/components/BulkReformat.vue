<script setup>
defineProps({ count: Number, loading: Boolean })
defineEmits(['run'])
</script>

<template><button class="button" :disabled="loading || !count || count > 20" @click="$emit('run')">{{ loading ? 'AI sedang bekerja...' : `Reformat ${count || 0} produk` }}</button></template>
