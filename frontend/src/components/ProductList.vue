<script setup>
import ShareButton from './ShareButton.vue'
const props = defineProps({ items: { type: Array, default: () => [] }, selectedIds: { type: Array, default: () => [] } })
const emit = defineEmits(['delete', 'update-model', 'update:selectedIds'])
const toggle = (id) => emit('update:selectedIds', props.selectedIds.includes(id) ? props.selectedIds.filter((item) => item !== id) : [...props.selectedIds, id])
const toggleAll = (checked) => emit('update:selectedIds', checked ? props.items.map((item) => item.id) : [])
const money = (value) => value == null ? '' : new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(value)
const sourceLabel = (value) => ({ import_x: 'Import X', scrape_shopee: 'Scrape Shopee', raw_text: 'Raw text' }[value] || 'Raw text')
const savedDate = (value) => value ? new Intl.DateTimeFormat('id-ID', { day: 'numeric', month: 'long', year: 'numeric' }).format(new Date(value)) : '-'
</script>

<template>
  <div class="product-list">
    <div class="list-header">
      <span><input type="checkbox" :checked="props.items.length > 0 && props.items.every((item) => props.selectedIds.includes(item.id))" @change="toggleAll($event.target.checked)" /></span>
      <span>Produk</span>
      <span>Model</span>
      <span>Status</span>
      <span>Harga</span>
      <span>Post</span>
      <span></span>
    </div>
    <article v-for="product in items" :key="product.id" class="product-row">
      <input class="row-check" type="checkbox" :checked="props.selectedIds.includes(product.id)" @click.stop @change="toggle(product.id)" />
      <div class="row-main">
        <RouterLink :to="`/products/${product.id}`" class="row-title">{{ product.product_name || 'Raw product belum direformat' }}</RouterLink>
        <div class="row-sub">{{ sourceLabel(product.source_category) }} · {{ product.cluster || 'uncategorized' }} · tag: {{ product.tracking_tag || '-' }}</div>
        <div class="row-date">Disimpan {{ savedDate(product.created_at) }}</div>
        <div v-if="product.niches?.length" class="niche-list"><span v-for="niche in product.niches" :key="niche.id" class="niche-pill">{{ niche.name }}</span></div>
      </div>
      <select class="model-select" :value="product.content_model || ''" @change="emit('update-model', product.id, $event.target.value)"><option value="">Pilih angle</option><option value="trending">Trending</option><option value="branded">Branded</option><option value="cheap">Murah</option><option value="curated">Curated</option></select>
      <span class="status" :class="product.status">{{ product.status }}</span>
      <span class="row-price">{{ product.sale_price ? money(product.sale_price) : '-' }}</span>
      <span class="row-post">{{ product.post_count || 0 }} post · {{ product.click_count || 0 }} klik</span>
      <div class="row-actions"><ShareButton :caption="product.reformatted_text || ''" :disabled="!product.reformatted_text?.trim()" :show-copy="false" /><RouterLink :to="`/products/${product.id}`" class="row-link">Detail →</RouterLink><button class="icon-delete" title="Hapus produk" @click.stop="emit('delete', product.id)">🗑</button></div>
    </article>
  </div>
</template>

<style scoped>
.product-list{display:grid; gap:8px}
.list-header{display:none; grid-template-columns:24px 1fr 90px 110px 110px 60px 110px; gap:12px; padding:0 16px 6px; font:600 10px 'DM Mono'; letter-spacing:.08em; text-transform:uppercase; color:#8a978d}
.product-row{display:grid; grid-template-columns:24px 1fr 90px 110px 110px 60px 110px; gap:12px; align-items:center; padding:14px 16px; background:rgba(255,255,255,.62); border:1px solid #d9ded6; border-radius:10px; transition:background .15s}
.product-row:hover{background:#fff}
.row-main{min-width:0}.row-check{accent-color:#1f6b4f;justify-self:center}
.row-title{ display:block; font:600 14px 'Space Grotesk'; color:#1f2721; white-space:nowrap; overflow:hidden; text-overflow:ellipsis }
.row-sub{ font:11px 'DM Mono'; color:#8a978d; white-space:nowrap; overflow:hidden; text-overflow:ellipsis; margin-top:2px}.row-date{font:11px 'DM Mono';color:#718077;margin-top:4px}
.niche-list{display:flex;gap:5px;flex-wrap:wrap;margin-top:6px}.niche-pill{font:10px 'DM Mono';color:#1f6b4f;background:#e7f1e8;border-radius:999px;padding:3px 7px}
.badge.model{ justify-self:start; font:600 11px 'DM Mono'; text-transform:capitalize; color:#5a6b5e; background:#e7eee6; padding:5px 8px; border-radius:999px}
.model-select{width:100%; border:1px solid #d9ded6; border-radius:7px; padding:7px 5px; background:#f8fbf7; color:#52655a; font:600 11px 'DM Mono'}
.row-price{ font:600 13px 'Space Grotesk'; color:#1f6b4f}
.row-post{ font:12px 'DM Mono'; color:#6b7a6e; text-align:center}
.row-link{ font:700 13px sans-serif; color:#1f6b4f; white-space:nowrap}
.row-actions{ display:flex; align-items:center; gap:8px; justify-self:end}.row-actions :deep(.share-actions){flex-wrap:nowrap}.row-actions :deep(.button-primary){padding:7px 9px;font-size:11px;white-space:nowrap}.row-actions :deep(.button-primary:disabled){opacity:.45;cursor:not-allowed}
.icon-delete{ border:1px solid #e1c4be; background:#fff1ee; color:#a24c41; border-radius:7px; padding:6px 8px; font-size:13px; line-height:1}
.icon-delete:hover{ background:#f8ddd6}
@media(max-width:1100px){ .list-header{display:none} .product-row{grid-template-columns:24px 1fr auto} .model-select,.row-price,.row-post{ display:none} .row-actions :deep(.button-primary){font-size:0;padding:7px 8px}.row-actions :deep(.button-primary)::before{content:'↗';font-size:15px} }
@media(min-width:901px){ .list-header{display:grid} }
</style>
