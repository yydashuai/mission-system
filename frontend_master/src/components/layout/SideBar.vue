<script setup>
import { computed } from 'vue'
import { storeToRefs } from 'pinia'
import { useRoute } from 'vue-router'
import { useUiStore } from '../../stores/ui'
import { useDataStore } from '../../stores/data'
import { useSystemStore } from '../../stores/system'

const route = useRoute()
const uiStore = useUiStore()
const dataStore = useDataStore()
const systemStore = useSystemStore()
const { nodes, events } = storeToRefs(dataStore)

const isActive = (path) => route.path === path
const apiOk = computed(() => systemStore.apiStatus === 'ok')

const totalNodes = computed(() => nodes.value.length)
const readyNodes = computed(() => nodes.value.filter((item) => item.status === 'Ready').length)
const controlNode = computed(() => (
  nodes.value.find((item) => item.role === 'control-plane' || item.role === 'master') || null
))

const nodesValue = computed(() => {
  if (!apiOk.value) return 'Loading'
  if (!totalNodes.value) return '--'
  return `${readyNodes.value} / ${totalNodes.value}`
})

const nodesMeta = computed(() => {
  if (!apiOk.value) return 'Control Plane: --'
  if (!controlNode.value) return 'Control Plane: --'
  return `Control Plane: ${controlNode.value.status}`
})

const eventCount = computed(() => events.value.length)
const warnCount = computed(() => events.value.filter((item) => item.level === 'warn' || item.level === 'err').length)

const eventsValue = computed(() => {
  if (!apiOk.value) return 'Loading'
  if (!eventCount.value) return '--'
  return String(eventCount.value)
})

const eventsMeta = computed(() => {
  if (!apiOk.value) return 'Loading'
  if (!eventCount.value) return 'No events'
  return warnCount.value ? `${warnCount.value} warnings` : 'No warnings'
})
</script>

<template>
  <aside class="sidebar">
    <div class="section-title">Navigation</div>
    <nav class="nav-list">
      <router-link
        v-for="item in uiStore.navItems"
        :key="item.label"
        :to="item.path"
        class="nav-item"
        :class="{ active: isActive(item.path) }"
      >
        <span class="nav-dot"></span>
        {{ item.label }}
      </router-link>
    </nav>

    <div class="section-title">System Health</div>
    <div class="health-grid">
      <div class="health-card">
        <div class="health-title">Nodes Ready</div>
        <div class="health-value">{{ nodesValue }}</div>
        <div class="health-meta">{{ nodesMeta }}</div>
      </div>
      <div class="health-card">
        <div class="health-title">Events (24h)</div>
        <div class="health-value">{{ eventsValue }}</div>
        <div class="health-meta">{{ eventsMeta }}</div>
      </div>
    </div>
  </aside>
</template>
