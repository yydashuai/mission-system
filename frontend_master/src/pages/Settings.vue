<script setup>
import { computed } from 'vue'
import { useSystemStore } from '../stores/system'

const systemStore = useSystemStore()

const config = computed(() => ([
  { label: 'API Base', value: systemStore.config.apiBase || '--' },
  { label: 'API Mode', value: systemStore.config.apiMode },
  { label: 'Namespace', value: systemStore.config.namespace },
  { label: 'Auth', value: systemStore.authSummary },
  { label: 'Refresh Interval', value: systemStore.refreshLabel },
  { label: 'Detail Polling', value: systemStore.detailPollLabel },
  { label: 'Mode', value: systemStore.modeLabel },
]))

const apiBadge = computed(() => {
  if (systemStore.apiStatus === 'ok') return { label: 'Connected', tone: 'ok' }
  if (systemStore.apiStatus === 'checking') return { label: 'Checking', tone: 'warn' }
  if (systemStore.apiStatus === 'down') return { label: 'Down', tone: 'err' }
  if (systemStore.apiStatus === 'disabled') return { label: 'Disabled', tone: 'muted' }
  return { label: 'Unknown', tone: 'muted' }
})

const quickLinks = [
  { label: 'kubectl get missions -A', hint: 'List mission CRDs' },
  { label: 'kubectl get ft -A', hint: 'FlightTask summary' },
  { label: 'kubectl describe ft <name>', hint: 'Detailed scheduling view' },
]
</script>

<template>
  <section class="page">
    <header class="page-header">
      <div>
        <div class="page-title">Settings</div>
        <div class="page-sub">API connectivity, refresh, and access.</div>
      </div>
      <div class="page-actions">
        <button class="ghost small" type="button" @click="systemStore.checkApi">Test API</button>
        <button class="ghost small">Copy Config</button>
      </div>
    </header>

    <section class="split-grid">
      <div class="panel">
        <div class="panel-header">
          <div>
            <div class="panel-title">Connection</div>
            <div class="panel-sub">Runtime configuration injected via window.__APP_CONFIG__.</div>
          </div>
          <span class="badge" :class="apiBadge.tone">{{ apiBadge.label }}</span>
        </div>
        <div class="detail-card">
          <div class="detail-title">Frontend Config</div>
          <div class="detail-info">
            <div v-for="item in config" :key="item.label" class="kv">
              <span>{{ item.label }}</span>
              <span>{{ item.value }}</span>
            </div>
            <div v-if="systemStore.lastCheckedLabel !== '--'" class="kv">
              <span>Last Checked</span>
              <span>UTC {{ systemStore.lastCheckedLabel }}</span>
            </div>
          </div>
        </div>

        <div class="panel">
          <div class="panel-header">
            <div>
              <div class="panel-title">Controls</div>
              <div class="panel-sub">Runtime safety defaults.</div>
            </div>
          </div>
          <div class="switch-row">
            <div>
              <div class="cell-title">Read-only mode</div>
              <div class="cell-sub">Disable write operations from UI.</div>
            </div>
            <div class="switch" :class="{ 'is-on': systemStore.config.readOnly }"></div>
          </div>
          <div class="switch-row">
            <div>
              <div class="cell-title">Auto refresh</div>
              <div class="cell-sub">Poll mission status every 10 seconds.</div>
            </div>
            <div class="switch" :class="{ 'is-on': systemStore.config.refreshInterval > 0 }"></div>
          </div>
          <div class="switch-row">
            <div>
              <div class="cell-title">Verbose events</div>
              <div class="cell-sub">Show scheduler warnings and kubelet info.</div>
            </div>
            <div class="switch" :class="{ 'is-on': systemStore.config.verboseEvents }"></div>
          </div>
        </div>
      </div>

      <div class="panel">
        <div class="panel-header">
          <div>
            <div class="panel-title">Help & Shortcuts</div>
            <div class="panel-sub">Operator guidance for quick debugging.</div>
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
              <div class="panel-title">Security Notes</div>
              <div class="panel-sub">RBAC defaults and access policies.</div>
            </div>
          </div>
          <div class="note-list">
            <div class="note-item">UI runs in read-only mode unless API tokens are provided.</div>
            <div class="note-item">Cluster access is limited to airforce namespace resources.</div>
            <div class="note-item">Write operations require explicit confirmation + audit logging.</div>
          </div>
        </div>
      </div>
    </section>
  </section>
</template>
