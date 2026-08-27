(() => {
  if (window.__AFFILIATOR_SHOPEE_NETWORK_HOOK__) return
  window.__AFFILIATOR_SHOPEE_NETWORK_HOOK__ = true
  const interesting = (url) => /shopee|susercontent|\/api\/v\d+\/(?:item|product|shop|pdp|model|search)/i.test(String(url || ''))
  const publish = (url, data) => {
    if (!interesting(url) || !data || typeof data !== 'object') return
    window.postMessage({ type: 'AFFILIATOR_SHOPEE_NETWORK_RESPONSE', url: String(url), data }, '*')
  }
  const fetchOriginal = window.fetch
  window.fetch = function (...args) {
    return fetchOriginal.apply(this, args).then((response) => {
      const url = typeof args[0] === 'string' ? args[0] : args[0]?.url
      response.clone().json().then((data) => publish(url, data)).catch(() => {})
      return response
    })
  }
  const openOriginal = XMLHttpRequest.prototype.open
  const sendOriginal = XMLHttpRequest.prototype.send
  XMLHttpRequest.prototype.open = function (method, url, ...rest) {
    this.__affiliatorURL = url
    return openOriginal.call(this, method, url, ...rest)
  }
  XMLHttpRequest.prototype.send = function (...args) {
    this.addEventListener('load', () => {
      if (!interesting(this.__affiliatorURL)) return
      try { publish(this.__affiliatorURL, JSON.parse(this.responseText)) } catch {}
    })
    return sendOriginal.apply(this, args)
  }
})()
