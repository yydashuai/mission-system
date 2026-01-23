const DEFAULT_CONFIG = {
  API_BASE: '',
  REFRESH_INTERVAL: 10000,
  DETAIL_POLL_INTERVAL: 5000,
  READ_ONLY: true,
  VERBOSE_EVENTS: false,
}

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

export const loadAppConfig = () => {
  const raw = typeof window !== 'undefined' ? window.__APP_CONFIG__ || {} : {}
  const merged = { ...DEFAULT_CONFIG, ...raw }

  return {
    apiBase: merged.API_BASE || '',
    refreshInterval: toNumber(merged.REFRESH_INTERVAL, DEFAULT_CONFIG.REFRESH_INTERVAL),
    detailPollInterval: toNumber(merged.DETAIL_POLL_INTERVAL, DEFAULT_CONFIG.DETAIL_POLL_INTERVAL),
    readOnly: toBoolean(merged.READ_ONLY, DEFAULT_CONFIG.READ_ONLY),
    verboseEvents: toBoolean(merged.VERBOSE_EVENTS, DEFAULT_CONFIG.VERBOSE_EVENTS),
  }
}
