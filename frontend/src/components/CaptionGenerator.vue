<script setup>
import { computed, ref } from 'vue'
import { useCaptionStore } from '@/stores/captionStore'
import HashtagSelector from './HashtagSelector.vue'
import ShareButton from './ShareButton.vue'

const props = defineProps({ product: { type: Object, required: true } })
const captions = useCaptionStore()
const template = ref(props.product.caption_template || 'direct_product')
const hashtags = ref((props.product.hashtag_pool || []).slice(0, 3))
const generated = computed(() => captions.current)
const suggestions = computed(() => props.product.hashtag_pool || [])

async function generate() { await captions.generate(props.product.id, template.value, hashtags.value) }
async function variations() { await captions.generateVariations(props.product.id, template.value, hashtags.value) }
</script>

<template>
  <section class="panel">
    <div class="section-heading"><div><h2>Caption studio</h2><p class="muted">Caption yang dibagikan sudah termasuk hashtag.</p></div><span v-if="generated" class="counter" :class="{ warn: generated.over_limit }">{{ generated.character_count }}/280</span></div>
    <div class="caption-toolbar"><select v-model="template" class="select"><option value="direct_product">Direct product</option><option value="keyword_recommendation">Keyword + recommendation</option><option value="problem_specific">Problem-specific</option><option value="cheap_value">Cheap / value</option></select><button class="button-primary" :disabled="captions.loading" @click="generate">Generate</button><button class="button" :disabled="captions.loading" @click="variations">Buat 3 variasi</button></div>
    <HashtagSelector v-model="hashtags" :suggestions="suggestions" />
    <p v-if="captions.error" class="inline-error">{{ captions.error }}</p>
    <div v-if="generated" class="caption-result"><div class="caption-box">{{ generated.caption }}</div><div class="caption-toolbar"><ShareButton :caption="generated.caption" /></div></div>
    <div v-if="captions.variations.length" class="variation-list"><article v-for="(variation, index) in captions.variations" :key="variation.id || index" class="variation"><div class="variation-head"><span>{{ variation.label || `Variation ${index + 1}` }}</span><span>{{ variation.character_count || variation.caption?.length || 0 }} chars</span></div><div class="variation-text">{{ variation.caption || variation }}</div></article></div>
  </section>
</template>

<style scoped>.section-heading { display: flex; justify-content: space-between; gap: 12px; align-items: start; }.muted { color: #78867c; font-size: 13px; margin: 7px 0 18px; }.inline-error { color: #a24c41; font-size: 13px; }.caption-result { margin-top: 18px; }</style>
