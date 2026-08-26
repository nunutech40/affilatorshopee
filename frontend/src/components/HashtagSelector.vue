<script setup>
import { computed, ref } from 'vue'
const props = defineProps({ modelValue: { type: Array, default: () => [] }, suggestions: { type: Array, default: () => [] } })
const emit = defineEmits(['update:modelValue'])
const custom = ref('')
const tags = computed(() => props.modelValue)
function toggle(tag) { const next = tags.value.includes(tag) ? tags.value.filter((item) => item !== tag) : [...tags.value, tag]; if (next.length <= 3) emit('update:modelValue', next) }
function addCustom() { const tag = custom.value.trim(); if (!tag) return; toggle(tag.startsWith('#') ? tag : `#${tag}`); custom.value = '' }
</script>

<template>
  <div class="hashtag-picker">
    <div class="tag-row"><button v-for="tag in suggestions" :key="tag" class="tag" :class="{ selected: tags.includes(tag) }" @click="toggle(tag)">{{ tag }}</button></div>
    <div class="tag-input-row"><input v-model="custom" class="input tag-input" placeholder="#customhashtag" @keyup.enter="addCustom" /><button class="button" :disabled="tags.length >= 3" @click="addCustom">Tambah</button></div>
    <small>{{ tags.length }}/3 hashtag dipilih</small>
  </div>
</template>

<style scoped>.tag-row { display: flex; flex-wrap: wrap; gap: 7px; margin-bottom: 9px; }.tag { border: 1px solid #ccd5cc; border-radius: 99px; padding: 7px 10px; color: #607265; background: #fffdf8; font-size: 12px; }.tag.selected { color: #fff; border-color: #1f6b4f; background: #1f6b4f; }.tag-input-row { display: flex; gap: 7px; }.hashtag-picker small { display: block; color: #829086; margin-top: 8px; font-size: 11px; }</style>
