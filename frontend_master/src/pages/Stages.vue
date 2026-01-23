<script setup>
import { ref, watch } from 'vue'
import { useFocusStore } from '../stores/focus'

const stages = [
  {
    name: 'Recon',
    mission: 'SeaStrike-02',
    index: 1,
    mode: 'Parallel',
    status: 'Succeeded',
    timeout: '30m',
    dependsOn: [],
    tasks: [
      { name: 'ft-recon-1', status: 'Succeeded', eta: 'done', node: 'worker-1' },
      { name: 'ft-recon-2', status: 'Succeeded', eta: 'done', node: 'worker-2' },
    ],
  },
  {
    name: 'Jammer',
    mission: 'SeaStrike-02',
    index: 2,
    mode: 'Sequential',
    status: 'Running',
    timeout: '25m',
    dependsOn: ['Recon'],
    tasks: [
      { name: 'ft-jammer-1', status: 'Running', eta: '12m', node: 'worker-1' },
      { name: 'ft-jammer-2', status: 'Scheduled', eta: '7m', node: 'pending' },
    ],
  },
  {
    name: 'Strike',
    mission: 'SeaStrike-02',
    index: 3,
    mode: 'Parallel',
    status: 'Pending',
    timeout: '40m',
    dependsOn: ['Jammer'],
    tasks: [
      { name: 'ft-strike-1', status: 'Pending', eta: '--', node: '--' },
      { name: 'ft-strike-2', status: 'Pending', eta: '--', node: '--' },
      { name: 'ft-strike-3', status: 'Pending', eta: '--', node: '--' },
    ],
  },
  {
    name: 'Escort',
    mission: 'PolarShield-01',
    index: 2,
    mode: 'Parallel',
    status: 'Scheduled',
    timeout: '20m',
    dependsOn: ['Scramble'],
    tasks: [
      { name: 'ft-escort-1', status: 'Scheduled', eta: '4m', node: 'pending' },
      { name: 'ft-escort-2', status: 'Scheduled', eta: '4m', node: 'pending' },
    ],
  },
]

const selected = ref(stages[1])
const focusStore = useFocusStore()

watch(selected, (value) => {
  focusStore.setFocus('stage', value)
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
          <span class="badge">{{ stages.length }} stages</span>
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
            :class="{ active: selected.name === stage.name && selected.mission === stage.mission }"
            style="--cols: 1.2fr 0.9fr 0.7fr 0.7fr 0.6fr"
            type="button"
            @click="selected = stage"
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
        </div>
      </div>

      <div class="panel">
        <div class="panel-header">
          <div>
            <div class="panel-title">Stage Detail</div>
            <div class="panel-sub">{{ selected.mission }}</div>
          </div>
          <span class="badge" :class="selected.status.toLowerCase()">{{ selected.status }}</span>
        </div>

        <div class="detail-card">
          <div class="detail-title">{{ selected.name }}</div>
          <div class="detail-meta">
            <span class="badge muted">{{ selected.mode }}</span>
            <span class="badge">Timeout {{ selected.timeout }}</span>
          </div>
          <div class="detail-info">
            <div class="kv">
              <span>Sequence</span>
              <span>Stage {{ selected.index }}</span>
            </div>
            <div class="kv">
              <span>Depends On</span>
              <span>{{ selected.dependsOn.length ? selected.dependsOn.join(', ') : 'None' }}</span>
            </div>
            <div class="kv">
              <span>FlightTasks</span>
              <span>{{ selected.tasks.length }}</span>
            </div>
          </div>
        </div>

        <div class="panel">
          <div class="panel-header">
            <div>
              <div class="panel-title">FlightTasks</div>
              <div class="panel-sub">Execution order and current node.</div>
            </div>
            <span class="badge">{{ selected.tasks.length }} tasks</span>
          </div>
          <div class="task-table">
            <div class="task-row task-head">
              <span>Name</span>
              <span>Status</span>
              <span>ETA</span>
              <span>Node</span>
            </div>
            <div v-for="task in selected.tasks" :key="task.name" class="task-row">
              <span class="task-name">{{ task.name }}</span>
              <span class="badge" :class="task.status.toLowerCase()">{{ task.status }}</span>
              <span>{{ task.eta }}</span>
              <span class="muted">{{ task.node }}</span>
            </div>
          </div>
        </div>
      </div>
    </section>
  </section>
</template>
