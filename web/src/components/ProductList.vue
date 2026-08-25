<script setup>
defineProps({ items: { type: Array, default: () => [] }, selected: { type: Array, default: () => [] } })
const emit = defineEmits(['toggle'])
const money = (value) => value == null ? '' : new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(value)
</script>

<template>
  <div class="product-grid">
    <article v-for="product in items" :key="product.id" class="product-card">
      <div>
        <div class="card-top"><label class="check"><input type="checkbox" :checked="selected.includes(product.id)" @change="emit('toggle', product.id)" /><span></span></label><span class="status" :class="product.status">{{ product.status }}</span></div>
        <RouterLink :to="`/products/${product.id}`" class="card-title">{{ product.product_name || 'Raw product belum direformat' }}</RouterLink>
        <div class="cluster">{{ product.cluster || 'uncategorized' }}</div>
      </div>
      <div><div class="card-meta"><span v-if="product.sale_price">{{ money(product.sale_price) }}</span><span>{{ product.post_count || 0 }} post</span></div><RouterLink :to="`/products/${product.id}`" class="card-link">Buka detail →</RouterLink></div>
    </article>
  </div>
</template>

<style scoped>.check { display: inline-flex; margin-bottom: 15px; cursor: pointer; }.check input { accent-color: #1f6b4f; }</style>
