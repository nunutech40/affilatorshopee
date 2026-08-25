document.getElementById('pasteBtn').addEventListener('click', async () => {
  const [tab] = await chrome.tabs.query({ active: true, currentWindow: true })
  if (!tab?.id) return
  chrome.tabs.sendMessage(tab.id, { type: 'AFFILIATOR_PASTE_NOW' }, () => {
    // fallback: inject content script if not yet injected
    if (chrome.runtime.lastError) {
      chrome.scripting.executeScript({ target: { tabId: tab.id }, files: ['content.js'] })
    }
  })
  window.close()
})
