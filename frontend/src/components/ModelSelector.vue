<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useProductStore } from '@/stores/productStore'

const emit = defineEmits(['update:modelValue'])
const props = defineProps({ modelValue: String })
const products = useProductStore()
const models = ref([])
const searchQuery = ref('')
const selected = ref(props.modelValue || localStorage.getItem('ai_model') || 'stealth/ox-alpha')
const filteredModels = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  if (!query) return models.value
  return models.value.filter(model => `${model.name} ${model.id} ${model.provider}`.toLowerCase().includes(query))
})
const providers = ['openrouter', '9router', 'opencode', 'codex']
const providerLabel = (provider) => provider === '9router' ? '9router' : provider === 'openrouter' ? 'OpenRouter' : provider === 'opencode' ? 'OpenCode' : 'Codex CLI (lokal)'

onMounted(async () => {
  try {
    models.value = await products.fetchModels()
    if (!models.value.find(m => m.id === selected.value)) {
      selected.value = models.value[0]?.id || selected.value
    }
  } catch { models.value = [{ id: 'stealth/ox-alpha', name: 'Ox Alpha (OpenRouter)', free: false, note: 'OpenRouter' }] }
})
watch(selected, v => { localStorage.setItem('ai_model', v); emit('update:modelValue', v) }, { immediate: true })
watch(() => props.modelValue, v => {
  if (v && v !== selected.value) selected.value = v
})
</script>

<template>
  <label class="field" style="margin:0; min-width:260px">
    <span style="font:500 11px 'DM Mono'; color:#53665a; letter-spacing:.06em; text-transform:uppercase">Model AI</span>
    <input v-model="searchQuery" class="input" style="margin-top:6px" placeholder="Cari model..." />
    <select v-model="selected" class="select" style="margin-top:6px">
      <template v-for="provider in providers" :key="provider">
        <optgroup v-if="filteredModels.some(item => item.provider === provider)" :label="providerLabel(provider)">
          <option v-for="m in filteredModels.filter(item => item.provider === provider)" :key="m.id" :value="m.id">{{ m.name }}{{ m.free ? ' — Gratis' : '' }}</option>
        </optgroup>
      </template>
    </select>
    <small v-if="searchQuery && !filteredModels.length" style="display:block; margin-top:6px; color:#a24c41">Model tidak ditemukan.</small>
  </label>
</template>
