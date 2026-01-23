<script setup>
import { computed, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useFocusStore } from '../stores/focus'
import { useDataStore } from '../stores/data'

const dataStore = useDataStore()
const { missions } = storeToRefs(dataStore)
const isLoading = computed(() => dataStore.loading.missions)

const selectedKey = ref('')
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

const selected = computed(() => (
  missions.value.find((item) => item.name === selectedKey.value) || null
))

const selectedSafe = computed(() => selected.value || emptyMission)

watch(missions, (list) => {
  if (!list.length) return
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
          <span v-else class="badge">{{ missions.length }} tracked</span>
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
            v-for="mission in missions"
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
          <div v-if="!missions.length && !isLoading" class="empty-state">No missions available.</div>
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
              <span class="badge" :class="stage.status.toLowerCase()">{{ stage.status }}</span>
            </div>
            <div v-if="!selectedSafe.stages.length" class="empty-state">No stages available.</div>
          </div>
        </div>
      </div>
    </section>
  </section>
</template>
