import { defineStore } from 'pinia'
import {
  getClusterEvents,
  getClusterNodes,
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
  missions: clone(seedMissions),
  stages: clone(seedStages),
  flightTasks: clone(seedFlightTasks),
  weapons: clone(seedWeapons),
  nodes: clone(seedClusterNodes),
  events: clone(seedClusterEvents),
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
    async loadResource(key, loader, normalizer, fallback) {
      this.loading[key] = true
      this.errors[key] = ''
      try {
        const payload = await loader()
        const normalized = normalizer ? normalizer(payload) : []
        this[key] = clone(normalized)
      } catch (error) {
        this.errors[key] = error?.message || 'load failed'
        this[key] = clone(fallback)
      } finally {
        this.loading[key] = false
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
        this.loadResource('missions', () => getMissions(baseUrl, options), normalizeMissionList, seedMissions),
        this.loadResource('stages', () => getStages(baseUrl, options), normalizeStageList, seedStages),
        this.loadResource('flightTasks', () => getFlightTasks(baseUrl, options), normalizeFlightTaskList, seedFlightTasks),
        this.loadResource('weapons', () => getWeapons(baseUrl, options), normalizeWeaponList, seedWeapons),
        this.loadResource('nodes', () => getClusterNodes(baseUrl, options), normalizeNodeList, seedClusterNodes),
        this.loadResource('events', () => getClusterEvents(baseUrl, options), normalizeEventList, seedClusterEvents),
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
