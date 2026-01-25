<script setup>
const kpis = [
  { label: 'Active Missions', value: '3', delta: '+1 in 1h', tone: 'ok' },
  { label: 'Stages Running', value: '5', delta: '2 parallel', tone: 'warn' },
  { label: 'FlightTasks', value: '18', delta: '3 waiting', tone: 'warn' },
  { label: 'Weapons Ready', value: '42', delta: '4 online', tone: 'ok' },
]

const tasks = [
  { name: 'Recon-Phase', status: 'Running', eta: '12m', node: 'worker-2' },
  { name: 'Jammer-Phase', status: 'Scheduled', eta: '5m', node: 'pending' },
  { name: 'Strike-Phase', status: 'Pending', eta: '--', node: '--' },
]
</script>

<template>
  <section class="page">
    <section class="hero">
      <div>
        <div class="hero-title">Operational Overview</div>
        <div class="hero-sub">
          Mission chain status, scheduling pressure, and deployment health.
        </div>
      </div>
      <div class="hero-tags">
        <span class="badge ok">All Controllers Healthy</span>
        <span class="badge warn">2 Tasks Waiting</span>
      </div>
    </section>

    <section class="kpi-grid">
      <div
        v-for="(card, index) in kpis"
        :key="card.label"
        class="kpi-card"
        :data-idx="index"
      >
        <div class="kpi-label">{{ card.label }}</div>
        <div class="kpi-value">{{ card.value }}</div>
        <div class="kpi-delta" :class="card.tone">{{ card.delta }}</div>
      </div>
    </section>

    <section class="content-grid">
      <div class="panel">
        <div class="panel-header">
          <div>
            <div class="panel-title">Active Stage Timeline</div>
            <div class="panel-sub">SeaStrike-02 / Stage-2</div>
          </div>
          <span class="badge">Parallel</span>
        </div>
        <div class="timeline">
          <div class="timeline-step done">
            <span class="dot"></span>
            <div>
              <div class="step-title">Stage Initialize</div>
              <div class="step-meta">07:40 completed</div>
            </div>
          </div>
          <div class="timeline-step active">
            <span class="dot"></span>
            <div>
              <div class="step-title">FlightTasks Scheduling</div>
              <div class="step-meta">02 tasks running</div>
            </div>
          </div>
          <div class="timeline-step">
            <span class="dot"></span>
            <div>
              <div class="step-title">Weapons Check</div>
              <div class="step-meta">waiting for ft-alpha-9</div>
            </div>
          </div>
        </div>
      </div>

      <div class="panel">
        <div class="panel-header">
          <div>
            <div class="panel-title">FlightTask Queue</div>
            <div class="panel-sub">Scheduling pressure by stage</div>
          </div>
          <button class="ghost small">View All</button>
        </div>
        <div class="task-table">
          <div class="task-row task-head">
            <div class="cell-start">Name</div>
            <div class="cell-center">Status</div>
            <div class="cell-center">ETA</div>
            <div class="cell-center">Node</div>
          </div>
          <div v-for="task in tasks" :key="task.name" class="task-row">
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
        </div>
      </div>
    </section>
  </section>
</template>
