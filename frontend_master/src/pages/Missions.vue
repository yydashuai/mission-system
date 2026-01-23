<script setup>
import { ref, watch } from 'vue'
import { useFocusStore } from '../stores/focus'

const missions = [
  {
    name: 'SeaStrike-02',
    type: 'Maritime Strike',
    priority: 'High',
    status: 'Running',
    commander: 'Cmdr Lin',
    region: 'Blue Trench',
    updated: '08:22',
    objective: 'Suppress carrier group radar and open strike corridor.',
    failurePolicy: 'Continue',
    tasks: 9,
    stages: [
      { name: 'Recon', mode: 'Parallel', status: 'Succeeded', tasks: 3 },
      { name: 'Jammer', mode: 'Sequential', status: 'Running', tasks: 2 },
      { name: 'Strike', mode: 'Parallel', status: 'Pending', tasks: 4 },
    ],
  },
  {
    name: 'PolarShield-01',
    type: 'Air Defense',
    priority: 'Normal',
    status: 'Scheduled',
    commander: 'Cmdr Zhao',
    region: 'North Sector',
    updated: '08:10',
    objective: 'Maintain CAP and protect inbound convoy lanes.',
    failurePolicy: 'Abort',
    tasks: 6,
    stages: [
      { name: 'Scramble', mode: 'Parallel', status: 'Scheduled', tasks: 2 },
      { name: 'Escort', mode: 'Parallel', status: 'Pending', tasks: 4 },
    ],
  },
  {
    name: 'HarborWatch-07',
    type: 'ISR',
    priority: 'Low',
    status: 'Succeeded',
    commander: 'Cmdr Mei',
    region: 'Harbor Delta',
    updated: '07:58',
    objective: 'Confirm infrastructure damage assessment.',
    failurePolicy: 'Continue',
    tasks: 4,
    stages: [
      { name: 'Recon', mode: 'Parallel', status: 'Succeeded', tasks: 2 },
      { name: 'Analyze', mode: 'Sequential', status: 'Succeeded', tasks: 2 },
    ],
  },
  {
    name: 'SilentPath-03',
    type: 'Covert Ops',
    priority: 'High',
    status: 'Failed',
    commander: 'Cmdr Gao',
    region: 'Black Ridge',
    updated: '07:40',
    objective: 'Disable listening post without alerting patrols.',
    failurePolicy: 'Abort',
    tasks: 5,
    stages: [
      { name: 'Insertion', mode: 'Parallel', status: 'Failed', tasks: 3 },
      { name: 'Extraction', mode: 'Parallel', status: 'Pending', tasks: 2 },
    ],
  },
]

const selected = ref(missions[0])
const focusStore = useFocusStore()

watch(selected, (value) => {
  focusStore.setFocus('mission', value)
}, { immediate: true })

const priorityTone = (priority) => {
  if (priority === 'High') return 'warn'
  if (priority === 'Low') return 'muted'
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
          <span class="badge">4 tracked</span>
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
            :class="{ active: selected.name === mission.name }"
            style="--cols: 1.3fr 0.7fr 0.7fr 0.6fr 0.7fr"
            type="button"
            @click="selected = mission"
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
        </div>
      </div>

      <div class="panel">
        <div class="panel-header">
          <div>
            <div class="panel-title">Mission Detail</div>
            <div class="panel-sub">{{ selected.region }}</div>
          </div>
          <span class="badge" :class="selected.status.toLowerCase()">{{ selected.status }}</span>
        </div>

        <div class="detail-card">
          <div class="detail-title">{{ selected.name }}</div>
          <div class="detail-meta">
            <span class="badge" :class="priorityTone(selected.priority)">{{ selected.priority }} Priority</span>
            <span class="badge muted">{{ selected.type }}</span>
            <span class="badge">Failure: {{ selected.failurePolicy }}</span>
          </div>
          <div class="detail-info">
            <div class="kv">
              <span>Commander</span>
              <span>{{ selected.commander }}</span>
            </div>
            <div class="kv">
              <span>Region</span>
              <span>{{ selected.region }}</span>
            </div>
            <div class="kv">
              <span>FlightTasks</span>
              <span>{{ selected.tasks }}</span>
            </div>
            <div class="kv">
              <span>Objective</span>
              <span>{{ selected.objective }}</span>
            </div>
          </div>
        </div>

        <div class="panel">
          <div class="panel-header">
            <div>
              <div class="panel-title">Stage Flow</div>
              <div class="panel-sub">Parallel and sequential segments.</div>
            </div>
            <span class="badge">{{ selected.stages.length }} stages</span>
          </div>
          <div class="flow-line">
            <div
              v-for="stage in selected.stages"
              :key="stage.name"
              class="flow-node"
              :class="stage.status.toLowerCase()"
            >
              <div class="flow-title">{{ stage.name }}</div>
              <div class="flow-meta">{{ stage.mode }} · {{ stage.tasks }} tasks</div>
              <span class="badge" :class="stage.status.toLowerCase()">{{ stage.status }}</span>
            </div>
          </div>
        </div>
      </div>
    </section>
  </section>
</template>
