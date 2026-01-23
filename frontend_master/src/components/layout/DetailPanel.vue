<script setup>
import { computed } from 'vue'
import { useFocusStore } from '../../stores/focus'

const focusStore = useFocusStore()

const focusLabel = computed(() => {
  const labels = {
    mission: 'Mission',
    stage: 'MissionStage',
    flighttask: 'FlightTask',
    weapon: 'Weapon',
    node: 'Node',
  }
  return labels[focusStore.kind] || 'Focus'
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
      { label: `${safe(item.priority)} Priority`, tone: priorityTone(item.priority) },
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
      { label: 'Type', value: safe(item.type) },
      { label: 'Commander', value: safe(item.commander) },
      { label: 'Region', value: safe(item.region) },
      { label: 'Tasks', value: safe(item.tasks) },
      { label: 'Updated', value: safe(item.updated) },
    ]
  }
  if (focusStore.kind === 'stage') {
    return [
      { label: 'Mission', value: safe(item.mission) },
      { label: 'Index', value: safe(item.index) },
      { label: 'Timeout', value: safe(item.timeout) },
      { label: 'Depends On', value: joinList(item.dependsOn) },
      { label: 'Tasks', value: safe(item.tasks?.length) },
    ]
  }
  if (focusStore.kind === 'flighttask') {
    return [
      { label: 'Mission', value: safe(item.mission) },
      { label: 'Stage', value: safe(item.stage) },
      { label: 'Node', value: safe(item.node) },
      { label: 'Weapon', value: safe(item.weapon) },
      { label: 'Attempts', value: safe(item.attempts) },
      { label: 'Scheduled', value: safe(item.scheduledAt) },
    ]
  }
  if (focusStore.kind === 'weapon') {
    return [
      { label: 'Image', value: safe(item.image) },
      { label: 'Usage', value: safe(item.usage) },
      { label: 'Resources', value: safe(item.resources) },
      { label: 'Aircraft', value: joinList(item.aircraft) },
      { label: 'Hardpoints', value: joinList(item.hardpoints) },
    ]
  }
  if (focusStore.kind === 'node') {
    const cpu = item.cpu === 0 ? '0%' : item.cpu ? `${item.cpu}%` : '--'
    const memory = item.memory === 0 ? '0%' : item.memory ? `${item.memory}%` : '--'
    return [
      { label: 'Role', value: safe(item.role) },
      { label: 'Zone', value: safe(item.zone) },
      { label: 'Pods', value: safe(item.pods) },
      { label: 'CPU', value: cpu },
      { label: 'Memory', value: memory },
    ]
  }
  return []
})

const events = computed(() => focusStore.events)
</script>

<template>
  <aside class="detail-panel">
    <div class="section-title">Focus: {{ focusLabel }}</div>
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

    <div class="section-title">Recent Events</div>
    <div class="event-list">
      <div v-for="event in events" :key="event.time + event.label" class="event-item">
        <span class="event-time">{{ event.time }}</span>
        <span class="event-dot" :class="event.tone"></span>
        <span class="event-text">{{ event.label }}</span>
      </div>
    </div>
  </aside>
</template>
