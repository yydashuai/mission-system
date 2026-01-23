<script setup>
import { computed } from 'vue'
import { useSystemStore } from '../../stores/system'

const systemStore = useSystemStore()

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
</script>

<template>
  <header class="topbar">
    <div class="brand">
      <span class="brand-mark">AFC</span>
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
      <button class="ghost" type="button" @click="systemStore.checkApi">Sync</button>
      <button class="primary" type="button">Launch Demo</button>
    </div>
  </header>
</template>
