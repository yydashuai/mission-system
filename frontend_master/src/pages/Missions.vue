<script setup>
import { computed, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useFocusStore } from '../stores/focus'
import { useDataStore } from '../stores/data'

const dataStore = useDataStore()
const { missions } = storeToRefs(dataStore)
const isLoading = computed(() => dataStore.loading.missions)

const selectedKey = ref('')
const query = ref('')
const statusFilter = ref('all')
const sortKey = ref('updated')
const sortDir = ref('desc')
const focusStore = useFocusStore()

const emptyMission = {
  name: '--',
  type: '--',
  priority: '--',
  status: '--',
  commander: '--',
  region: '--',
  updated: '--',
  objective: '--',
  failurePolicy: '--',
  tasks: '--',
  stages: [],
}

const statusOptions = [
  { value: 'all', label: 'All' },
  { value: 'Running', label: 'Running' },
  { value: 'Scheduled', label: 'Scheduled' },
  { value: 'Pending', label: 'Pending' },
  { value: 'Succeeded', label: 'Succeeded' },
  { value: 'Failed', label: 'Failed' },
]

const priorityScore = (value) => {
  const key = String(value || '').toLowerCase()
  if (key === 'critical') return 4
  if (key === 'high') return 3
  if (key === 'normal') return 2
  if (key === 'low') return 1
  return 0
}

const statusScore = (value) => {
  const key = String(value || '').toLowerCase()
  if (key === 'succeeded') return 5
  if (key === 'running') return 4
  if (key === 'scheduled') return 3
  if (key === 'pending') return 2
  if (key === 'failed') return 1
  return 0
}

const timeScore = (value) => {
  if (!value || value === '--') return 0
  const parts = String(value).split(':')
  if (parts.length === 2) {
    const hours = Number(parts[0])
    const minutes = Number(parts[1])
    if (!Number.isNaN(hours) && !Number.isNaN(minutes)) {
      return hours * 60 + minutes
    }
  }
  const date = new Date(value)
  if (!Number.isNaN(date.getTime())) return date.getTime()
  return 0
}

const missionCounts = computed(() => {
  const counts = { all: missions.value.length }
  missions.value.forEach((mission) => {
    const key = mission.status || 'Unknown'
    counts[key] = (counts[key] || 0) + 1
  })
  return counts
})

const filteredMissions = computed(() => {
  const text = String(query.value || '').trim().toLowerCase()
  const statusKey = statusFilter.value
  let list = missions.value

  if (text) {
    list = list.filter((mission) => {
      const haystack = [
        mission.name,
        mission.type,
        mission.commander,
        mission.region,
        mission.objective,
      ].join(' ').toLowerCase()
      return haystack.includes(text)
    })
  }

  if (statusKey !== 'all') {
    list = list.filter((mission) => mission.status === statusKey)
  }

  const sorted = [...list].sort((a, b) => {
    const aValue = (() => {
      if (sortKey.value === 'name') return a.name
      if (sortKey.value === 'priority') return priorityScore(a.priority)
      if (sortKey.value === 'status') return statusScore(a.status)
      if (sortKey.value === 'tasks') return a.tasks || 0
      if (sortKey.value === 'stages') return a.stages.length
      return timeScore(a.updated)
    })()

    const bValue = (() => {
      if (sortKey.value === 'name') return b.name
      if (sortKey.value === 'priority') return priorityScore(b.priority)
      if (sortKey.value === 'status') return statusScore(b.status)
      if (sortKey.value === 'tasks') return b.tasks || 0
      if (sortKey.value === 'stages') return b.stages.length
      return timeScore(b.updated)
    })()

    let result = 0
    if (typeof aValue === 'number' && typeof bValue === 'number') {
      result = aValue - bValue
    } else {
      result = String(aValue).localeCompare(String(bValue))
    }
    return sortDir.value === 'asc' ? result : -result
  })

  return sorted
})

const selected = computed(() => (
  filteredMissions.value.find((item) => item.name === selectedKey.value) || null
))

const selectedSafe = computed(() => selected.value || emptyMission)

const toggleSort = () => {
  sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
}

watch(filteredMissions, (list) => {
  if (!list.length) {
    selectedKey.value = ''
    return
  }
  const exists = list.some((item) => item.name === selectedKey.value)
  if (!exists) {
    selectedKey.value = list[0].name
  }
}, { immediate: true })

watch(selected, (value) => {
  if (value) {
    focusStore.setFocus('mission', value)
  }
}, { immediate: true })

const priorityTone = (priority) => {
  const key = String(priority || '').toLowerCase()
  if (key === 'high' || key === 'critical') return 'warn'
  if (key === 'low') return 'muted'
  return 'ok'
}
</script>

