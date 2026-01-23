<script setup>
import { computed } from 'vue'
import { useSystemStore } from '../../stores/system'
import { useDataStore } from '../../stores/data'

const systemStore = useSystemStore()
const dataStore = useDataStore()

const statusTone = computed(() => {
  if (systemStore.apiStatus === 'ok') return 'ok'
  if (systemStore.apiStatus === 'checking') return 'warn'
  if (systemStore.apiStatus === 'down') return 'err'
  return 'muted'
})

const statusLabel = computed(() => {
  if (systemStore.apiStatus === 'ok') return 'API OK'
  if (systemStore.apiStatus === 'checking') return 'API CHECK'
  if (systemStore.apiStatus === 'down') return 'API DOWN'
  if (systemStore.apiStatus === 'disabled') return 'API OFF'
  return 'API --'
})

const handleSync = async () => {
  systemStore.tick()
  await systemStore.checkApi()
  await dataStore.refreshAll()
}
</script>

<template>
  <header class="topbar">
    <div class="brand">
      <img class="brand-mark" src="/airforce-icon.png" alt="AFC" />
      <div>
        <div class="brand-title">Mission Control</div>
        <div class="brand-sub">K8s Airforce Orchestration</div>
      </div>
    </div>
    <div class="topbar-status">
      <span class="badge" :class="statusTone">{{ statusLabel }}</span>
      <span class="badge">Refresh {{ systemStore.refreshLabel }}</span>
      <span class="badge muted">UTC {{ systemStore.timeLabel }}</span>
    </div>
    <div class="topbar-actions">
      <button class="ghost" type="button" @click="handleSync">Sync</button>
      <button class="primary" type="button">Launch Demo</button>
    </div>
  </header>
</template>
