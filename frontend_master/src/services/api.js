const withTimeout = (promise, timeoutMs) => new Promise((resolve, reject) => {
  const timer = setTimeout(() => {
    reject(new Error('request timeout'))
  }, timeoutMs)

  promise
    .then((value) => {
      clearTimeout(timer)
      resolve(value)
    })
    .catch((error) => {
      clearTimeout(timer)
      reject(error)
    })
})

const resolveUrl = (base, path) => {
  if (!base) return null
  try {
    return new URL(path, base).toString()
  } catch (error) {
    return null
  }
}

export const pingApi = async (baseUrl) => {
  const url = resolveUrl(baseUrl, '/healthz')
  if (!url) {
    return { ok: false, disabled: true, error: 'API base not set' }
  }

  try {
    const response = await withTimeout(fetch(url, { method: 'GET' }), 3000)
    return { ok: response.ok, status: response.status }
  } catch (error) {
    return { ok: false, error: error?.message || 'request failed' }
  }
}
