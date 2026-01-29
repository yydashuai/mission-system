<script setup>
import { computed } from 'vue'
import { useFocusStore } from '../../stores/focus'

const focusStore = useFocusStore()

const focusLabel = computed(() => {
  const labels = {
    mission: '任务',
    stage: '任务阶段',
    flighttask: '飞行任务',
    weapon: '武器',
    node: '节点',
  }
  return labels[focusStore.kind] || '焦点'
})

const safe = (value) => {
  if (value === undefined || value === null || value === '') return '--'
  return value
}

const joinList = (value) => {
  if (!Array.isArray(value) || value.length === 0) return '--'
  return value.join(', ')
}

const phaseTone = (value) => {
  const key = String(value || '').toLowerCase()
  if (['running', 'scheduled', 'pending', 'failed', 'succeeded'].includes(key)) return key
  return 'muted'
}

const priorityTone = (value) => {
  const key = String(value || '').toLowerCase()
  if (key === 'high' || key === 'critical') return 'warn'
  if (key === 'low') return 'muted'
  return 'ok'
}

const weaponTone = (value) => {
  if (value === 'Available') return 'ok'
  if (value === 'Updating' || value === 'Degraded') return 'warn'
  if (value === 'Deprecated' || value === 'Retired') return 'muted'
  return 'muted'
}

const focusTitle = computed(() => safe(focusStore.item?.name))

const focusBadges = computed(() => {
  const item = focusStore.item || {}
  if (focusStore.kind === 'mission') {
    return [
      { label: safe(item.status), tone: phaseTone(item.status) },
      { label: `${safe(item.priority)} 优先级`, tone: priorityTone(item.priority) },
    ]
  }
  if (focusStore.kind === 'stage') {
    return [
      { label: safe(item.status), tone: phaseTone(item.status) },
      { label: safe(item.mode), tone: 'muted' },
    ]
  }
  if (focusStore.kind === 'flighttask') {
    return [
      { label: safe(item.status), tone: phaseTone(item.status) },
      { label: `Pod ${safe(item.podStatus)}`, tone: phaseTone(item.podStatus) },
    ]
  }
  if (focusStore.kind === 'weapon') {
    const version = item.version ? `v${item.version}` : 'v--'
    return [
      { label: safe(item.status), tone: weaponTone(item.status) },
      { label: version, tone: 'muted' },
    ]
  }
  if (focusStore.kind === 'node') {
    const tone = item.status === 'Ready' ? 'ok' : 'err'
    return [
      { label: safe(item.status), tone },
      { label: safe(item.role), tone: 'muted' },
    ]
  }
  return []
})

const focusInfo = computed(() => {
  const item = focusStore.item || {}
  if (focusStore.kind === 'mission') {
    return [
      { label: '类型', value: safe(item.type) },
      { label: '指挥官', value: safe(item.commander) },
      { label: '区域', value: safe(item.region) },
      { label: '任务数', value: safe(item.tasks) },
      { label: '更新时间', value: safe(item.updated) },
    ]
  }
  if (focusStore.kind === 'stage') {
    return [
      { label: '任务', value: safe(item.mission) },
      { label: '序号', value: safe(item.index) },
      { label: '超时', value: safe(item.timeout) },
      { label: '依赖', value: joinList(item.dependsOn) },
      { label: '任务数', value: safe(item.tasks?.length) },
    ]
  }
  if (focusStore.kind === 'flighttask') {
    return [
      { label: '任务', value: safe(item.mission) },
      { label: '阶段', value: safe(item.stage) },
      { label: '节点', value: safe(item.node) },
      { label: '武器', value: safe(item.weapon) },
      { label: '尝试次数', value: safe(item.attempts) },
      { label: '调度时间', value: safe(item.scheduledAt) },
    ]
  }
  if (focusStore.kind === 'weapon') {
    return [
      { label: '镜像', value: safe(item.image) },
      { label: '使用次数', value: safe(item.usage) },
      { label: '资源', value: safe(item.resources) },
      { label: '机型', value: joinList(item.aircraft) },
      { label: '挂载点', value: joinList(item.hardpoints) },
    ]
  }
  if (focusStore.kind === 'node') {
    const cpu = item.cpu === 0 ? '0%' : item.cpu ? `${item.cpu}%` : '--'
    const memory = item.memory === 0 ? '0%' : item.memory ? `${item.memory}%` : '--'
    return [
      { label: '角色', value: safe(item.role) },
      { label: '区域', value: safe(item.zone) },
      { label: 'Pod 数', value: safe(item.pods) },
      { label: 'CPU', value: cpu },
      { label: '内存', value: memory },
    ]
  }
  return []
})

const events = computed(() => focusStore.events)
</script>

<template>
  <aside class="detail-panel">
    <div class="section-title">焦点: {{ focusLabel }}</div>
    <div class="detail-card">
      <div class="detail-title">{{ focusTitle }}</div>
      <div class="detail-meta">
        <span
          v-for="badge in focusBadges"
          :key="badge.label"
          class="badge"
          :class="badge.tone"
        >
          {{ badge.label }}
        </span>
      </div>
      <div class="detail-info">
        <div v-for="row in focusInfo" :key="row.label" class="kv">
          <span>{{ row.label }}</span>
          <span>{{ row.value }}</span>
        </div>
      </div>
    </div>

    <div class="section-title">最近事件</div>
    <div class="event-list">
      <div v-for="event in events" :key="event.time + event.label" class="event-item">
        <span class="event-time">{{ event.time }}</span>
        <span class="event-dot" :class="event.tone"></span>
        <span class="event-text">{{ event.label }}</span>
      </div>
    </div>
  </aside>
</template>
