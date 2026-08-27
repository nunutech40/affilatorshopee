let scraped = null
const statusEl = document.getElementById('status')
const scrapeBtn = document.getElementById('scrape')
const sendBtn = document.getElementById('send')

function setStatus(message, type = '') { statusEl.textContent = message; statusEl.className = `status ${type}` }
function activeTab() { return chrome.tabs.query({ active: true, currentWindow: true }).then(([tab]) => tab) }
function productPageKey(url) {
  try {
    const parsed = new URL(url)
    if (!/(^|\.)shopee\.co\.id$/i.test(parsed.hostname)) return ''
    const path = parsed.pathname.replace(/\/+$/, '')
    const match = path.match(/\/product\/\d+\/\d+$/i) || path.match(/\/[^/]+-i\.\d+\.\d+$/i)
    return match ? `${parsed.hostname.toLowerCase()}${path.toLowerCase()}` : ''
  } catch { return '' }
}

Promise.all([
  activeTab(),
  new Promise((resolve) => chrome.storage.local.get(['lastShopeeProduct'], resolve))
]).then(([tab, data]) => {
  const saved = data.lastShopeeProduct
  if (!saved || !productPageKey(tab?.url) || productPageKey(saved.source_url) !== productPageKey(tab.url)) {
    scraped = null
    sendBtn.disabled = true
    setStatus('Klik “Ambil data halaman ini” untuk membaca produk ini.')
    return
  }
  scraped = saved
  sendBtn.disabled = false
  setStatus(`${scraped.product_name || 'Produk'} — ${scraped.image_urls?.length || 0} gambar, ${scraped.video_url ? '1 video' : 'tanpa video'} siap di-insert.`, 'ok')
})

async function sendToTab(tab, message) {
  try { return await chrome.tabs.sendMessage(tab.id, message) } catch (error) {
    await chrome.scripting.executeScript({ target: { tabId: tab.id }, files: ['content.js'] })
    return chrome.tabs.sendMessage(tab.id, message)
  }
}

scrapeBtn.addEventListener('click', async () => {
  scrapeBtn.disabled = true
  setStatus('Membaca halaman Shopee...')
  try {
    const tab = await activeTab()
    if (!productPageKey(tab?.url)) throw new Error(`Tab aktif bukan halaman detail produk Shopee: ${tab?.url || 'URL tidak tersedia'}`)
    const result = await sendToTab(tab, { type: 'AFFILIATOR_SCRAPE_PRODUCT_V2' })
    if (!result?.ok) throw new Error(result?.error || 'Halaman ini bukan detail produk Shopee')
    scraped = result.product
    await chrome.storage.local.set({ lastShopeeProduct: scraped })
    sendBtn.disabled = false
    setStatus(`${scraped.product_name || 'Produk'} — ${scraped.image_urls.length} gambar, ${scraped.video_url ? '1 video' : 'tanpa video'} siap dikirim.`, 'ok')
  } catch (error) { setStatus(error.message, 'error') } finally { scrapeBtn.disabled = false }
})

sendBtn.addEventListener('click', async () => {
  if (!scraped) return
  sendBtn.disabled = true
  try {
    const result = await new Promise((resolve) => chrome.runtime.sendMessage({
      type: 'AFFILIATOR_CREATE_PRODUCT',
      product: {
        raw_text: scraped.raw_text || '',
        product_name: scraped.product_name || null,
        // Halaman produk Shopee bukan link affiliate. Link affiliate diisi
        // kemudian dari halaman detail aplikasi.
        shopee_link: '',
        source_category: 'scrape_shopee',
        image_url: scraped.image_urls?.[0] || null,
        image_urls: scraped.image_urls || [],
        video_url: scraped.video_url || null,
        content_model: scraped.content_model || null,
        notes: `Diambil dari halaman detail Shopee melalui extension scraper. Sumber: ${scraped.source_url || 'halaman Shopee'}`
      }
    }, resolve))
    if (!result?.ok || !result.data?.product?.id) throw new Error(result?.error || 'Data gagal di-insert ke web app')
    const productID = result.data.product.id
    const tabs = await chrome.tabs.query({ url: ['http://localhost:8080/*', 'http://127.0.0.1:8080/*'] })
    const appTab = tabs.find((tab) => tab.id)
    const detailURL = `http://localhost:8080/products/${productID}`
    if (appTab) await chrome.tabs.update(appTab.id, { url: detailURL, active: true })
    else await chrome.tabs.create({ url: detailURL, active: true })
    scraped = null
    sendBtn.disabled = true
    await chrome.storage.local.remove(['lastShopeeProduct', 'updatedAt'])
    setStatus('Produk berhasil di-insert. Reformat AI bisa dijalankan dari halaman detail.', 'ok')
  } catch (error) { setStatus(error.message, 'error') } finally { sendBtn.disabled = false }
})
