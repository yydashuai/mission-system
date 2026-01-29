import { defineStore } from 'pinia'
import {
  getClusterEvents,
  getClusterNodes,
  getClusterNodeMetrics,
  getClusterPods,
  getFlightTasks,
  getMissions,
  getStages,
  getWeapons,
} from '../services/api'
import {
  seedClusterEvents,
  seedClusterNodes,
  seedFlightTasks,
  seedMissions,
  seedStages,
  seedWeapons,
} from '../data/seed'
import {
  normalizeEventList,
  normalizeFlightTaskList,
  normalizeMissionList,
  normalizeNodeList,
  normalizeStageList,
  normalizeWeaponList,
} from '../data/normalize'
import { useSystemStore } from './system'

let refreshTimer = null

const clone = (value) => JSON.parse(JSON.stringify(value))

const stateDefaults = () => ({
  missions: normalizeMissionList(clone(seedMissions)),
  stages: normalizeStageList(clone(seedStages)),
  flightTasks: normalizeFlightTaskList(clone(seedFlightTasks)),
  weapons: normalizeWeaponList(clone(seedWeapons)),
  nodes: normalizeNodeList(clone(seedClusterNodes)),
  events: normalizeEventList(clone(seedClusterEvents)),
  loading: {
    missions: false,
    stages: false,
    flightTasks: false,
    weapons: false,
    nodes: false,
    events: false,
  },
  errors: {
    missions: '',
    stages: '',
    flightTasks: '',
    weapons: '',
    nodes: '',
    events: '',
  },
  lastUpdated: null,
})

export const useDataStore = defineStore('data', {
  state: stateDefaults,
  actions: {
    async loadResource(key, loader, normalizer) {
      const current = this[key]
      const showLoading = !Array.isArray(current) || current.length === 0
      if (showLoading) {
        this.loading[key] = true
      }
      this.errors[key] = ''
      try {
        const payload = await loader()
        const normalized = normalizer ? normalizer(payload) : []
        this[key] = clone(normalized)
      } catch (error) {
        this.errors[key] = error?.message || 'load failed'
        if (!Array.isArray(this[key])) {
          this[key] = []
        }
      } finally {
        if (showLoading) {
          this.loading[key] = false
        }
      }
    },
    async loadNodesWithMetrics(baseUrl, options) {
      const showLoading = !Array.isArray(this.nodes) || this.nodes.length === 0
      if (showLoading) {
        this.loading.nodes = true
      }
      this.errors.nodes = ''
      try {
        const nodesPayload = await getClusterNodes(baseUrl, options)
        let metricsPayload = null
        let metricsError = null
        let podsPayload = null
        let podsError = null
        try {
          metricsPayload = await getClusterNodeMetrics(baseUrl, options)
        } catch (error) {
          metricsError = error
        }
        try {
          podsPayload = await getClusterPods(baseUrl, options)
        } catch (error) {
          podsError = error
        }

        const normalized = normalizeNodeList(nodesPayload, metricsPayload, podsPayload)
        this.nodes = clone(normalized)

        const warnings = []
        if (metricsError) warnings.push(metricsError?.message || 'metrics unavailable')
        if (podsError) warnings.push(podsError?.message || 'pods unavailable')
        if (warnings.length) {
          this.errors.nodes = warnings.join('; ')
        }
      } catch (error) {
        this.errors.nodes = error?.message || 'load failed'
        if (!Array.isArray(this.nodes)) {
          this.nodes = []
        }
      } finally {
        if (showLoading) {
          this.loading.nodes = false
        }
      }
    },
    async refreshAll() {
      const systemStore = useSystemStore()
      const baseUrl = systemStore.config.apiBase
      const options = {
        mode: systemStore.config.apiMode,
        namespace: systemStore.config.namespace,
        headers: systemStore.authHeaders,
      }

      await Promise.all([
        this.loadResource('missions', () => getMissions(baseUrl, options), normalizeMissionList),
        this.loadResource('stages', () => getStages(baseUrl, options), normalizeStageList),
        this.loadResource('flightTasks', () => getFlightTasks(baseUrl, options), normalizeFlightTaskList),
        this.loadResource('weapons', () => getWeapons(baseUrl, options), normalizeWeaponList),
        this.loadNodesWithMetrics(baseUrl, options),
        this.loadResource('events', () => getClusterEvents(baseUrl, options), normalizeEventList),
      ])

      this.lastUpdated = new Date()
    },
    startPolling() {
      if (refreshTimer) return
      const systemStore = useSystemStore()
      const interval = systemStore.config.refreshInterval
      if (!interval || interval <= 0) return

      refreshTimer = setInterval(() => {
        this.refreshAll()
      }, interval)
    },
    stopPolling() {
      if (!refreshTimer) return
      clearInterval(refreshTimer)
      refreshTimer = null
    },
  },
})
