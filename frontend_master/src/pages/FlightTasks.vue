<script setup>
import { computed, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useFocusStore } from '../stores/focus'
import { useDataStore } from '../stores/data'

const dataStore = useDataStore()
const { flightTasks } = storeToRefs(dataStore)
const isLoading = computed(() => dataStore.loading.flightTasks)

const selectedKey = ref('')
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

const selected = computed(() => (
  flightTasks.value.find((item) => item.name === selectedKey.value) || null
))

const selectedSafe = computed(() => selected.value || emptyTask)

watch(flightTasks, (list) => {
  if (!list.length) return
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
          <span v-else class="badge">{{ flightTasks.length }} tasks</span>
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
            v-for="task in flightTasks"
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
          <div v-if="!flightTasks.length && !isLoading" class="empty-state">No tasks available.</div>
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
