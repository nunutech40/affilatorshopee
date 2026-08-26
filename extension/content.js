(() => {
  if (window.__AFFILIATOR_SHOPEE_CONTENT_LOADED__) return
  window.__AFFILIATOR_SHOPEE_CONTENT_LOADED__ = true

  const DEBUG = false
  const log = (...a) => DEBUG && console.log('[AffiliatorShopee]', ...a)
  let attachingMedia = false

  function getCaptionFromIntent() {
    try {
      const url = new URL(location.href)
      const text = url.searchParams.get('text')
      if (text) return decodeURIComponent(text)
    } catch {}
    return ''
  }

  function getMediaFromIntent() {
    try {
      const url = new URL(location.href)
      return url.searchParams.getAll('affiliator_media')
    } catch { return [] }
  }

  async function getCaptionFromStorage() {
    return new Promise((resolve) => {
      try {
      chrome.storage.local.get(['lastCaption', 'pending'], (data) => resolve(data.pending ? (data.lastCaption || '') : ''))
      } catch { resolve('') }
    })
  }

  async function getMediaFromStorage() {
    return new Promise((resolve) => {
      try { chrome.storage.local.get(['lastMedia'], (data) => resolve(data.lastMedia || [])) } catch { resolve([]) }
    })
  }

  async function attachMedia(urls) {
    const uniqueUrls = [...new Set(urls)]
    if (!uniqueUrls.length) return 0
    let input = document.querySelector('input[data-testid="fileInput"], input[type="file"][accept*="image"], input[type="file"]')
    if (!input) {
      const mediaButton = document.querySelector('button[aria-label*="photo" i], button[aria-label*="video" i], [data-testid="attachments"]')
      if (mediaButton) {
        mediaButton.click()
        await new Promise((resolve) => setTimeout(resolve, 300))
        input = document.querySelector('input[data-testid="fileInput"], input[type="file"][accept*="image"], input[type="file"]')
      }
    }
    if (!input) return 0
    const attached = new Set((input.getAttribute('data-affiliator-media-keys') || '').split('|').filter(Boolean))
    const transfer = new DataTransfer()
    let attachedCount = 0
    for (const url of uniqueUrls.slice(0, 4)) {
      if (attached.has(url)) continue
      try {
        const result = await new Promise((resolve) => chrome.runtime.sendMessage({ type: 'AFFILIATOR_FETCH_MEDIA', url }, resolve))
        if (!result?.ok) continue
        const sourceBlob = new Blob([new Uint8Array(result.data)], { type: result.type })
        const sourceName = (new URL(url).pathname.split('/').pop() || 'media').replace(/[^\w.-]/g, '_')
        const file = await prepareXFile(sourceBlob, sourceName)
        transfer.items.add(file)
        attached.add(url)
        attachedCount++
      } catch (error) { log('media fetch failed', error) }
    }
    if (transfer.files.length) {
      const target = findComposer() || document.querySelector('[role="dialog"]')
      if (target) {
        const eventInit = { bubbles: true, cancelable: true, composed: true, dataTransfer: transfer }
        target.dispatchEvent(new DragEvent('dragenter', eventInit))
        target.dispatchEvent(new DragEvent('dragover', eventInit))
        target.dispatchEvent(new DragEvent('drop', eventInit))
      }
      input.setAttribute('data-affiliator-media-keys', [...attached].join('|'))
    }
    return attachedCount || (attached.size ? -1 : 0)
  }

  async function prepareXFile(blob, name) {
    if (blob.type !== 'image/webp') return new File([blob], name, { type: blob.type || 'application/octet-stream' })
    const bitmap = await createImageBitmap(blob)
    const canvas = document.createElement('canvas')
    canvas.width = bitmap.width
    canvas.height = bitmap.height
    canvas.getContext('2d').drawImage(bitmap, 0, 0)
    bitmap.close()
    const jpeg = await new Promise((resolve, reject) => canvas.toBlob((value) => value ? resolve(value) : reject(new Error('gagal konversi WebP')), 'image/jpeg', 0.92))
    return new File([jpeg], name.replace(/\.webp$/i, '.jpg'), { type: 'image/jpeg' })
  }

  async function rememberIntentMedia() {
    const caption = getCaptionFromIntent()
    const intentMedia = getMediaFromIntent()
    try {
      const stored = await new Promise((resolve) => chrome.storage.local.get(['lastMedia'], resolve))
      const media = [...new Set([...(stored.lastMedia || []), ...intentMedia])]
      if (!caption && !media.length) return
      await chrome.storage.local.set({ lastCaption: caption, lastMedia: media, pending: true })
    } catch {}
  }

  async function attachMediaWithRetry(urls) {
    const uniqueUrls = [...new Set(urls)]
    if (!uniqueUrls.length) return 0
    if (attachingMedia) return -1
    attachingMedia = true
    try {
      for (let attempt = 0; attempt < 3; attempt++) {
        const attached = await attachMedia(uniqueUrls)
        if (attached !== 0) return attached
        await new Promise((resolve) => setTimeout(resolve, 700))
      }
      return 0
    } finally {
      attachingMedia = false
    }
  }

  function findComposer() {
    // X intent / home
    const selectors = [
      '[data-testid="tweetTextarea_0"]',
      'div[data-testid="tweetTextarea_0"] div[contenteditable="true"]',
      'div[aria-label][role="textbox"][contenteditable="true"]',
      'div.DraftEditor-root div[contenteditable="true"]',
      // Threads
      'div[contenteditable="true"][role="textbox"]',
      'div[contenteditable="true"]'
    ]
    for (const sel of selectors) {
      const el = document.querySelector(sel)
      if (el && el.offsetParent !== null) return el
    }
    return null
  }

  function insertText(el, text) {
    el.focus()
    // Untuk contenteditable div (X & Threads)
    const sel = window.getSelection()
    const range = document.createRange()
    range.selectNodeContents(el)
    range.collapse(false)
    sel.removeAllRanges()
    sel.addRange(range)
    // Coba execCommand dulu (paling kompatibel)
    const ok = document.execCommand('insertText', false, text)
    if (!ok) {
      // Fallback: textContent + input event
      el.textContent = text
    }
    el.dispatchEvent(new InputEvent('input', { bubbles: true }))
    el.dispatchEvent(new Event('change', { bubbles: true }))
    // X perlu trigger React
    el.dispatchEvent(new KeyboardEvent('keyup', { bubbles: true }))
  }

  function notify(ok) {
    const div = document.createElement('div')
    div.textContent = ok ? '✓ Caption AffiliatorShopee ter-paste' : '✗ Gagal auto-paste, silakan Cmd+V'
    div.style.cssText = 'position:fixed;top:12px;right:12px;z-index:999999;background:' + (ok ? '#1f6b4f' : '#a24c41') + ';color:#fff;padding:10px 14px;border-radius:8px;font:600 13px sans-serif;box-shadow:0 8px 20px rgba(0,0,0,.2)'
    document.body.appendChild(div)
    setTimeout(() => div.remove(), 3500)
  }

  function notifyMedia(count, total) {
    if (!count) return
    const div = document.createElement('div')
    div.textContent = `✓ ${Math.min(count, total)}/${total} media lokal terlampir`
    div.style.cssText = 'position:fixed;top:58px;right:12px;z-index:999999;background:#1f6b4f;color:#fff;padding:10px 14px;border-radius:8px;font:600 13px sans-serif;box-shadow:0 8px 20px rgba(0,0,0,.2)'
    document.body.appendChild(div)
    setTimeout(() => div.remove(), 3500)
  }

  function notifyMediaFailure() {
    const div = document.createElement('div')
    div.textContent = '✗ Media lokal gagal dilampirkan; preview Shopee bukan file upload'
    div.style.cssText = 'position:fixed;top:58px;right:12px;z-index:999999;background:#a24c41;color:#fff;padding:10px 14px;border-radius:8px;font:600 13px sans-serif;box-shadow:0 8px 20px rgba(0,0,0,.2)'
    document.body.appendChild(div)
    setTimeout(() => div.remove(), 5000)
  }

  async function tryPaste() {
    let caption = getCaptionFromIntent()
    if (!caption) caption = await getCaptionFromStorage()
    if (!caption) { log('no caption'); return false }

    const composer = findComposer()
    if (!composer) { log('composer not found'); return false }

    // Jangan timpa jika user sudah mengetik
    const existing = (composer.innerText || composer.textContent || '').trim()
    if (existing.length > 8 && existing !== caption.trim().slice(0, existing.length)) {
      log('composer already filled, skip'); return false
    }

    insertText(composer, caption)
    try { await chrome.storage.local.set({ pending: false }) } catch {}
    notify(true)
    return true
  }

  // Simpan media segera, sebelum X berpindah dari URL intent ke composer SPA.
  rememberIntentMedia()

  // Coba beberapa kali karena composer X render async
  let attempts = 0
  async function poll() {
    attempts++
    const done = await tryPaste()
    if (!done && attempts < 12) setTimeout(poll, 700)
  }

  // Observer untuk SPA navigation
  const observer = new MutationObserver(() => {
    if (location.href.includes('intent/tweet') || location.href.includes('compose')) poll()
  })

  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', poll)
  else poll()
  observer.observe(document.documentElement, { childList: true, subtree: true })

  // Listen postMessage dari web app (jika web app kirim caption langsung)
  window.addEventListener('message', (e) => {
    if (e.data && (e.data.type === 'AFFILIATOR_CAPTION' || e.data.type === 'AFFILIATOR_SET_CONTENT')) {
      chrome.storage.local.set({ lastCaption: e.data.caption || '', lastMedia: e.data.media || [], pending: true })
      setTimeout(poll, 300)
    }
  })

  // Expose manual trigger via runtime message
  try {
    chrome.runtime.onMessage.addListener((msg) => {
      if (msg.type === 'AFFILIATOR_PASTE_NOW') poll()
      if (msg.type === 'AFFILIATOR_ATTACH_MEDIA') {
        getMediaFromStorage().then(async (media) => {
          const attached = await attachMediaWithRetry(media)
          if (attached > 0) notifyMedia(attached, media.length)
          else if (attached === 0) notifyMediaFailure()
        })
      }
    })
  } catch {}
})()
