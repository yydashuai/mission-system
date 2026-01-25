<script setup>
import { computed, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useFocusStore } from '../stores/focus'
import { useDataStore } from '../stores/data'

const dataStore = useDataStore()
const { stages } = storeToRefs(dataStore)
const isLoading = computed(() => dataStore.loading.stages)

const selectedKey = ref('')
const query = ref('')
const statusFilter = ref('all')
const missionFilter = ref('all')
const modeFilter = ref('all')
const sortKey = ref('index')
const sortDir = ref('asc')
const focusStore = useFocusStore()

const emptyStage = {
  name: '--',
  mission: '--',
  index: '--',
  mode: '--',
  status: '--',
  timeout: '--',
  dependsOn: [],
  tasks: [],
}

const buildKey = (item) => `${item.mission}::${item.name}`

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

const stageIndexScore = (value) => {
  const index = Number(value)
  return Number.isNaN(index) ? 0 : index
}

const stageCounts = computed(() => {
  const counts = { all: stages.value.length }
  stages.value.forEach((stage) => {
    const key = stage.status || 'Unknown'
    counts[key] = (counts[key] || 0) + 1
  })
  return counts
})

const missionOptions = computed(() => {
  const set = new Set()
  stages.value.forEach((stage) => {
    if (stage.mission && stage.mission !== '--') set.add(stage.mission)
  })
  return ['all', ...Array.from(set).sort((a, b) => a.localeCompare(b))]
})

const modeOptions = computed(() => {
  const set = new Set()
  stages.value.forEach((stage) => {
    if (stage.mode && stage.mode !== '--') set.add(stage.mode)
  })
  return ['all', ...Array.from(set)]
})

