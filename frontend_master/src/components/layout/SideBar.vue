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
  if (!apiOk.value) return '加载中'
  if (!totalNodes.value) return '--'
  return `${readyNodes.value} / ${totalNodes.value}`
})

const nodesMeta = computed(() => {
  if (!apiOk.value) return '控制平面: --'
  if (!controlNode.value) return '控制平面: --'
  return `控制平面: ${controlNode.value.status}`
})

const eventCount = computed(() => events.value.length)
const warnCount = computed(() => events.value.filter((item) => item.level === 'warn' || item.level === 'err').length)

const eventsValue = computed(() => {
  if (!apiOk.value) return '加载中'
  if (!eventCount.value) return '--'
  return String(eventCount.value)
})

const eventsMeta = computed(() => {
  if (!apiOk.value) return '加载中'
  if (!eventCount.value) return '无事件'
  return warnCount.value ? `${warnCount.value} 条警告` : '无警告'
})
</script>

<template>
  <aside class="sidebar">
    <div class="section-title">导航</div>
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

    <div class="section-title">系统健康</div>
    <div class="health-grid">
      <div class="health-card">
        <div class="health-title">节点就绪</div>
        <div class="health-value">{{ nodesValue }}</div>
        <div class="health-meta">{{ nodesMeta }}</div>
      </div>
      <div class="health-card">
        <div class="health-title">事件 (24h)</div>
        <div class="health-value">{{ eventsValue }}</div>
        <div class="health-meta">{{ eventsMeta }}</div>
      </div>
    </div>
  </aside>
</template>
