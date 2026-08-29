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

  function extract(article) {
    const link = [...article.querySelectorAll('a[href*="/status/"]')].find((item) => /\/status\/\d+/.test(item.getAttribute('href') || ''))
    const href = link ? new URL(link.getAttribute('href'), location.origin).href : location.href
    const text = article.querySelector('[data-testid="tweetText"]')?.innerText?.trim() || article.innerText.trim()
    const author = [...article.querySelectorAll('a[href^="/"]')].map((item) => item.getAttribute('href')).find((value) => /^\/[A-Za-z0-9_]{1,15}$/.test(value || '')) || ''
    const time = article.querySelector('time')?.dateTime || ''
    const media = [...article.querySelectorAll('img, video')]
      .map((item) => item.currentSrc || item.src || item.querySelector('source')?.src)
      .filter((src) => src && !/profile_images|emoji|abs-0\.twimg\.com\/emoji/i.test(src))
    return {
      platform: 'x', canonical_url: href, external_post_id: href.match(/\/status\/(\d+)/)?.[1] || '',
      author_handle: author.replace(/^\//, ''), original_text: text, media: [...new Set(media)],
      published_at: time, stats: { like_count: metric(article, 'like'), repost_count: metric(article, 'repost|retweet'), reply_count: metric(article, 'repl'), view_count: metric(article, 'view') },
    }
  }

  chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
    if (message.type !== 'AFFILIATOR_CAPTURE_X') return
    const article = document.querySelector('article[data-testid="tweet"]')
    if (!article) { sendResponse({ ok: false, error: 'Buka satu halaman detail post X atau pastikan post terlihat.' }); return }
    sendResponse({ ok: true, item: extract(article) })
  })
})()
