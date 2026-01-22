<script setup>
const events = [
  { time: '08:21', label: 'FT-ghost-12 pending: node affinity unmet', tone: 'warn' },
  { time: '08:18', label: 'Weapon PL-15 injected into ft-rapid-3', tone: 'ok' },
  { time: '08:07', label: 'Mission SeaStrike-02 stage-2 running', tone: 'ok' },
  { time: '07:55', label: 'ImagePullBackOff on ft-alpha-9', tone: 'err' },
]
</script>

<template>
  <aside class="detail-panel">
    <div class="section-title">Focus: FlightTask</div>
    <div class="detail-card">
      <div class="detail-title">ft-alpha-9</div>
      <div class="detail-meta">
        <span class="badge err">ImagePullBackOff</span>
        <span class="badge muted">Scheduled</span>
      </div>
      <div class="detail-info">
        <div class="kv">
          <span>Mission</span>
          <span>SeaStrike-02</span>
        </div>
        <div class="kv">
          <span>Stage</span>
          <span>Strike-Phase</span>
        </div>
        <div class="kv">
          <span>Node</span>
          <span>worker-1</span>
        </div>
        <div class="kv">
          <span>Weapon</span>
          <span>PL-15</span>
        </div>
      </div>
    </div>

    <div class="section-title">Recent Events</div>
    <div class="event-list">
      <div v-for="event in events" :key="event.time + event.label" class="event-item">
        <span class="event-time">{{ event.time }}</span>
        <span class="event-dot" :class="event.tone"></span>
        <span class="event-text">{{ event.label }}</span>
      </div>
    </div>
  </aside>
</template>
