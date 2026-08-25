<script setup>
import { onMounted, ref, watch } from 'vue'
import { useProductStore } from '@/stores/productStore'

const emit = defineEmits(['update:modelValue'])
const props = defineProps({ modelValue: String })
const products = useProductStore()
const models = ref([])
const selected = ref(props.modelValue || localStorage.getItem('ai_model') || 'opencode/muse-spark-1.2-contributor-free')

onMounted(async () => {
  try {
    models.value = await products.fetchModels()
    if (!models.value.find(m => m.id === selected.value)) {
      selected.value = models.value[0]?.id || selected.value
    }
  } catch { models.value = [{ id: 'opencode/muse-spark-1.2-contributor-free', name: 'Muse Spark 1.2 (Contributor Free)', free: true, note: 'Default' }] }
})
watch(selected, v => { localStorage.setItem('ai_model', v); emit('update:modelValue', v) }, { immediate: true })
</script>

<template>
  <label class="field" style="margin:0; min-width:260px">
    <span style="font:500 11px 'DM Mono'; color:#53665a; letter-spacing:.06em; text-transform:uppercase">Model AI</span>
    <select v-model="selected" class="select" style="margin-top:6px">
      <option v-for="m in models" :key="m.id" :value="m.id">{{ m.name }}{{ m.free ? ' — Gratis' : '' }} — {{ m.note }}</option>
    </select>
  </label>
</template>
