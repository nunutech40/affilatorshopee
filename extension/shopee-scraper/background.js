chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  if (message.type === 'AFFILIATOR_SCRAPER_STORE') {
    chrome.storage.local.set({ lastShopeeProduct: message.product, updatedAt: Date.now() }, () => sendResponse({ ok: true }))
    return true
  }
  if (message.type === 'AFFILIATOR_CREATE_PRODUCT') {
    fetch('http://localhost:8080/api/products', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(message.product)
    })
      .then(async (response) => {
        const body = await response.json().catch(() => null)
        if (!response.ok || !body?.success) throw new Error(body?.error?.message || `API status ${response.status}`)
        sendResponse({ ok: true, data: body.data })
      })
      .catch((error) => sendResponse({ ok: false, error: error.message }))
    return true
  }
})
