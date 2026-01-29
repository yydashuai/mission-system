const DEFAULT_CONFIG = {
  API_BASE: '',
  API_MODE: 'k8s',
  NAMESPACE: 'default',
  AUTH_HEADER: 'Authorization',
  AUTH_SCHEME: 'Bearer',
  AUTH_TOKEN: '',
  REFRESH_INTERVAL: 10000,
  DETAIL_POLL_INTERVAL: 5000,
  READ_ONLY: false,
  VERBOSE_EVENTS: false,
}

const STORAGE_KEY = 'airforce_config'

const toNumber = (value, fallback) => {
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : fallback
}

const toBoolean = (value, fallback) => {
  if (typeof value === 'boolean') return value
  if (typeof value === 'string') {
    const normalized = value.toLowerCase().trim()
    if (normalized === 'true') return true
    if (normalized === 'false') return false
  }
  return fallback
}

const loadFromStorage = () => {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    return stored ? JSON.parse(stored) : {}
  } catch {
    return {}
  }
}

export const saveToStorage = (config) => {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(config))
    return true
  } catch {
    return false
  }
}

export const loadAppConfig = () => {
  const raw = typeof window !== 'undefined' ? window.__APP_CONFIG__ || {} : {}
  const stored = loadFromStorage()
  const merged = { ...DEFAULT_CONFIG, ...raw, ...stored }

  return {
    apiBase: merged.API_BASE || merged.apiBase || '',
    apiMode: merged.API_MODE || merged.apiMode || DEFAULT_CONFIG.API_MODE,
    namespace: merged.NAMESPACE || merged.namespace || DEFAULT_CONFIG.NAMESPACE,
    authHeader: merged.AUTH_HEADER || merged.authHeader || DEFAULT_CONFIG.AUTH_HEADER,
    authScheme: merged.AUTH_SCHEME || merged.authScheme || DEFAULT_CONFIG.AUTH_SCHEME,
    authToken: merged.AUTH_TOKEN || merged.authToken || DEFAULT_CONFIG.AUTH_TOKEN,
    refreshInterval: toNumber(merged.REFRESH_INTERVAL || merged.refreshInterval, DEFAULT_CONFIG.REFRESH_INTERVAL),
    detailPollInterval: toNumber(merged.DETAIL_POLL_INTERVAL || merged.detailPollInterval, DEFAULT_CONFIG.DETAIL_POLL_INTERVAL),
    readOnly: toBoolean(merged.READ_ONLY !== undefined ? merged.READ_ONLY : merged.readOnly, DEFAULT_CONFIG.READ_ONLY),
    verboseEvents: toBoolean(merged.VERBOSE_EVENTS !== undefined ? merged.VERBOSE_EVENTS : merged.verboseEvents, DEFAULT_CONFIG.VERBOSE_EVENTS),
  }
}
