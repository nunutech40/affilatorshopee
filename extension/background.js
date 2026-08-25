// AffiliatorShopee Helper - background service worker
// Menyimpan caption terakhir dari web app agar content script bisa paste walau URL intent ter-encode berbeda
chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  if (message.type === 'AFFILIATOR_SET_CAPTION') {
    chrome.storage.local.set({ lastCaption: message.caption, updatedAt: Date.now() }, () => sendResponse({ ok: true }))
    return true
  }
  if (message.type === 'AFFILIATOR_GET_CAPTION') {
    chrome.storage.local.get(['lastCaption'], (data) => sendResponse({ caption: data.lastCaption || '' }))
    return true
  }
})

// Optional: klik icon untuk coba paste di tab aktif
chrome.action.onClicked.addListener(async (tab) => {
  if (!tab.id) return
  chrome.scripting.executeScript({ target: { tabId: tab.id }, files: ['content.js'] })
})
