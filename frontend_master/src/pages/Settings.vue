<script setup>
import { computed, ref, watch } from 'vue'
import { useSystemStore } from '../stores/system'
import { useDataStore } from '../stores/data'

const systemStore = useSystemStore()
const dataStore = useDataStore()

const refreshSeconds = ref(Math.round(systemStore.config.refreshInterval / 1000))
const detailSeconds = ref(Math.round(systemStore.config.detailPollInterval / 1000))

const isEditable = computed(() => !systemStore.config.readOnly)

watch(() => systemStore.config.refreshInterval, (value) => {
  refreshSeconds.value = Math.round(value / 1000)
})

watch(() => systemStore.config.detailPollInterval, (value) => {
  detailSeconds.value = Math.round(value / 1000)
})

const config = computed(() => ([
  { label: 'API 基址', value: systemStore.config.apiBase || '--' },
  { label: 'API 模式', value: systemStore.config.apiMode },
  { label: '命名空间', value: systemStore.config.namespace },
  { label: '认证', value: systemStore.authSummary },
  { label: '刷新间隔', value: systemStore.refreshLabel },
  { label: '详情轮询', value: systemStore.detailPollLabel },
  { label: '模式', value: systemStore.modeLabel },
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
  return Math.max(0, Math.round(parsed))
}

const applyIntervals = () => {
  if (!isEditable.value) return
  const refreshValue = clampSeconds(refreshSeconds.value)
  const detailValue = clampSeconds(detailSeconds.value)
  if (refreshValue === null || detailValue === null) return

  systemStore.updateIntervals({
    refreshMs: refreshValue * 1000,
    detailMs: detailValue * 1000,
  })
  dataStore.stopPolling()
  dataStore.startPolling()
  systemStore.stopApiPolling()
  systemStore.startApiPolling()
}

const toggleReadOnly = () => {
  systemStore.setReadOnly(!systemStore.config.readOnly)
}

const quickLinks = [
  { label: 'kubectl get missions -A', hint: '列出任务 CRD' },
  { label: 'kubectl get ft -A', hint: '飞行任务概览' },
  { label: 'kubectl describe ft <name>', hint: '详细调度视图' },
]
</script>

<template>
  <section class="page">
    <header class="page-header">
      <div>
        <div class="page-title">系统设置</div>
        <div class="page-sub">API 连接、刷新与访问控制。</div>
      </div>
      <div class="page-actions">
        <button class="ghost small" type="button" @click="systemStore.checkApi">测试 API</button>
        <button class="ghost small">复制配置</button>
      </div>
    </header>

    <section class="split-grid">
      <div class="panel">
        <div class="panel-header">
          <div>
            <div class="panel-title">连接状态</div>
            <div class="panel-sub">通过 window.__APP_CONFIG__ 注入的运行时配置。</div>
          </div>
          <span class="badge" :class="apiBadge.tone">{{ apiBadge.label }}</span>
        </div>
        <div class="detail-card">
          <div class="detail-title">前端配置</div>
          <div class="detail-info">
            <div v-for="item in config" :key="item.label" class="kv">
              <span>{{ item.label }}</span>
              <span>{{ item.value }}</span>
            </div>
            <div v-if="systemStore.lastCheckedLabel !== '--'" class="kv">
              <span>最后检测</span>
              <span>UTC {{ systemStore.lastCheckedLabel }}</span>
            </div>
          </div>
        </div>

        <div class="panel">
          <div class="panel-header">
            <div>
              <div class="panel-title">控制选项</div>
              <div class="panel-sub">运行时安全默认设置。</div>
            </div>
          </div>
          <div class="switch-row">
            <div>
              <div class="cell-title">只读模式</div>
              <div class="cell-sub">禁用界面写操作。</div>
            </div>
            <button
              class="switch"
              type="button"
              :class="{ 'is-on': systemStore.config.readOnly }"
              :aria-checked="systemStore.config.readOnly"
              role="switch"
              @click="toggleReadOnly"
            ></button>
          </div>
          <div class="switch-row">
            <div>
              <div class="cell-title">刷新间隔</div>
              <div class="cell-sub">列表数据轮询周期。</div>
            </div>
            <div class="interval-input">
              <input
                v-model.number="refreshSeconds"
                class="input"
                type="number"
                min="0"
                :disabled="!isEditable"
              />
              <span class="muted">s</span>
            </div>
          </div>
          <div class="switch-row">
            <div>
              <div class="cell-title">详情轮询</div>
              <div class="cell-sub">单项刷新周期。</div>
            </div>
            <div class="interval-input">
              <input
                v-model.number="detailSeconds"
                class="input"
                type="number"
                min="0"
                :disabled="!isEditable"
              />
              <span class="muted">s</span>
            </div>
          </div>
          <div class="switch-row">
            <div>
              <div class="cell-title">应用间隔</div>
              <div class="cell-sub">启用读写模式以编辑。</div>
            </div>
            <button class="ghost small" type="button" :disabled="!isEditable" @click="applyIntervals">
              应用
            </button>
          </div>
          <div class="switch-row">
            <div>
              <div class="cell-title">自动刷新</div>
              <div class="cell-sub">每 10 秒轮询任务状态。</div>
            </div>
            <div class="switch" :class="{ 'is-on': systemStore.config.refreshInterval > 0 }"></div>
          </div>
          <div class="switch-row">
            <div>
              <div class="cell-title">详细事件</div>
              <div class="cell-sub">显示调度器警告和 kubelet 信息。</div>
            </div>
            <div class="switch" :class="{ 'is-on': systemStore.config.verboseEvents }"></div>
          </div>
        </div>
      </div>

      <div class="panel">
        <div class="panel-header">
          <div>
            <div class="panel-title">帮助与快捷方式</div>
            <div class="panel-sub">操作员快速调试指南。</div>
          </div>
        </div>
        <div class="command-list">
          <div v-for="item in quickLinks" :key="item.label" class="command-item">
            <code>{{ item.label }}</code>
            <span class="muted">{{ item.hint }}</span>
          </div>
        </div>

        <div class="panel">
          <div class="panel-header">
            <div>
              <div class="panel-title">安全说明</div>
              <div class="panel-sub">RBAC 默认设置与访问策略。</div>
            </div>
          </div>
          <div class="note-list">
            <div class="note-item">除非提供 API 令牌，否则界面以只读模式运行。</div>
            <div class="note-item">集群访问仅限于 airforce 命名空间资源。</div>
            <div class="note-item">写操作需要明确确认并记录审计日志。</div>
          </div>
        </div>
      </div>
    </section>
  </section>
</template>
