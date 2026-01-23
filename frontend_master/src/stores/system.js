import { defineStore } from 'pinia'
import { loadAppConfig } from '../utils/config'
import { pingApi } from '../services/api'

let clockTimer = null

const formatTime = (date) => {
  if (!date) return '--'
  return date.toISOString().slice(11, 16)
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
    modeLabel: (state) => (state.config.readOnly ? 'Read-only' : 'Read-write'),
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
      clockTimer = setInterval(() => this.tick(), 60000)
    },
    stopClock() {
      if (!clockTimer) return
      clearInterval(clockTimer)
      clockTimer = null
    },
    async checkApi() {
      this.apiMessage = ''
      if (!this.config.apiBase) {
        this.apiStatus = 'disabled'
        this.lastCheckedAt = new Date()
        return
      }

      this.apiStatus = 'checking'
      const result = await pingApi(this.config.apiBase)
      this.lastCheckedAt = new Date()

      if (result.disabled) {
        this.apiStatus = 'disabled'
        this.apiMessage = result.error || 'API base not set'
        return
      }

      if (result.ok) {
        this.apiStatus = 'ok'
        return
      }

      this.apiStatus = 'down'
      this.apiMessage = result.error || (result.status ? `HTTP ${result.status}` : 'Unreachable')
    },
  },
})
