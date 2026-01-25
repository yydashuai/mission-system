<script setup>
import { computed, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useFocusStore } from '../stores/focus'
import { useDataStore } from '../stores/data'

const dataStore = useDataStore()
const { weapons } = storeToRefs(dataStore)
const isLoading = computed(() => dataStore.loading.weapons)

const selectedKey = ref('')
const query = ref('')
const statusFilter = ref('all')
const sortKey = ref('name')
const sortDir = ref('asc')
const focusStore = useFocusStore()

const emptyWeapon = {
  name: '--',
  status: '--',
  image: '--',
  version: '--',
  usage: '--',
  aircraft: [],
  hardpoints: [],
  resources: '--',
}

const statusOptions = [
  { value: 'all', label: 'All' },
  { value: 'Available', label: 'Available' },
  { value: 'Updating', label: 'Updating' },
  { value: 'Degraded', label: 'Degraded' },
  { value: 'Deprecated', label: 'Deprecated' },
  { value: 'Retired', label: 'Retired' },
]

const statusScore = (value) => {
  const key = String(value || '').toLowerCase()
  if (key === 'available') return 5
  if (key === 'updating') return 4
  if (key === 'degraded') return 3
  if (key === 'deprecated') return 2
  if (key === 'retired') return 1
  return 0
}

const usageScore = (value) => {
  if (!value) return 0
  const match = String(value).match(/\d+/)
  return match ? Number(match[0]) : 0
}

const weaponCounts = computed(() => {
  const counts = { all: weapons.value.length }
  weapons.value.forEach((weapon) => {
    const key = weapon.status || 'Unknown'
    counts[key] = (counts[key] || 0) + 1
  })
  return counts
})

const filteredWeapons = computed(() => {
  const text = String(query.value || '').trim().toLowerCase()
  let list = weapons.value

  if (text) {
    list = list.filter((weapon) => {
      const haystack = [weapon.name, weapon.image, weapon.version, weapon.resources]
        .join(' ')
        .toLowerCase()
      return haystack.includes(text)
    })
  }

  if (statusFilter.value !== 'all') {
    list = list.filter((weapon) => weapon.status === statusFilter.value)
  }

  const sorted = [...list].sort((a, b) => {
    const aValue = (() => {
      if (sortKey.value === 'status') return statusScore(a.status)
      if (sortKey.value === 'usage') return usageScore(a.usage)
      if (sortKey.value === 'version') return a.version
      return a.name
    })()

    const bValue = (() => {
      if (sortKey.value === 'status') return statusScore(b.status)
      if (sortKey.value === 'usage') return usageScore(b.usage)
      if (sortKey.value === 'version') return b.version
      return b.name
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
  filteredWeapons.value.find((item) => item.name === selectedKey.value) || null
))

const selectedSafe = computed(() => selected.value || emptyWeapon)

const toggleSort = () => {
  sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
}

watch(filteredWeapons, (list) => {
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
    focusStore.setFocus('weapon', value)
  }
}, { immediate: true })

const statusTone = (status) => {
  if (status === 'Available') return 'ok'
  if (status === 'Updating' || status === 'Degraded') return 'warn'
  if (status === 'Deprecated' || status === 'Retired') return 'muted'
  return 'muted'
}
</script>

<template>
  <section class="page">
    <header class="page-header">
      <div>
        <div class="page-title">Weapons</div>
        <div class="page-sub">Compatibility, container spec, and usage.</div>
      </div>
      <div class="page-actions">
        <button class="ghost small">Pull Image</button>
        <button class="ghost small">Compatibility Matrix</button>
      </div>
    </header>

    <section class="split-grid">
      <div class="panel">
      <div class="panel-header">
        <div>
          <div class="panel-title">Weapon Registry</div>
          <div class="panel-sub">Sidecar packages available to FlightTasks.</div>
        </div>
        <span v-if="isLoading" class="badge warn">Loading</span>
        <span v-else class="badge">{{ filteredWeapons.length }} / {{ weapons.length }} packages</span>
      </div>
      <div class="panel-toolbar">
        <div class="filter-row">
          <input
            v-model="query"
            class="input"
            type="search"
            placeholder="Search weapon, image, resources"
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
              <span class="count-pill">{{ weaponCounts[option.value] || 0 }}</span>
            </button>
          </div>
        </div>
        <div class="filter-row">
          <div class="filter-group">
            <span class="filter-label">Sort</span>
            <select v-model="sortKey" class="select">
              <option value="name">Name</option>
              <option value="status">Status</option>
              <option value="usage">Usage</option>
              <option value="version">Version</option>
            </select>
            <button type="button" class="ghost small" @click="toggleSort">
              {{ sortDir === 'asc' ? 'Asc' : 'Desc' }}
            </button>
          </div>
        </div>
      </div>
      <div class="data-table">
        <div class="data-row is-head" style="--cols: 1.1fr 0.8fr 0.9fr 0.7fr 0.7fr">
          <span>Weapon</span>
          <span>Status</span>
            <span>Image</span>
            <span>Version</span>
            <span>Usage</span>
          </div>
          <button
            v-for="weapon in filteredWeapons"
            :key="weapon.name"
            class="data-row is-selectable"
            :class="{ active: selectedKey === weapon.name }"
            style="--cols: 1.1fr 0.8fr 0.9fr 0.7fr 0.7fr"
            type="button"
            @click="selectedKey = weapon.name"
          >
            <div class="cell-main">
              <div class="cell-title">{{ weapon.name }}</div>
              <div class="cell-sub">{{ weapon.resources }}</div>
            </div>
            <span class="badge" :class="statusTone(weapon.status)">{{ weapon.status }}</span>
            <span class="muted">{{ weapon.image }}</span>
            <span class="muted">{{ weapon.version }}</span>
            <span class="badge muted">{{ weapon.usage }}</span>
          </button>
          <div v-if="!filteredWeapons.length && !isLoading" class="empty-state">
            No weapons match the current filters.
          </div>
        </div>
      </div>

      <div class="panel">
        <div class="panel-header">
          <div>
            <div class="panel-title">Weapon Detail</div>
            <div class="panel-sub">{{ selectedSafe.image }}</div>
          </div>
          <span class="badge" :class="statusTone(selectedSafe.status)">{{ selectedSafe.status }}</span>
        </div>

        <div class="detail-card">
          <div class="detail-title">{{ selectedSafe.name }}</div>
          <div class="detail-meta">
            <span class="badge muted">Version {{ selectedSafe.version }}</span>
            <span class="badge">{{ selectedSafe.usage }}</span>
          </div>
          <div class="detail-info">
            <div class="kv">
              <span>Resources</span>
              <span>{{ selectedSafe.resources }}</span>
            </div>
            <div class="kv">
              <span>Image</span>
              <span>{{ selectedSafe.image }}</span>
            </div>
          </div>
        </div>

        <div class="panel">
          <div class="panel-header">
            <div>
              <div class="panel-title">Compatibility</div>
              <div class="panel-sub">Verified aircraft and hardpoints.</div>
            </div>
          </div>
          <div class="chip-row">
            <span v-for="item in selectedSafe.aircraft" :key="item" class="chip">{{ item }}</span>
            <div v-if="!selectedSafe.aircraft.length" class="empty-state">No aircraft listed.</div>
          </div>
          <div class="chip-row">
            <span v-for="item in selectedSafe.hardpoints" :key="item" class="chip">{{ item }}</span>
            <div v-if="!selectedSafe.hardpoints.length" class="empty-state">No hardpoints listed.</div>
          </div>
        </div>
      </div>
    </section>
  </section>
</template>
