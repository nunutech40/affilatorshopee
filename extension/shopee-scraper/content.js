(() => {
  const scraperVersion = '1.2.9'
  if (window.__AFFILIATOR_SHOPEE_SCRAPER_VERSION__ === scraperVersion) return
  window.__AFFILIATOR_SHOPEE_SCRAPER_VERSION__ = scraperVersion

  const clean = (value) => String(value || '').replace(/\s+/g, ' ').trim()
  const unique = (values) => [...new Set(values.map(clean).filter(Boolean))]
  let networkData = []
  let observedURL = location.href
  window.addEventListener('message', (event) => {
    if (event.source === window && event.data?.type === 'AFFILIATOR_SHOPEE_NETWORK_RESPONSE') networkData.push(event.data.data)
  })
  function currentProductJson() {
    const jsonScripts = [...document.querySelectorAll('script[type="application/ld+json"]')].flatMap((node) => {
      try { const parsed = JSON.parse(node.textContent); return Array.isArray(parsed) ? parsed : [parsed] } catch { return [] }
    })
    return jsonScripts.find((item) => item?.['@type'] === 'Product') || {}
  }
  function resetForNavigation() {
    if (location.href !== observedURL) {
      observedURL = location.href
      networkData = []
    }
  }
  setInterval(resetForNavigation, 500)
  const pageText = () => String(document.body?.innerText || '')
  const pageLines = () => pageText().split(/\r?\n/).map((line) => line.trim()).filter(Boolean)

  function imageCandidates() {
    const productJson = currentProductJson()
    const isImageURL = (value) => {
      const url = clean(value)
      if (!/^https?:\/\//i.test(url) || !/shopee|seaimg|susercontent|cdn/i.test(url)) return false
      return !/(?:icon|logo|avatar|profile|sprite|favicon|badge|emoji|flag)/i.test(url)
    }
    const domImages = [...document.images]
      .filter((image) => (image.naturalWidth || image.width) >= 240 && (image.naturalHeight || image.height) >= 240)
      .map((image) => image.currentSrc || image.src || image.dataset.src || image.getAttribute('data-src'))
      .filter(isImageURL)
    const structuredImages = [productJson.image, document.querySelector('meta[property="og:image"]')?.content]
      .flat(Infinity)
      .filter(isImageURL)
    const networkImages = networkData
      .flatMap((data) => findNetworkValues(data, ['image', 'images', 'image_url', 'imageUrl', 'cover']))
      .filter(isImageURL)
    // DOM images are the rendered product gallery. Only fall back to network
    // payloads when the gallery has not rendered enough images yet; otherwise
    // API payloads also contain icons, avatars, and recommendation thumbnails.
    const values = domImages.length > 0
      ? [...structuredImages, ...domImages]
      : [...structuredImages, ...networkImages]
    const seen = new Set()
    return values.filter((url) => {
      const canonical = String(url).replace(/[?#].*$/, '')
      if (seen.has(canonical)) return false
      seen.add(canonical)
      return true
    }).slice(0, 20)
  }

  function findNetworkValues(value, keys, depth = 0) {
    if (!value || depth > 5) return []
    if (Array.isArray(value)) return value.flatMap((item) => findNetworkValues(item, keys, depth + 1))
    if (typeof value !== 'object') return []
    const wanted = new Set(keys.map((key) => String(key).toLowerCase()))
    const found = []
    for (const [key, child] of Object.entries(value)) {
      if (wanted.has(String(key).toLowerCase()) && (typeof child === 'string' || typeof child === 'number' || Array.isArray(child))) found.push(...(Array.isArray(child) ? child : [child]))
      if (child && typeof child === 'object') found.push(...findNetworkValues(child, keys, depth + 1))
    }
    return found
  }

  function networkProduct() {
    const candidate = { title: '', description: '', images: [], price: '', rating: '', sold: '', brand: '', type: '' }
    for (const data of networkData) {
      const title = findNetworkValues(data, ['name', 'title', 'product_name', 'item_name']).find((value) => isUsefulTitle(value))
      const descriptions = findNetworkValues(data, ['description', 'desc', 'item_description', 'itemDescription']).filter((value) => clean(value).length > 20)
      const description = descriptions.sort((a, b) => String(b).length - String(a).length)[0] || ''
      if (title && !candidate.title) candidate.title = clean(title)
      if (String(description).length > candidate.description.length) candidate.description = clean(description)
      if (!candidate.price) candidate.price = findNetworkValues(data, ['price', 'price_min', 'priceMin']).find(Boolean) || ''
      if (!candidate.rating) candidate.rating = findNetworkValues(data, ['rating', 'rating_star', 'ratingStar', 'item_rating']).find(Boolean) || ''
      if (!candidate.sold) candidate.sold = findNetworkValues(data, ['sold', 'historical_sold', 'historicalSold']).find(Boolean) || ''
      if (!candidate.brand) candidate.brand = findNetworkValues(data, ['brand', 'brand_name', 'brandName', 'manufacturer']).find((value) => clean(value).length > 1) || ''
      if (!candidate.type) candidate.type = findNetworkValues(data, ['item_category', 'itemCategory', 'category_name', 'categoryName', 'product_type', 'productType']).find((value) => clean(value).length > 1) || ''
    }
    return candidate
  }

  function videoCandidate() {
    const values = [...document.querySelectorAll('video, video source')].map((node) => node.currentSrc || node.src || node.getAttribute('src'))
    return unique(values).find((url) => /^https?:\/\//i.test(url)) || null
  }

  function findText(pattern) { return clean(document.body.innerText.match(pattern)?.[0]) }

  function visibleProductPrice() {
    const pricePattern = /Rp\s*[\d]+(?:[.,]\d+)*(?:\s*[-–—]\s*Rp?\s*[\d]+(?:[.,]\d+)*)?/i
    const lines = pageLines()
    // Harga PDP selalu lebih dapat dipercaya daripada field API yang kadang
    // memakai unit mikro-rupiah. Scan seluruh teks agar tetap bekerja saat
    // judul atau panel harga dirender oleh layout Shopee yang berbeda.
    const match = lines.map((line) => line.match(pricePattern)).find(Boolean)
    return clean(match?.[0] || '')
  }

  function normalizeShopeePrice(value) {
    const text = clean(value)
    const match = text.match(/[\d]+(?:[.,]\d+)*/)
    if (!match) return ''
    const rawNumber = match[0]
    const numeric = Number(rawNumber.replace(/[.,]/g, ''))
    if (!Number.isFinite(numeric)) return text
    // Shopee's item API commonly returns IDR in 1/100000 rupiah units.
    if (numeric >= 10000000 && numeric % 100000 === 0) {
      const normalized = Math.round(numeric / 100000)
      // Guard against an already formatted/display value being scaled twice.
      if (normalized >= 1000 && normalized <= 100000000) return String(normalized)
    }
    return String(Math.round(numeric))
  }

  function isUsefulTitle(value) {
    const title = clean(value)
    return title.length > 5 && !/^shopee__?language$/i.test(title) && !/^beli .+ di shopee/i.test(title)
  }

  function visibleProductTitle() {
    const selectors = [
      'h1',
      '[data-testid="pdp-product-title"]',
      '[class*="product-title" i]',
      '[class*="productName" i]'
    ]
    for (const selector of selectors) {
      const value = [...document.querySelectorAll(selector)].map((node) => node.innerText || node.textContent).find(isUsefulTitle)
      if (value) return clean(value)
    }
    return pageLines().find((line) => isUsefulTitle(line) && line.length > 20 && !/^(spesifikasi|deskripsi produk|kategori|merek)$/i.test(line)) || ''
  }

  function firstNetworkValue(keys) {
    for (const data of networkData) {
      const value = findNetworkValues(data, keys).find((item) => clean(item).length > 1)
      if (value) return clean(value)
    }
    return ''
  }

  function conciseProductTitle(fullTitle, api) {
    const productJson = currentProductJson()
    const brand = clean(api.brand || productJson.brand?.name || firstNetworkValue(['brand', 'brand_name', 'brandName', 'manufacturer']))
    const type = clean(api.type || productJson.category || firstNetworkValue(['item_category', 'itemCategory', 'category_name', 'categoryName', 'product_type', 'productType']))
    const base = fullTitle.split(/\s+[|｜-]\s+/)[0] || type
    if (brand && base && !base.toLowerCase().startsWith(brand.toLowerCase())) return `${brand} ${base}`
    return base || brand || fullTitle
  }

  function detailSectionText(section) {
    const lines = pageLines()
    const headings = ['spesifikasi produk', 'detail produk', 'deskripsi produk']
    const start = lines.findIndex((line) => new RegExp(`^${section}\\b`, 'i').test(line))
    if (start < 0) return ''
    let end = lines.length
    for (let index = start + 1; index < lines.length; index++) {
      if (headings.some((heading) => new RegExp(`^${heading}\\b`, 'i').test(lines[index])) || /^(jadwal kirim|pengiriman|cara perawatan|note\s*:|produk terkait|ulasan|penilaian)$/i.test(lines[index])) {
        end = index
        break
      }
    }
    return lines.slice(start, end).filter((line) => !/^(laporkan|bagikan)$/i.test(line)).join('\n')
  }

  function visibleMetric(patterns) {
    const text = pageText()
    for (const pattern of patterns) {
      const match = text.match(pattern)
      if (match?.[1]) return clean(match[1])
    }
    return ''
  }

  function networkSpecificationText() {
    const output = []
    const add = (value) => {
      if (Array.isArray(value)) return value.forEach(add)
      if (!value || typeof value !== 'object') return
      const name = value.name || value.attribute_name || value.attributeName || value.label || value.key
      const detail = value.value || value.attribute_value || value.attributeValue || value.value_name
      if (name && detail && (typeof detail === 'string' || typeof detail === 'number')) {
        const line = `${clean(name)}: ${clean(detail)}`
        if (!output.includes(line)) output.push(line)
        return
      }
      Object.values(value).forEach((child) => {
        if (child && typeof child === 'object') add(child)
      })
    }
    for (const data of networkData) {
      findNetworkValues(data, ['attributes', 'item_attributes', 'itemAttributes', 'specifications', 'specs']).forEach(add)
      if (output.length >= 80) break
    }
    return output.slice(0, 80).join('\n')
  }

  function scrape() {
    resetForNavigation()
    const productJson = currentProductJson()
    const api = networkProduct()
    const fullTitle = clean(visibleProductTitle() || (isUsefulTitle(api.title) ? api.title : '') || productJson.name || document.querySelector('meta[property="og:title"]')?.content || document.title.replace(/\s*[|｜].*$/, ''))
    const title = conciseProductTitle(fullTitle, api)
    const description = clean(api.description || productJson.description || document.querySelector('meta[name="description"]')?.content)
    const specification = detailSectionText('spesifikasi produk')
    const descriptionBlock = detailSectionText('deskripsi produk') || detailSectionText('detail produk')
    const networkSpecs = networkSpecificationText()
    const visibleRating = visibleMetric([/\b([0-5](?:[.,]\d)?)\s+(?=⭐|★)/i, /\b([0-5](?:[.,]\d)?)\s+\d+[.,]?\d*\s*[KkMm]?\s+penilaian\b/i])
    const rating = visibleRating || productJson.aggregateRating?.ratingValue || api.rating
    const reviews = visibleMetric([/\b(\d+[.,]?\d*\s*[KkMm]?\+?)\s+penilaian\b/i]) || firstNetworkValue(['rating_count', 'ratingCount', 'review_count', 'reviewCount'])
    const visibleSold = visibleMetric([/\b(\d+[.,]?\d*\s*[KkMm]?\+?)\s+terjual\b/i, /\bterjual\s*(\d+[.,]?\d*\s*[KkMm]?\+?)/i])
    const sold = visibleSold ? `${visibleSold} terjual` : (api.sold ? `${api.sold} terjual` : '')
    const visiblePrice = visibleProductPrice()
    const price = visiblePrice || normalizeShopeePrice(api.price || productJson.offers?.price)
    const discount = findText(/\b\d{1,2}%\b/)
    const images = imageCandidates()
    const video = videoCandidate()
    const raw = [
      'RINGKASAN PRODUK',
      title,
      fullTitle !== title ? `Judul lengkap: ${fullTitle}` : '',
      rating || reviews ? `⭐️ Rating ${rating}${reviews ? ` · ${reviews} penilaian` : ''}` : '',
      sold ? `🔥 ${sold}` : '',
      price ? `💸 Harga ${clean(price)}` : '',
      discount ? `⚡️ Diskon ${discount}` : '',
      'SPESIFIKASI PRODUK',
      specification || (networkSpecs ? networkSpecs : ''),
      'DESKRIPSI PRODUK',
      descriptionBlock || description,
      'Media produk diambil dari halaman detail Shopee.'
    ].filter(Boolean).join('\n')
    return { source_category: 'scrape_shopee', source_url: location.href, shopee_link: '', raw_text: raw, image_urls: images, video_url: video, product_name: title || null }
  }

  chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
    if (message.type === 'AFFILIATOR_SCRAPE_PRODUCT' || message.type === 'AFFILIATOR_SCRAPE_PRODUCT_V2') {
      ;(async () => {
        const currentURL = new URL(location.href)
        const isShopeeHost = /(^|\.)shopee\.co\.id$/i.test(currentURL.hostname)
        const isProductPage = /\/product\/\d+\/\d+(?:\/|$)/i.test(currentURL.pathname) || /\/[^/]+-i\.\d+\.\d+(?:\/|$)/i.test(currentURL.pathname)
        if (!isShopeeHost || !isProductPage) {
          sendResponse({ ok: false, error: `Buka halaman detail produk Shopee terlebih dahulu. URL aktif: ${currentURL.href}` })
          return
        }
        // Shopee sering merender spesifikasi/deskripsi setelah section masuk viewport.
        window.scrollTo(0, document.documentElement.scrollHeight)
        await new Promise((resolve) => setTimeout(resolve, 2000))
        window.scrollTo(0, 0)
        await new Promise((resolve) => setTimeout(resolve, 300))
        const product = scrape()
        chrome.storage.local.set({ lastShopeeProduct: product }, () => sendResponse({ ok: true, product }))
      })().catch((error) => sendResponse({ ok: false, error: error.message || 'Gagal membaca halaman Shopee' }))
      return true
    }
    if (message.type === 'AFFILIATOR_IMPORT_SHOPEE_PRODUCT') {
      window.postMessage({ type: 'AFFILIATOR_IMPORT_SHOPEE_PRODUCT', product: message.product }, location.origin)
      sendResponse({ ok: true })
    }
  })
})()
