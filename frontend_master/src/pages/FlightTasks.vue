<script setup>
import { ref, watch } from 'vue'
import { useFocusStore } from '../stores/focus'

const flightTasks = [
  {
    name: 'ft-alpha-9',
    stage: 'Jammer',
    mission: 'SeaStrike-02',
    status: 'Scheduled',
    pod: 'ft-alpha-9-pod',
    node: 'worker-1',
    weapon: 'PL-15',
    attempts: 2,
    scheduledAt: '08:14',
    conditions: [
      { label: 'PodScheduled', tone: 'ok', detail: 'Assigned to worker-1' },
      { label: 'ImagePullBackOff', tone: 'err', detail: 'task container pull failed' },
    ],
    constraints: ['type:J-20', 'hardpoint:2+', 'fuel>70%', 'zone:alpha'],
    podStatus: 'Pending',
    sidecars: ['weapon-pl-15'],
  },
  {
    name: 'ft-jammer-1',
    stage: 'Jammer',
    mission: 'SeaStrike-02',
    status: 'Running',
    pod: 'ft-jammer-1-pod',
    node: 'worker-2',
    weapon: 'PL-10',
    attempts: 1,
    scheduledAt: '08:03',
    conditions: [
      { label: 'PodScheduled', tone: 'ok', detail: 'Assigned to worker-2' },
      { label: 'ContainersReady', tone: 'ok', detail: '2/2 running' },
    ],
    constraints: ['type:J-16', 'jammer', 'zone:alpha'],
    podStatus: 'Running',
    sidecars: ['weapon-pl-10'],
  },
  {
    name: 'ft-strike-2',
    stage: 'Strike',
    mission: 'SeaStrike-02',
    status: 'Pending',
    pod: '--',
    node: '--',
    weapon: 'PL-15',
    attempts: 0,
    scheduledAt: '--',
    conditions: [
      { label: 'PodScheduled', tone: 'warn', detail: 'Awaiting stage readiness' },
    ],
    constraints: ['type:J-20', 'hardpoint:4+', 'fuel>85%'],
    podStatus: 'Pending',
    sidecars: ['weapon-pl-15'],
  },
  {
    name: 'ft-escort-2',
    stage: 'Escort',
    mission: 'PolarShield-01',
    status: 'Scheduled',
    pod: 'ft-escort-2-pod',
    node: 'pending',
    weapon: 'PL-10',
    attempts: 3,
    scheduledAt: '08:19',
    conditions: [
      { label: 'FailedScheduling', tone: 'warn', detail: '0/2 nodes match affinity' },
    ],
    constraints: ['type:J-11', 'fuel>60%', 'zone:north'],
    podStatus: 'Pending',
    sidecars: ['weapon-pl-10'],
  },
]

const selected = ref(flightTasks[0])
const focusStore = useFocusStore()

watch(selected, (value) => {
  focusStore.setFocus('flighttask', value)
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
          <span class="badge">{{ flightTasks.length }} tasks</span>
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
            :class="{ active: selected.name === task.name }"
            style="--cols: 1.2fr 0.8fr 0.7fr 0.7fr 0.6fr"
            type="button"
            @click="selected = task"
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
        </div>
      </div>

      <div class="panel">
        <div class="panel-header">
          <div>
            <div class="panel-title">Task Detail</div>
            <div class="panel-sub">{{ selected.mission }} / {{ selected.stage }}</div>
          </div>
          <span class="badge" :class="selected.status.toLowerCase()">{{ selected.status }}</span>
        </div>

        <div class="detail-card">
          <div class="detail-title">{{ selected.name }}</div>
          <div class="detail-meta">
            <span class="badge muted">Pod {{ selected.pod }}</span>
            <span class="badge">Node {{ selected.node }}</span>
          </div>
          <div class="detail-info">
            <div class="kv">
              <span>Scheduled At</span>
              <span>{{ selected.scheduledAt }}</span>
            </div>
            <div class="kv">
              <span>Scheduling Attempts</span>
              <span>{{ selected.attempts }}</span>
            </div>
            <div class="kv">
              <span>Pod Status</span>
              <span>{{ selected.podStatus }}</span>
            </div>
            <div class="kv">
              <span>Sidecars</span>
              <span>{{ selected.sidecars.join(', ') }}</span>
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
            <span v-for="item in selected.constraints" :key="item" class="chip">{{ item }}</span>
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
            <div v-for="item in selected.conditions" :key="item.label" class="condition-item">
              <span class="badge" :class="item.tone">{{ item.label }}</span>
              <span class="condition-detail">{{ item.detail }}</span>
            </div>
          </div>
        </div>
      </div>
    </section>
  </section>
</template>
