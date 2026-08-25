(() => {
  const DEBUG = false
  const log = (...a) => DEBUG && console.log('[AffiliatorShopee]', ...a)

  function getCaptionFromIntent() {
    try {
      const url = new URL(location.href)
      const text = url.searchParams.get('text')
      if (text) return decodeURIComponent(text)
    } catch {}
    return ''
  }

  async function getCaptionFromStorage() {
    return new Promise((resolve) => {
      try {
        chrome.storage.local.get(['lastCaption'], (data) => resolve(data.lastCaption || ''))
      } catch { resolve('') }
    })
  }

  async function getCaptionFromClipboard() {
    try {
      const t = await navigator.clipboard.readText()
      return t || ''
    } catch { return '' }
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

  async function tryPaste() {
    let caption = getCaptionFromIntent()
    if (!caption) caption = await getCaptionFromStorage()
    if (!caption) caption = await getCaptionFromClipboard()
    if (!caption) { log('no caption'); return false }

    const composer = findComposer()
    if (!composer) { log('composer not found'); return false }

    // Jangan timpa jika user sudah mengetik
    const existing = (composer.innerText || composer.textContent || '').trim()
    if (existing.length > 8 && existing !== caption.trim().slice(0, existing.length)) {
      log('composer already filled, skip'); return false
    }

    insertText(composer, caption)
    notify(true)
    return true
  }

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
    if (e.data && e.data.type === 'AFFILIATOR_CAPTION') {
      chrome.storage.local.set({ lastCaption: e.data.caption })
      setTimeout(poll, 300)
    }
  })

  // Expose manual trigger via runtime message
  try {
    chrome.runtime.onMessage.addListener((msg) => {
      if (msg.type === 'AFFILIATOR_PASTE_NOW') poll()
    })
  } catch {}
})()
