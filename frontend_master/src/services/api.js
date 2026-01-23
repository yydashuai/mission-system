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
  if (!path) return null
  if (!base) return path
  try {
    return new URL(path, base).toString()
  } catch (error) {
    return null
  }
}

const buildUrl = (base, path) => {
  const url = resolveUrl(base, path)
  if (!url) {
    throw new Error('API base not set')
  }
  return url
}

const fetchJson = async (url, options = {}) => {
  const response = await withTimeout(fetch(url, options), 8000)
  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`)
  }
  return response.json()
}

const resolveResourcePath = (resource, options = {}) => {
  const mode = (options.mode || 'gateway').toLowerCase()
  const namespace = options.namespace || 'default'

  if (mode === 'k8s') {
    const base = `/apis/airforce.airforce.mil/v1alpha1/namespaces/${namespace}`
    if (resource === 'missions') return `${base}/missions`
    if (resource === 'stages') return `${base}/missionstages`
    if (resource === 'flighttasks') return `${base}/flighttasks`
    if (resource === 'weapons') return `${base}/weapons`
    if (resource === 'nodes') return '/api/v1/nodes'
    if (resource === 'events') return `/api/v1/namespaces/${namespace}/events`
  }

  if (resource === 'missions') return '/api/missions'
  if (resource === 'stages') return '/api/stages'
  if (resource === 'flighttasks') return '/api/flighttasks'
  if (resource === 'weapons') return '/api/weapons'
  if (resource === 'nodes') return '/api/cluster/nodes'
  if (resource === 'events') return '/api/cluster/events'
  return ''
}

export const pingApi = async (baseUrl, options = {}) => {
  const url = resolveUrl(baseUrl, '/healthz')
  if (!url) {
    return { ok: false, disabled: true, error: 'API base not set' }
  }

  try {
    const response = await withTimeout(fetch(url, {
      method: 'GET',
      headers: options.headers,
    }), 3000)
    return { ok: response.ok, status: response.status }
  } catch (error) {
    return { ok: false, error: error?.message || 'request failed' }
  }
}

export const getMissions = async (baseUrl, options = {}) => (
  fetchJson(buildUrl(baseUrl, resolveResourcePath('missions', options)), {
    headers: options.headers,
  })
)

export const getStages = async (baseUrl, options = {}) => (
  fetchJson(buildUrl(baseUrl, resolveResourcePath('stages', options)), {
    headers: options.headers,
  })
)

export const getFlightTasks = async (baseUrl, options = {}) => (
  fetchJson(buildUrl(baseUrl, resolveResourcePath('flighttasks', options)), {
    headers: options.headers,
  })
)

export const getWeapons = async (baseUrl, options = {}) => (
  fetchJson(buildUrl(baseUrl, resolveResourcePath('weapons', options)), {
    headers: options.headers,
  })
)

export const getClusterNodes = async (baseUrl, options = {}) => (
  fetchJson(buildUrl(baseUrl, resolveResourcePath('nodes', options)), {
    headers: options.headers,
  })
)

export const getClusterEvents = async (baseUrl, options = {}) => (
  fetchJson(buildUrl(baseUrl, resolveResourcePath('events', options)), {
    headers: options.headers,
  })
)
