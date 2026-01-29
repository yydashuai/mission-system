import { defineStore } from 'pinia'
import { loadAppConfig } from '../utils/config'
import { pingApi } from '../services/api'

let clockTimer = null
let apiTimer = null

const formatTime = (date) => {
  if (!date) return '--'
  // 转换为中国时区（UTC+8）
  const chinaTime = new Date(date.getTime() + 8 * 60 * 60 * 1000)
  // 格式化为 YYYY-MM-DD HH:mm:ss
  return chinaTime.toISOString().slice(0, 19).replace('T', ' ')
}

const formatSeconds = (value) => `${Math.round(value / 1000)}s`

export const useSystemStore = defineStore('system', {
  state: () => ({
    config: loadAppConfig(),
    apiStatus: 'unknown',
    apiMessage: '',
    lastCheckedAt: null,
    now: new Date(),
  }),
  getters: {
    refreshLabel: (state) => formatSeconds(state.config.refreshInterval),
    detailPollLabel: (state) => formatSeconds(state.config.detailPollInterval),
    modeLabel: (state) => (state.config.readOnly ? '只读' : '读写'),
    authSummary: (state) => {
      if (!state.config.authToken) return '无'
      const scheme = state.config.authScheme ? `${state.config.authScheme} ` : ''
      return `${state.config.authHeader}: ${scheme}***`
    },
    authHeaders: (state) => {
      if (!state.config.authToken) return {}
      const scheme = state.config.authScheme ? `${state.config.authScheme} ` : ''
      const value = `${scheme}${state.config.authToken}`.trim()
      return { [state.config.authHeader]: value }
    },
    timeLabel: (state) => formatTime(state.now),
    lastCheckedLabel: (state) => formatTime(state.lastCheckedAt),
  },
  actions: {
    init() {
      this.config = loadAppConfig()
    },
    tick() {
      this.now = new Date()
    },
    startClock() {
      if (clockTimer) return
      this.tick()
      clockTimer = setInterval(() => this.tick(), 1000)
    },
    stopClock() {
      if (!clockTimer) return
      clearInterval(clockTimer)
      clockTimer = null
    },
    startApiPolling() {
      if (apiTimer) return
      const interval = this.config.refreshInterval
      if (!interval || interval <= 0) return
      this.checkApi({ silent: true })
      apiTimer = setInterval(() => {
        this.checkApi({ silent: true })
      }, interval)
    },
    stopApiPolling() {
      if (!apiTimer) return
      clearInterval(apiTimer)
      apiTimer = null
    },
    async checkApi(options = {}) {
      const silent = options.silent === true
      this.apiMessage = ''
      if (!silent) {
        this.apiStatus = 'checking'
      }
      const result = await pingApi(this.config.apiBase, { headers: this.authHeaders })
      this.lastCheckedAt = new Date()

      if (result.disabled) {
        this.apiStatus = 'disabled'
        this.apiMessage = result.error || 'API 地址未设置'
        return
      }

      if (result.ok) {
        this.apiStatus = 'ok'
        return
      }

      this.apiStatus = 'down'
      this.apiMessage = result.error || (result.status ? `HTTP ${result.status}` : '无法访问')
    },
    updateIntervals({ refreshMs, detailMs }) {
      if (Number.isFinite(refreshMs)) {
        this.config.refreshInterval = refreshMs
      }
      if (Number.isFinite(detailMs)) {
        this.config.detailPollInterval = detailMs
      }
    },
    setReadOnly(value) {
      this.config.readOnly = Boolean(value)
    },
  },
})
