<script setup>
import { ref } from 'vue'
const props = defineProps({ caption: { type: String, required: true }, media: { type: Array, default: () => [] }, showCopy: { type: Boolean, default: true }, disabled: { type: Boolean, default: false } })
const copied = ref(false)

function normalizeCaption(value) {
  let text = String(value || '').replace(/(#[\p{L}\p{N}_]+)(?=#)/gu, '$1 ')
  const seen = new Set()
  text = text.replace(/(^|\s)#([\p{L}\p{N}_]+)/gu, (match, prefix, tag) => {
    const key = tag.toLocaleLowerCase()
    if (seen.has(key)) return ''
    seen.add(key)
    return `${prefix}#${tag}`
  })
  text = text.replace(/[ \t]{2,}/g, ' ').replace(/[ \t]+\n/g, '\n').trim()
  const seenLines = new Set()
  text = text.split('\n').filter((line) => {
    const lineKey = line.trim().toLocaleLowerCase()
    if (!lineKey || !/^(?:[💸⚡️👇🔥⭐️✅]|harga|https?:\/\/|#)/u.test(lineKey)) return true
    if (seenLines.has(lineKey)) return false
    seenLines.add(lineKey)
    return true
  }).join('\n').trim()
  if (text.length <= 280) return text

  const lines = text.split('\n').map((line) => line.trim()).filter(Boolean)
  const suffixStart = lines.findIndex((line, index) => index > 0 && (/https?:\/\//i.test(line) || /(^|\s)#/u.test(line)))
  if (suffixStart < 0) return text.slice(0, 280).trim()
  const suffix = lines.slice(suffixStart).join('\n')
  const prefix = lines.slice(0, suffixStart)
  while (prefix.length && `${prefix.join('\n')}\n${suffix}`.length > 280) prefix.pop()
  const result = prefix.length ? `${prefix.join('\n')}\n${suffix}` : suffix
  return result.slice(0, 280).trim()
}

async function share() {
  if (props.disabled) return
  const shareCaption = normalizeCaption(props.caption)
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
async function copy() { if (props.disabled) return; try { await navigator.clipboard.writeText(normalizeCaption(props.caption)); copied.value = true } catch { copied.value = false }; setTimeout(() => { copied.value = false }, 2000) }
</script>

<template><div class="share-actions"><button class="button-primary" :disabled="props.disabled" :title="props.disabled ? 'Isi promo text terlebih dahulu' : 'Share ke X'" @click="share">{{ copied ? 'Caption copied' : 'Share ke X' }}</button><button v-if="props.showCopy" class="button" :disabled="props.disabled" @click="copy">Copy caption</button></div></template>

<style scoped>.share-actions { display: flex; gap: 8px; flex-wrap: wrap; }</style>