<template>
  <section class="page">
    <header class="page-header">
      <div>
        <div class="page-title">Missions</div>
        <div class="page-sub">Mission list, lifecycle, and stage composition.</div>
      </div>
      <div class="page-actions">
        <button class="ghost small">Export JSON</button>
        <button class="primary">Launch Demo</button>
      </div>
    </header>

    <section class="split-grid">
      <div class="panel">
      <div class="panel-header">
        <div>
          <div class="panel-title">Active Missions</div>
          <div class="panel-sub">Select a mission to inspect stage flow.</div>
        </div>
        <span v-if="isLoading" class="badge warn">Loading</span>
        <span v-else class="badge">{{ filteredMissions.length }} / {{ missions.length }} tracked</span>
      </div>

      <div class="panel-toolbar">
        <div class="filter-row">
          <input
            v-model="query"
            class="input"
            type="search"
            placeholder="Search mission, commander, region"
          />
          <div class="segmented">
            <button
              v-for="option in statusOptions"
              :key="option.value"
              type="button"
              class="segmented-btn"
              :class="{ active: statusFilter === option.value }"
              @click="statusFilter = option.value"
            >
              <span>{{ option.label }}</span>
              <span class="count-pill">{{ missionCounts[option.value] || 0 }}</span>
            </button>
          </div>
        </div>
        <div class="filter-row">
          <div class="filter-group">
            <span class="filter-label">Sort</span>
            <select v-model="sortKey" class="select">
              <option value="updated">Updated</option>
              <option value="name">Name</option>
              <option value="priority">Priority</option>
              <option value="status">Status</option>
              <option value="tasks">FlightTasks</option>
              <option value="stages">Stages</option>
            </select>
            <button type="button" class="ghost small" @click="toggleSort">
              {{ sortDir === 'asc' ? 'Asc' : 'Desc' }}
            </button>
          </div>
        </div>
      </div>

      <div class="data-table">
        <div class="data-row is-head" style="--cols: 1.3fr 0.7fr 0.7fr 0.6fr 0.7fr">
          <span>Mission</span>
          <span>Status</span>
            <span>Priority</span>
            <span>Stages</span>
            <span>Updated</span>
          </div>
          <button
            v-for="mission in filteredMissions"
            :key="mission.name"
            class="data-row is-selectable"
            :class="{ active: selectedKey === mission.name }"
            style="--cols: 1.3fr 0.7fr 0.7fr 0.6fr 0.7fr"
            type="button"
            @click="selectedKey = mission.name"
          >
            <div class="cell-main">
              <div class="cell-title">{{ mission.name }}</div>
              <div class="cell-sub">{{ mission.type }}</div>
            </div>
            <span class="badge" :class="mission.status.toLowerCase()">{{ mission.status }}</span>
            <span class="badge" :class="priorityTone(mission.priority)">{{ mission.priority }}</span>
            <span>{{ mission.stages.length }}</span>
            <span class="muted">{{ mission.updated }}</span>
          </button>
          <div v-if="!filteredMissions.length && !isLoading" class="empty-state">
            No missions match the current filters.
          </div>
        </div>
      </div>

      <div class="panel">
        <div class="panel-header">
          <div>
            <div class="panel-title">Mission Detail</div>
            <div class="panel-sub">{{ selectedSafe.region }}</div>
          </div>
          <span class="badge" :class="String(selectedSafe.status).toLowerCase()">{{ selectedSafe.status }}</span>
        </div>

        <div class="detail-card">
          <div class="detail-title">{{ selectedSafe.name }}</div>
          <div class="detail-meta">
            <span class="badge" :class="priorityTone(selectedSafe.priority)">{{ selectedSafe.priority }} Priority</span>
            <span class="badge muted">{{ selectedSafe.type }}</span>
            <span class="badge">Failure: {{ selectedSafe.failurePolicy }}</span>
          </div>
          <div class="detail-info">
            <div class="kv">
              <span>Commander</span>
              <span>{{ selectedSafe.commander }}</span>
            </div>
            <div class="kv">
              <span>Region</span>
              <span>{{ selectedSafe.region }}</span>
            </div>
            <div class="kv">
              <span>FlightTasks</span>
              <span>{{ selectedSafe.tasks }}</span>
            </div>
            <div class="kv">
              <span>Objective</span>
              <span>{{ selectedSafe.objective }}</span>
            </div>
          </div>
        </div>

        <div class="panel">
          <div class="panel-header">
            <div>
              <div class="panel-title">Stage Flow</div>
              <div class="panel-sub">Parallel and sequential segments.</div>
            </div>
            <span class="badge">{{ selectedSafe.stages.length }} stages</span>
          </div>
          <div class="flow-line">
          <div
            v-for="stage in selectedSafe.stages"
            :key="stage.name"
            class="flow-node"
            :class="stage.status.toLowerCase()"
          >
            <div class="flow-title">{{ stage.name }}</div>
            <div class="flow-meta">{{ stage.mode }} · {{ stage.tasks }} tasks</div>
            <div v-if="stage.dependsOn && stage.dependsOn.length" class="flow-deps">
              Depends: {{ stage.dependsOn.join(', ') }}
            </div>
            <span class="badge" :class="stage.status.toLowerCase()">{{ stage.status }}</span>
          </div>
          <div v-if="!selectedSafe.stages.length" class="empty-state">No stages available.</div>
        </div>
        </div>
      </div>
    </section>
  </section>
</template>
