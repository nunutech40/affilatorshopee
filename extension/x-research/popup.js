const statusEl = document.getElementById('status')
const button = document.getElementById('capture')

async function activeTab() { return (await chrome.tabs.query({ active: true, currentWindow: true }))[0] }
async function waitForTab(tabId) {
  const tab = await chrome.tabs.get(tabId)
  if (tab.status === 'complete') return
  await new Promise((resolve) => {
    const listener = (updatedId, info) => {
      if (updatedId === tabId && info.status === 'complete') { chrome.tabs.onUpdated.removeListener(listener); resolve() }
    }
    chrome.tabs.onUpdated.addListener(listener)
  })
}
async function capture() {
  button.disabled = true; statusEl.textContent = 'Mengambil post...'
  try {
    const tab = await activeTab()
    let result
    try { result = await chrome.tabs.sendMessage(tab.id, { type: 'AFFILIATOR_CAPTURE_X' }) } catch {
      await chrome.scripting.executeScript({ target: { tabId: tab.id }, files: ['content.js'] })
      result = await chrome.tabs.sendMessage(tab.id, { type: 'AFFILIATOR_CAPTURE_X' })
    }
    if (!result?.ok) throw new Error(result?.error || 'Post tidak ditemukan')
    const appTabs = await chrome.tabs.query({ url: ['http://localhost:8080/*', 'http://127.0.0.1:8080/*'] })
    const app = appTabs[0]
    const target = app?.url?.includes('/content-bank/capture') ? app : await chrome.tabs.create({ url: 'http://localhost:8080/content-bank/capture', active: true })
    await waitForTab(target.id)
    await chrome.scripting.executeScript({ target: { tabId: target.id }, func: (item) => window.postMessage({ type: 'AFFILIATOR_X_RESEARCH_CAPTURE', item }, '*'), args: [result.item] })
    statusEl.textContent = 'Post ditangkap. Review lalu simpan di Bank konten.'
  } catch (error) { statusEl.textContent = `Gagal: ${error.message}` } finally { button.disabled = false }
}
button.addEventListener('click', capture)
