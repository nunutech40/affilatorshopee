// AffiliatorShopee Helper - background service worker
// Menyimpan caption terakhir dari web app agar content script bisa paste walau URL intent ter-encode berbeda
chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  if (message.type === 'AFFILIATOR_SET_CONTENT' || message.type === 'AFFILIATOR_SET_CAPTION') {
    chrome.storage.local.set({ lastCaption: message.caption, lastMedia: message.media || [], updatedAt: Date.now() }, () => sendResponse({ ok: true }))
    return true
  }
  if (message.type === 'AFFILIATOR_GET_CAPTION') {
    chrome.storage.local.get(['lastCaption'], (data) => sendResponse({ caption: data.lastCaption || '' }))
    return true
  }
  if (message.type === 'AFFILIATOR_FETCH_MEDIA') {
    fetch(message.url)
      .then(async (response) => {
        if (!response.ok) throw new Error(`media status ${response.status}`)
        const buffer = await response.arrayBuffer()
        if (buffer.byteLength > 25 * 1024 * 1024) throw new Error('media terlalu besar untuk extension')
        sendResponse({ ok: true, type: response.headers.get('content-type') || 'application/octet-stream', data: Array.from(new Uint8Array(buffer)) })
      })
      .catch((error) => sendResponse({ ok: false, error: error.message }))
    return true
  }
})

// Optional: klik icon untuk coba paste di tab aktif
chrome.action.onClicked.addListener(async (tab) => {
  if (!tab.id) return
  chrome.tabs.sendMessage(tab.id, { type: 'AFFILIATOR_ATTACH_MEDIA' })
})
