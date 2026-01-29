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
  if (systemStore.apiStatus === 'ok') return 'API 正常'
  if (systemStore.apiStatus === 'checking') return 'API 检查中'
  if (systemStore.apiStatus === 'down') return 'API 断开'
  if (systemStore.apiStatus === 'disabled') return 'API 关闭'
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
        <div class="brand-title">任务控制中心</div>
        <div class="brand-sub">K8s 空军编排系统</div>
      </div>
    </div>
    <div class="topbar-status">
      <span class="badge" :class="statusTone">{{ statusLabel }}</span>
      <span class="badge">刷新 {{ systemStore.refreshLabel }}</span>
      <span class="badge muted">UTC {{ systemStore.timeLabel }}</span>
    </div>
    <div class="topbar-actions">
      <button class="ghost" type="button" @click="handleSync">同步</button>
      <button class="primary" type="button">启动演示</button>
    </div>
  </header>
</template>
