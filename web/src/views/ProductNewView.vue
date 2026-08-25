<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import ProductParser from '@/components/ProductParser.vue'
import { useProductStore } from '@/stores/productStore'

const router = useRouter()
const products = useProductStore()
const error = ref('')
async function save(payload) { error.value = ''; try { const product = await products.createProduct(payload); router.push(`/products/${product.id}`) } catch (e) { error.value = e.message } }
</script>

<template><RouterLink to="/" class="back-link">← Kembali ke library</RouterLink><div v-if="error" class="error-box">{{ error }}</div><ProductParser @save="save" /></template>
