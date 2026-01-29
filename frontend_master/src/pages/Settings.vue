<script setup>
import { computed, ref, watch } from 'vue'
import { useSystemStore } from '../stores/system'
import { useDataStore } from '../stores/data'
import { saveToStorage } from '../utils/config'

const systemStore = useSystemStore()
const dataStore = useDataStore()

const refreshSeconds = ref(Math.round(systemStore.config.refreshInterval / 1000))
const hasChanges = ref(false)

watch(() => systemStore.config.refreshInterval, (value) => {
  const seconds = Math.round(value / 1000)
  if (refreshSeconds.value !== seconds) {
    refreshSeconds.value = seconds
  }
})

watch([refreshSeconds, () => systemStore.config.refreshInterval, () => systemStore.config.detailPollInterval], () => {
  const currentRefresh = Math.round(systemStore.config.refreshInterval / 1000)
  const currentDetail = Math.round(systemStore.config.detailPollInterval / 1000)
  hasChanges.value = refreshSeconds.value !== currentRefresh || refreshSeconds.value !== currentDetail
})

const config = computed(() => ([
  { label: 'API 基址', value: systemStore.config.apiBase || '(同源)' },
  { label: 'API 模式', value: systemStore.config.apiMode },
  { label: '命名空间', value: systemStore.config.namespace },
  { label: '认证', value: systemStore.authSummary },
]))

const apiBadge = computed(() => {
  if (systemStore.apiStatus === 'ok') return { label: '已连接', tone: 'ok' }
  if (systemStore.apiStatus === 'checking') return { label: '检测中', tone: 'warn' }
  if (systemStore.apiStatus === 'down') return { label: '断开', tone: 'err' }
  if (systemStore.apiStatus === 'disabled') return { label: '已禁用', tone: 'muted' }
  return { label: '未知', tone: 'muted' }
})

const clampSeconds = (value) => {
  const parsed = Number(value)
  if (!Number.isFinite(parsed)) return null
  return Math.max(1, Math.round(parsed))
}

const saveSettings = () => {
  const refreshValue = clampSeconds(refreshSeconds.value)
  if (refreshValue === null) return

  const newConfig = {
    ...systemStore.config,
    refreshInterval: refreshValue * 1000,
    detailPollInterval: refreshValue * 1000,
  }

  // 保存到 localStorage
  saveToStorage(newConfig)

  // 更新 store
  systemStore.updateIntervals({
    refreshMs: refreshValue * 1000,
    detailMs: refreshValue * 1000,
  })

  // 重启轮询
  dataStore.stopPolling()
  dataStore.startPolling()
  systemStore.stopApiPolling()
  systemStore.startApiPolling()

  hasChanges.value = false
}
</script>

<template>
  <section class="page">
    <header class="page-header">
      <div>
        <div class="page-title">系统设置</div>
        <div class="page-sub">API 连接与刷新配置</div>
      </div>
    </header>

    <section class="settings-layout">
      <div class="panel">
        <div class="panel-header">
          <div>
            <div class="panel-title">连接状态</div>
            <div class="panel-sub">当前 API 连接信息</div>
          </div>
          <div class="panel-actions">
            <button class="ghost small" type="button" @click="systemStore.checkApi">测试 API</button>
            <span class="badge" :class="apiBadge.tone">{{ apiBadge.label }}</span>
          </div>
        </div>
        <div class="detail-card">
          <div class="detail-info">
            <div v-for="item in config" :key="item.label" class="kv">
              <span>{{ item.label }}</span>
              <span>{{ item.value }}</span>
            </div>
            <div v-if="systemStore.lastCheckedLabel !== '--'" class="kv">
              <span>最后检测</span>
              <span>{{ systemStore.lastCheckedLabel }}</span>
            </div>
          </div>
        </div>
      </div>

      <div class="panel">
        <div class="panel-header">
          <div>
            <div class="panel-title">刷新设置</div>
            <div class="panel-sub">配置数据轮询间隔</div>
          </div>
        </div>
        <div class="switch-row">
          <div>
            <div class="cell-title">刷新间隔</div>
            <div class="cell-sub">列表与详情数据的统一刷新周期</div>
          </div>
          <div class="interval-input">
            <input
              v-model.number="refreshSeconds"
              class="input"
              type="number"
              min="1"
            />
            <span class="muted">秒</span>
          </div>
        </div>
        <div class="switch-row is-actions">
          <button
            class="primary small"
            type="button"
            :disabled="!hasChanges"
            @click="saveSettings"
          >
            保存设置
          </button>
        </div>
      </div>
    </section>
  </section>
</template>

<style scoped>
.settings-layout {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
  max-width: 900px;
}

.panel-actions {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
}
</style>
