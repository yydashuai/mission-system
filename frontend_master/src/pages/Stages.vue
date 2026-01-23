<script setup>
import { computed, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useFocusStore } from '../stores/focus'
import { useDataStore } from '../stores/data'

const dataStore = useDataStore()
const { stages } = storeToRefs(dataStore)
const isLoading = computed(() => dataStore.loading.stages)

const selectedKey = ref('')
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

const selected = computed(() => (
  stages.value.find((item) => buildKey(item) === selectedKey.value) || null
))

const selectedSafe = computed(() => selected.value || emptyStage)

watch(stages, (list) => {
  if (!list.length) return
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
          <span v-else class="badge">{{ stages.length }} stages</span>
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
            v-for="stage in stages"
            :key="stage.name + stage.mission"
            class="data-row is-selectable"
            :class="{ active: selectedKey === buildKey(stage) }"
            style="--cols: 1.2fr 0.9fr 0.7fr 0.7fr 0.6fr"
            type="button"
            @click="selectedKey = buildKey(stage)"
          >
            <div class="cell-main">
              <div class="cell-title">{{ stage.name }}</div>
              <div class="cell-sub">Timeout {{ stage.timeout }}</div>
            </div>
            <span class="muted">{{ stage.mission }}</span>
            <span class="badge" :class="stage.status.toLowerCase()">{{ stage.status }}</span>
            <span class="badge muted">{{ stage.mode }}</span>
            <span class="muted">{{ stage.index }}</span>
          </button>
          <div v-if="!stages.length && !isLoading" class="empty-state">No stages available.</div>
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
              <div class="panel-title">FlightTasks</div>
              <div class="panel-sub">Execution order and current node.</div>
            </div>
            <span class="badge">{{ selectedSafe.tasks.length }} tasks</span>
          </div>
          <div class="task-table">
            <div class="task-row task-head">
              <span>Name</span>
              <span>Status</span>
              <span>ETA</span>
              <span>Node</span>
            </div>
            <div v-for="task in selectedSafe.tasks" :key="task.name" class="task-row">
              <span class="task-name">{{ task.name }}</span>
              <span class="badge" :class="task.status.toLowerCase()">{{ task.status }}</span>
              <span>{{ task.eta }}</span>
              <span class="muted">{{ task.node }}</span>
            </div>
            <div v-if="!selectedSafe.tasks.length" class="empty-state">No tasks available.</div>
          </div>
        </div>
      </div>
    </section>
  </section>
</template>
