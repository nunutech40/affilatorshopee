<script setup>
import { ref } from 'vue'
const props = defineProps({ caption: { type: String, required: true }, media: { type: Array, default: () => [] } })
const copied = ref(false)
async function share() {
  const shareCaption = props.caption
  try { await navigator.clipboard.writeText(shareCaption); copied.value = true } catch { copied.value = false }
  // Simpan ke extension storage jika helper terinstall (satu kesatuan dengan web app)
  try { if (window.chrome?.storage?.local) window.chrome.storage.local.set({ lastCaption: shareCaption, lastMedia: props.media }) } catch {}
  try { if (window.chrome?.runtime?.sendMessage) window.chrome.runtime.sendMessage({ type: 'AFFILIATOR_SET_CONTENT', caption: shareCaption, media: props.media }) } catch {}
  window.postMessage({ type: 'AFFILIATOR_SET_CONTENT', caption: shareCaption, media: props.media }, '*')
  const params = new URLSearchParams({ caption: shareCaption })
  props.media.forEach((item) => params.append('media', item))
  const url = `/api/share/x?${params.toString()}`
  window.open(url, '_blank', 'noopener,noreferrer')
  setTimeout(() => { copied.value = false }, 2000)
}
async function copy() { try { await navigator.clipboard.writeText(props.caption); copied.value = true } catch { copied.value = false }; setTimeout(() => { copied.value = false }, 2000) }
</script>

<template><div class="share-actions"><button class="button-primary" @click="share">{{ copied ? 'Caption copied' : 'Share ke X' }}</button><button class="button" @click="copy">Copy caption</button></div></template>

<style scoped>.share-actions { display: flex; gap: 8px; flex-wrap: wrap; }</style>
