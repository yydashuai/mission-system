<script setup>
import { computed, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useFocusStore } from '../stores/focus'
import { useDataStore } from '../stores/data'

const dataStore = useDataStore()
const { flightTasks } = storeToRefs(dataStore)
const isLoading = computed(() => dataStore.loading.flightTasks)

const selectedKey = ref('')
const query = ref('')
const statusFilter = ref('all')
const missionFilter = ref('all')
const stageFilter = ref('all')
const sortKey = ref('scheduledAt')
const sortDir = ref('desc')
const focusStore = useFocusStore()

const emptyTask = {
  name: '--',
  stage: '--',
  mission: '--',
  status: '--',
  pod: '--',
  node: '--',
  weapon: '--',
  attempts: '--',
  scheduledAt: '--',
  conditions: [],
  constraints: [],
  podStatus: '--',
  sidecars: [],
}

const statusOptions = [
  { value: 'all', label: 'All' },
  { value: 'Running', label: 'Running' },
  { value: 'Scheduled', label: 'Scheduled' },
  { value: 'Pending', label: 'Pending' },
  { value: 'Succeeded', label: 'Succeeded' },
  { value: 'Failed', label: 'Failed' },
]

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

const taskCounts = computed(() => {
  const counts = { all: flightTasks.value.length }
  flightTasks.value.forEach((task) => {
    const key = task.status || 'Unknown'
    counts[key] = (counts[key] || 0) + 1
  })
  return counts
})

const missionOptions = computed(() => {
  const set = new Set()
  flightTasks.value.forEach((task) => {
    if (task.mission && task.mission !== '--') set.add(task.mission)
  })
  return ['all', ...Array.from(set).sort((a, b) => a.localeCompare(b))]
})

const stageOptions = computed(() => {
  const set = new Set()
  flightTasks.value.forEach((task) => {
    if (task.stage && task.stage !== '--') set.add(task.stage)
  })
  return ['all', ...Array.from(set).sort((a, b) => a.localeCompare(b))]
})

