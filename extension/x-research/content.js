(() => {
  if (window.__AFFILIATOR_X_RESEARCH__) return
  window.__AFFILIATOR_X_RESEARCH__ = true

  function parseMetricValue(value) {
    const normalized = String(value || '').replace(/\s/g, '').replace(',', '.')
    const match = normalized.match(/([\d.]+)([KkMmBb])?/)
    if (!match) return 0
    const number = Number(match[1])
    if (!Number.isFinite(number)) return 0
    const multiplier = { k: 1e3, m: 1e6, b: 1e9 }[String(match[2] || '').toLowerCase()] || 1
    return Math.round(number * multiplier)
  }

  function metric(article, label) {
    const button = [...article.querySelectorAll('button')].find((item) => new RegExp(label, 'i').test(item.getAttribute('aria-label') || ''))
    const match = (button?.getAttribute('aria-label') || '').match(/[\d.,]+[KkMm]?/)
    return match ? parseMetricValue(match[0]) : 0
  }

  function extractArticle(article) {
    const link = [...article.querySelectorAll('a[href*="/status/"]')].find((item) => /\/status\/\d+/.test(item.getAttribute('href') || ''))
    const href = link ? new URL(link.getAttribute('href'), location.origin).href : location.href
    // Jangan fallback ke article.innerText: pada halaman detail X, innerText
    // dapat mencakup reply/nested reply yang dirender di dalam article.
    const text = article.querySelector('[data-testid="tweetText"]')?.innerText?.trim() || ''
    const authorLink = article.querySelector('[data-testid="User-Name"] a[href^="/"]') ||
      [...article.querySelectorAll('a[href^="/"]')].find((item) => /^\/[A-Za-z0-9_]{1,15}$/.test(item.getAttribute('href') || ''))
    const author = authorLink?.getAttribute('href') || ''
    const time = article.querySelector('time')?.dateTime || ''
    const socialContext = article.querySelector('[data-testid="socialContext"]')
    const socialContextText = socialContext?.innerText || ''
    const replyingToHandles = [...(socialContext?.querySelectorAll('a[href^="/"]') || [])]
      .map((item) => item.getAttribute('href') || '')
      .map((value) => value.replace(/^\//, '').split('/')[0])
      .filter((value) => /^[A-Za-z0-9_]{1,15}$/.test(value))
    const media = [...article.querySelectorAll('img, video')]
      .map((item) => item.currentSrc || item.src || item.querySelector('source')?.src)
      .filter((src) => src && !/profile_images|emoji|abs-0\.twimg\.com\/emoji/i.test(src))
    return {
      platform: 'x', canonical_url: href, external_post_id: href.match(/\/status\/(\d+)/)?.[1] || '',
      author_handle: author.replace(/^\//, ''), replying_to_handles: [...new Set(replyingToHandles)], original_text: text, media: [...new Set(media)],
      is_reply: /replying to|membalas/i.test(socialContextText),
      published_at: time, stats: { like_count: metric(article, 'like'), repost_count: metric(article, 'repost|retweet'), reply_count: metric(article, 'repl'), view_count: metric(article, 'view') },
    }
  }

  function normalizeHandle(value) {
    return String(value || '').replace(/^@/, '').toLowerCase()
  }

  function isThreadContinuation(item, rootAuthor) {
    if (normalizeHandle(item.author_handle) !== normalizeHandle(rootAuthor)) return false
    if ((item.replying_to_handles || []).length) {
      return (item.replying_to_handles || []).some((handle) => normalizeHandle(handle) === normalizeHandle(rootAuthor))
    }
    return !item.is_reply
  }

  function composeThread(extracted) {
    if (!extracted.length) return null

    // Pada halaman detail X, post dalam thread tampil sebagai article berurutan.
    // Gunakan author post utama sebagai pagar agar reply akun lain tidak ikut.
    const currentID = location.pathname.match(/\/status\/(\d+)/)?.[1] || ''
    if (!currentID) return { ...extracted[0], thread_post_count: 1, thread_posts: [extracted[0]] }
    const currentIndex = Math.max(0, extracted.findIndex((item) => item.external_post_id === currentID))
    const author = extracted[currentIndex]?.author_handle || extracted[0].author_handle
    // Reply akun lain boleh menyelip di antara post utama. Hanya post dari
    // author utama yang bukan reply, atau yang jelas self-reply, yang diambil.
    const isMainThreadPost = (item) => isThreadContinuation(item, author)
    const root = extracted[currentIndex] || extracted[0]
    const posts = [root]
    const seenIDs = new Set([root.external_post_id])
    for (const item of extracted) {
      if (seenIDs.has(item.external_post_id) || !isMainThreadPost(item)) continue
      seenIDs.add(item.external_post_id)
      posts.push(item)
    }
    const ordered = [...new Map(posts.map((item) => [item.external_post_id, item])).values()]
    const primary = ordered[0]
    const primaryMedia = [...new Set(primary.media || [])]
    return {
      ...primary,
      original_text: ordered.length === 1 ? primary.original_text : ordered.map((item, index) => `Post ${index + 1}\n${item.original_text}`).join('\n\n'),
      media: primaryMedia,
      thread_post_count: ordered.length,
      thread_posts: ordered.map(({ external_post_id, canonical_url, original_text, published_at }, index) => ({ external_post_id, canonical_url, original_text, media: index === 0 ? primaryMedia : [], published_at })),
    }
  }

  function extractThread(articles) {
    return composeThread(articles.map((article) => extractArticle(article)).filter((item) => item.external_post_id && item.original_text))
  }

  async function captureThread() {
    const initial = [...document.querySelectorAll('article[data-testid="tweet"]')]
    if (!initial.length) return null
    if (!location.pathname.match(/\/status\/\d+/)) return extractThread(initial)

    // X melakukan virtualisasi DOM. Kumpulkan snapshot per viewport agar post
    // thread yang keluar-masuk DOM tetap ikut. Posisi akhir dibiarkan di titik
    // pembacaan supaya pengguna bisa melihat proses auto-scroll memang selesai.
    const seen = new Map()
    window.scrollTo(0, 0)
    await new Promise((resolve) => setTimeout(resolve, 700))
    let stableRounds = 0
    let idleThreadRounds = 0
    const mainThreadIDs = new Set()
    const currentID = location.pathname.match(/\/status\/(\d+)/)?.[1] || ''
    for (let round = 0; round < 120; round++) {
      const articles = [...document.querySelectorAll('article[data-testid="tweet"]')]
      for (const article of articles) {
        const item = extractArticle(article)
        if (item.external_post_id) {
          if (!seen.has(item.external_post_id)) seen.set(item.external_post_id, item)
        }
      }
      const collectedItems = [...seen.values()]
      const beforeMainThreadCount = mainThreadIDs.size
      const rootIndex = collectedItems.findIndex((item) => item.external_post_id === currentID)
      if (rootIndex >= 0) {
        const rootItem = collectedItems[rootIndex]
        for (const item of collectedItems.slice(rootIndex + 1)) {
          if (
            item.external_post_id !== currentID &&
            isThreadContinuation(item, rootItem.author_handle)
          ) {
            mainThreadIDs.add(item.external_post_id)
          }
        }
      }
      const loadedNewMain = mainThreadIDs.size > beforeMainThreadCount
      if (loadedNewMain) idleThreadRounds = 0
      else if (mainThreadIDs.size > 0) idleThreadRounds++
      if (mainThreadIDs.size > 0 && idleThreadRounds >= 12) break
      const before = seen.size
      const beforeHeight = document.documentElement.scrollHeight
      window.scrollBy(0, Math.max(300, Math.floor(window.innerHeight * 0.65)))
      await new Promise((resolve) => setTimeout(resolve, 700))
      const atBottom = window.innerHeight + window.scrollY >= document.documentElement.scrollHeight - 24
      const loadedNewPost = seen.size > before
      const pageChanged = document.documentElement.scrollHeight > beforeHeight
      if (atBottom && !loadedNewPost && !pageChanged) {
        stableRounds++
        await new Promise((resolve) => setTimeout(resolve, 700))
        if (stableRounds >= 3) break
      } else {
        stableRounds = 0
      }
    }
    return composeThread([...seen.values()])
  }

  chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
    if (message.type !== 'AFFILIATOR_CAPTURE_X') return
    (async () => {
      const item = await captureThread()
      if (!item) { sendResponse({ ok: false, error: 'Data post X belum terbaca. Tunggu halaman selesai dimuat lalu coba lagi.' }); return }
      sendResponse({ ok: true, item })
    })().catch((error) => sendResponse({ ok: false, error: error.message || 'Gagal membaca thread X.' }))
    return true
  })
})()