const filteredStages = computed(() => {
  const text = String(query.value || '').trim().toLowerCase()
  let list = stages.value

  if (text) {
    list = list.filter((stage) => {
      const haystack = [stage.name, stage.mission].join(' ').toLowerCase()
      return haystack.includes(text)
    })
  }

  if (statusFilter.value !== 'all') {
    list = list.filter((stage) => stage.status === statusFilter.value)
  }

  if (missionFilter.value !== 'all') {
    list = list.filter((stage) => stage.mission === missionFilter.value)
  }

  if (modeFilter.value !== 'all') {
    list = list.filter((stage) => stage.mode === modeFilter.value)
  }

  const sorted = [...list].sort((a, b) => {
    const aValue = (() => {
      if (sortKey.value === 'name') return a.name
      if (sortKey.value === 'mission') return a.mission
      if (sortKey.value === 'status') return statusScore(a.status)
      if (sortKey.value === 'mode') return a.mode
      return stageIndexScore(a.index)
    })()

    const bValue = (() => {
      if (sortKey.value === 'name') return b.name
      if (sortKey.value === 'mission') return b.mission
      if (sortKey.value === 'status') return statusScore(b.status)
      if (sortKey.value === 'mode') return b.mode
      return stageIndexScore(b.index)
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
  filteredStages.value.find((item) => buildKey(item) === selectedKey.value) || null
))

const selectedSafe = computed(() => selected.value || emptyStage)

const stageSubLabel = (stage) => {
  const parts = [`Timeout ${stage.timeout}`]
  if (stage.dependsOn && stage.dependsOn.length) {
    parts.push(`Depends ${stage.dependsOn.join(', ')}`)
  }
  return parts.join(' · ')
}

const toggleSort = () => {
  sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
}

watch(filteredStages, (list) => {
  if (!list.length) {
    selectedKey.value = ''
    return
  }
  const exists = list.some((item) => buildKey(item) === selectedKey.value)
  if (!exists) {
    selectedKey.value = buildKey(list[0])
  }
}, { immediate: true })

watch(selected, (value) => {
  if (value) {
    focusStore.setFocus('stage', value)
  }
}, { immediate: true })
</script>

<template>
  <section class="page">
    <header class="page-header">
      <div>
        <div class="page-title">Mission Stages</div>
        <div class="page-sub">Stage execution flow and dependencies.</div>
      </div>
      <div class="page-actions">
        <button class="ghost small">View Gantt</button>
        <button class="ghost small">Export YAML</button>
      </div>
    </header>

    <section class="split-grid">
      <div class="panel">
      <div class="panel-header">
        <div>
          <div class="panel-title">Stages</div>
          <div class="panel-sub">Sorted by mission and sequence.</div>
        </div>
        <span v-if="isLoading" class="badge warn">Loading</span>
        <span v-else class="badge">{{ filteredStages.length }} / {{ stages.length }} stages</span>
      </div>
      <div class="panel-toolbar">
        <div class="filter-row">
          <input
            v-model="query"
            class="input"
            type="search"
            placeholder="Search stage or mission"
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
              <span class="count-pill">{{ stageCounts[option.value] || 0 }}</span>
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
            <span class="filter-label">Mode</span>
            <select v-model="modeFilter" class="select">
              <option v-for="mode in modeOptions" :key="mode" :value="mode">
                {{ mode === 'all' ? 'All modes' : mode }}
              </option>
            </select>
          </div>
          <div class="filter-group">
            <span class="filter-label">Sort</span>
            <select v-model="sortKey" class="select">
              <option value="index">Index</option>
              <option value="name">Name</option>
              <option value="mission">Mission</option>
              <option value="status">Status</option>
              <option value="mode">Mode</option>
            </select>
            <button type="button" class="ghost small" @click="toggleSort">
              {{ sortDir === 'asc' ? 'Asc' : 'Desc' }}
            </button>
          </div>
        </div>
      </div>
      <div class="data-table">
        <div class="data-row is-head" style="--cols: 1.2fr 0.9fr 0.7fr 0.7fr 0.6fr">
          <span>Stage</span>
          <span>Mission</span>
            <span>Status</span>
            <span>Mode</span>
            <span>Index</span>
          </div>
          <button
            v-for="stage in filteredStages"
            :key="stage.name + stage.mission"
            class="data-row is-selectable"
            :class="{ active: selectedKey === buildKey(stage) }"
            style="--cols: 1.2fr 0.9fr 0.7fr 0.7fr 0.6fr"
            type="button"
            @click="selectedKey = buildKey(stage)"
          >
            <div class="cell-main">
              <div class="cell-title">{{ stage.name }}</div>
              <div class="cell-sub">{{ stageSubLabel(stage) }}</div>
            </div>
            <span class="muted">{{ stage.mission }}</span>
            <span class="badge" :class="stage.status.toLowerCase()">{{ stage.status }}</span>
            <span class="badge muted">{{ stage.mode }}</span>
            <span class="muted">{{ stage.index }}</span>
          </button>
          <div v-if="!filteredStages.length && !isLoading" class="empty-state">
            No stages match the current filters.
          </div>
        </div>
      </div>

      <div class="panel">
        <div class="panel-header">
          <div>
            <div class="panel-title">Stage Detail</div>
            <div class="panel-sub">{{ selectedSafe.mission }}</div>
          </div>
          <span class="badge" :class="String(selectedSafe.status).toLowerCase()">{{ selectedSafe.status }}</span>
        </div>

        <div class="detail-card">
          <div class="detail-title">{{ selectedSafe.name }}</div>
          <div class="detail-meta">
            <span class="badge muted">{{ selectedSafe.mode }}</span>
            <span class="badge">Timeout {{ selectedSafe.timeout }}</span>
          </div>
          <div class="detail-info">
            <div class="kv">
              <span>Sequence</span>
              <span>Stage {{ selectedSafe.index }}</span>
            </div>
            <div class="kv">
              <span>Depends On</span>
              <span>{{ selectedSafe.dependsOn.length ? selectedSafe.dependsOn.join(', ') : 'None' }}</span>
            </div>
            <div class="kv">
              <span>FlightTasks</span>
              <span>{{ selectedSafe.tasks.length }}</span>
            </div>
          </div>
        </div>

        <div class="panel">
          <div class="panel-header">
            <div>
              <div class="panel-title">Dependencies</div>
              <div class="panel-sub">Upstream stages gating execution.</div>
            </div>
          </div>
          <div class="chip-row">
            <span v-for="item in selectedSafe.dependsOn" :key="item" class="chip">{{ item }}</span>
            <div v-if="!selectedSafe.dependsOn.length" class="empty-state">No dependencies.</div>
          </div>
        </div>

        <div class="panel">
          <div class="panel-header">
            <div>
              <div class="panel-title">FlightTasks</div>
              <div class="panel-sub">Execution order and current node.</div>
            </div>
            <span class="badge">{{ selectedSafe.tasks.length }} tasks</span>
          </div>
          <div class="task-table">
          <div class="task-row task-head">
            <div class="cell-start">Name</div>
            <div class="cell-center">Status</div>
            <div class="cell-center">ETA</div>
            <div class="cell-center">Node</div>
          </div>
          <div v-for="task in selectedSafe.tasks" :key="task.name" class="task-row">
            <div class="cell-start">
              <span class="task-name">{{ task.name }}</span>
            </div>
            <div class="cell-center">
              <span class="badge" :class="task.status.toLowerCase()">{{ task.status }}</span>
            </div>
            <div class="cell-center">
              <span>{{ task.eta }}</span>
            </div>
            <div class="cell-center">
              <span class="muted">{{ task.node }}</span>
            </div>
          </div>
            <div v-if="!selectedSafe.tasks.length" class="empty-state">No tasks available.</div>
          </div>
        </div>
      </div>
    </section>
  </section>
</template>