const filteredTasks = computed(() => {
  const text = String(query.value || '').trim().toLowerCase()
  let list = flightTasks.value

  if (text) {
    list = list.filter((task) => {
      const haystack = [task.name, task.mission, task.stage, task.weapon, task.node]
        .join(' ')
        .toLowerCase()
      return haystack.includes(text)
    })
  }

  if (statusFilter.value !== 'all') {
    list = list.filter((task) => task.status === statusFilter.value)
  }

  if (missionFilter.value !== 'all') {
    list = list.filter((task) => task.mission === missionFilter.value)
  }

  if (stageFilter.value !== 'all') {
    list = list.filter((task) => task.stage === stageFilter.value)
  }

  const sorted = [...list].sort((a, b) => {
    const aValue = (() => {
      if (sortKey.value === 'name') return a.name
      if (sortKey.value === 'status') return statusScore(a.status)
      if (sortKey.value === 'attempts') return a.attempts || 0
      if (sortKey.value === 'node') return a.node
      return timeScore(a.scheduledAt)
    })()

    const bValue = (() => {
      if (sortKey.value === 'name') return b.name
      if (sortKey.value === 'status') return statusScore(b.status)
      if (sortKey.value === 'attempts') return b.attempts || 0
      if (sortKey.value === 'node') return b.node
      return timeScore(b.scheduledAt)
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
  filteredTasks.value.find((item) => item.name === selectedKey.value) || null
))

const selectedSafe = computed(() => selected.value || emptyTask)

const toggleSort = () => {
  sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
}

watch(filteredTasks, (list) => {
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
    focusStore.setFocus('flighttask', value)
  }
}, { immediate: true })
</script>

<template>
  <section class="page">
    <header class="page-header">
      <div>
        <div class="page-title">FlightTasks</div>
        <div class="page-sub">Scheduling detail, pod binding, and status.</div>
      </div>
      <div class="page-actions">
        <button class="ghost small">Copy YAML</button>
        <button class="ghost small">Inspect Pod</button>
      </div>
    </header>

    <section class="split-grid">
      <div class="panel">
      <div class="panel-header">
        <div>
          <div class="panel-title">Task Queue</div>
          <div class="panel-sub">Select a task for scheduling context.</div>
        </div>
        <span v-if="isLoading" class="badge warn">Loading</span>
        <span v-else class="badge">{{ filteredTasks.length }} / {{ flightTasks.length }} tasks</span>
      </div>
      <div class="panel-toolbar">
        <div class="filter-row">
          <input
            v-model="query"
            class="input"
            type="search"
            placeholder="Search task, mission, weapon, node"
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
              <span class="count-pill">{{ taskCounts[option.value] || 0 }}</span>
            </button>
          </div>
        </div>
        <div class="filter-row">
          <div class="filter-group">
            <span class="filter-label">Mission</span>
            <select v-model="missionFilter" class="select">
              <option v-for="mission in missionOptions" :key="mission" :value="mission">
                {{ mission === 'all' ? 'All missions' : mission }}
              </option>
            </select>
          </div>
          <div class="filter-group">
            <span class="filter-label">Stage</span>
            <select v-model="stageFilter" class="select">
              <option v-for="stage in stageOptions" :key="stage" :value="stage">
                {{ stage === 'all' ? 'All stages' : stage }}
              </option>
            </select>
          </div>
          <div class="filter-group">
            <span class="filter-label">Sort</span>
            <select v-model="sortKey" class="select">
              <option value="scheduledAt">Scheduled</option>
              <option value="name">Name</option>
              <option value="status">Status</option>
              <option value="attempts">Attempts</option>
              <option value="node">Node</option>
            </select>
            <button type="button" class="ghost small" @click="toggleSort">
              {{ sortDir === 'asc' ? 'Asc' : 'Desc' }}
            </button>
          </div>
        </div>
      </div>
      <div class="data-table">
        <div class="data-row is-head" style="--cols: 1.2fr 0.8fr 0.7fr 0.7fr 0.6fr">
          <span>Task</span>
          <span>Stage</span>
            <span>Status</span>
            <span>Node</span>
            <span>Weapon</span>
          </div>
          <button
            v-for="task in filteredTasks"
            :key="task.name"
            class="data-row is-selectable"
            :class="{ active: selectedKey === task.name }"
            style="--cols: 1.2fr 0.8fr 0.7fr 0.7fr 0.6fr"
            type="button"
            @click="selectedKey = task.name"
          >
            <div class="cell-main">
              <div class="cell-title">{{ task.name }}</div>
              <div class="cell-sub">{{ task.mission }}</div>
            </div>
            <span class="muted">{{ task.stage }}</span>
            <span class="badge" :class="task.status.toLowerCase()">{{ task.status }}</span>
            <span class="muted">{{ task.node }}</span>
            <span class="badge muted">{{ task.weapon }}</span>
          </button>
          <div v-if="!filteredTasks.length && !isLoading" class="empty-state">
            No tasks match the current filters.
          </div>
        </div>
      </div>

      <div class="panel">
        <div class="panel-header">
          <div>
            <div class="panel-title">Task Detail</div>
            <div class="panel-sub">{{ selectedSafe.mission }} / {{ selectedSafe.stage }}</div>
          </div>
          <span class="badge" :class="String(selectedSafe.status).toLowerCase()">{{ selectedSafe.status }}</span>
        </div>

        <div class="detail-card">
          <div class="detail-title">{{ selectedSafe.name }}</div>
          <div class="detail-meta">
            <span class="badge muted">Pod {{ selectedSafe.pod }}</span>
            <span class="badge">Node {{ selectedSafe.node }}</span>
          </div>
          <div class="detail-info">
            <div class="kv">
              <span>Scheduled At</span>
              <span>{{ selectedSafe.scheduledAt }}</span>
            </div>
            <div class="kv">
              <span>Scheduling Attempts</span>
              <span>{{ selectedSafe.attempts }}</span>
            </div>
            <div class="kv">
              <span>Pod Status</span>
              <span>{{ selectedSafe.podStatus }}</span>
            </div>
            <div class="kv">
              <span>Sidecars</span>
              <span>{{ selectedSafe.sidecars.join(', ') }}</span>
            </div>
          </div>
        </div>

        <div class="panel">
          <div class="panel-header">
            <div>
              <div class="panel-title">Scheduling Constraints</div>
              <div class="panel-sub">Required node affinity and aircraft filters.</div>
            </div>
          </div>
          <div class="chip-row">
            <span v-for="item in selectedSafe.constraints" :key="item" class="chip">{{ item }}</span>
            <div v-if="!selectedSafe.constraints.length" class="empty-state">No constraints data.</div>
          </div>
        </div>

        <div class="panel">
          <div class="panel-header">
            <div>
              <div class="panel-title">Conditions</div>
              <div class="panel-sub">Live pod + scheduler signals.</div>
            </div>
          </div>
          <div class="condition-list">
            <div v-for="item in selectedSafe.conditions" :key="item.label" class="condition-item">
              <span class="badge" :class="item.tone">{{ item.label }}</span>
              <span class="condition-detail">{{ item.detail }}</span>
            </div>
            <div v-if="!selectedSafe.conditions.length" class="empty-state">No condition updates.</div>
          </div>
        </div>
      </div>
    </section>
  </section>
</template>
