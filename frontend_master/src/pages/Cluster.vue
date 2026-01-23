<script setup>
import { ref, watch } from 'vue'
import { useFocusStore } from '../stores/focus'

const nodes = [
  {
    name: 'master-0',
    role: 'control-plane',
    status: 'Ready',
    cpu: 72,
    memory: 64,
    pods: '22 / 110',
    zone: 'alpha',
  },
  {
    name: 'worker-1',
    role: 'worker',
    status: 'Ready',
    cpu: 54,
    memory: 58,
    pods: '31 / 110',
    zone: 'alpha',
  },
  {
    name: 'worker-2',
    role: 'worker',
    status: 'NotReady',
    cpu: 0,
    memory: 0,
    pods: '0 / 110',
    zone: 'beta',
  },
]

const events = [
  { time: '08:23', scope: 'scheduler', level: 'warn', message: 'FailedScheduling on ft-escort-2' },
  { time: '08:20', scope: 'kubelet', level: 'err', message: 'ImagePullBackOff ft-alpha-9' },
  { time: '08:18', scope: 'controller', level: 'ok', message: 'Weapon PL-15 available' },
  { time: '08:12', scope: 'controller', level: 'ok', message: 'Stage Jammer running' },
]

const selected = ref(nodes[0])
const focusStore = useFocusStore()

watch(selected, (value) => {
  focusStore.setFocus('node', value)
}, { immediate: true })
</script>

<template>
  <section class="page">
    <header class="page-header">
      <div>
        <div class="page-title">Cluster</div>
        <div class="page-sub">Nodes, pods, events, and basic health.</div>
      </div>
      <div class="page-actions">
        <button class="ghost small">Refresh</button>
        <button class="ghost small">Download Events</button>
      </div>
    </header>

    <section class="stat-grid">
      <div class="stat-card">
        <div class="stat-label">Nodes Ready</div>
        <div class="stat-value">2 / 3</div>
        <div class="stat-meta">1 node tainted</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">Pods Running</div>
        <div class="stat-value">38</div>
        <div class="stat-meta">6 Pending</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">CoreDNS</div>
        <div class="stat-value">Healthy</div>
        <div class="stat-meta">2 replicas</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">Calico</div>
        <div class="stat-value">Degraded</div>
        <div class="stat-meta">1 warning</div>
      </div>
    </section>

    <section class="split-grid">
      <div class="panel">
        <div class="panel-header">
          <div>
            <div class="panel-title">Nodes</div>
            <div class="panel-sub">Capacity and utilization snapshot.</div>
          </div>
          <span class="badge">3 nodes</span>
        </div>
        <div class="node-grid">
          <button
            v-for="node in nodes"
            :key="node.name"
            class="node-card"
            :class="{ 'is-active': selected.name === node.name }"
            type="button"
            @click="selected = node"
          >
            <div class="node-head">
              <div>
                <div class="node-name">{{ node.name }}</div>
                <div class="node-meta">{{ node.role }} · zone {{ node.zone }}</div>
              </div>
              <span class="badge" :class="node.status === 'Ready' ? 'ok' : 'err'">{{ node.status }}</span>
            </div>
            <div class="util-row">
              <span class="muted">CPU</span>
              <div class="util-track">
                <div class="util-fill" :style="{ width: node.cpu + '%' }"></div>
              </div>
              <span class="muted">{{ node.cpu }}%</span>
            </div>
            <div class="util-row">
              <span class="muted">Memory</span>
              <div class="util-track">
                <div class="util-fill" :style="{ width: node.memory + '%' }"></div>
              </div>
              <span class="muted">{{ node.memory }}%</span>
            </div>
            <div class="kv">
              <span>Pods</span>
              <span>{{ node.pods }}</span>
            </div>
          </button>
        </div>
      </div>

      <div class="panel">
        <div class="panel-header">
          <div>
            <div class="panel-title">Recent Events</div>
            <div class="panel-sub">Last 24h from scheduler and kubelet.</div>
          </div>
          <span class="badge">{{ events.length }} events</span>
        </div>
        <div class="event-stream">
          <div v-for="item in events" :key="item.time + item.message" class="event-row">
            <span class="event-time">{{ item.time }}</span>
            <span class="badge" :class="item.level">{{ item.scope }}</span>
            <span class="event-text">{{ item.message }}</span>
          </div>
        </div>

        <div class="panel">
          <div class="panel-header">
            <div>
              <div class="panel-title">Namespaces</div>
              <div class="panel-sub">Pod distribution snapshot.</div>
            </div>
          </div>
          <div class="namespace-list">
            <div class="namespace-row">
              <span>airforce-system</span>
              <span class="muted">14 pods</span>
            </div>
            <div class="namespace-row">
              <span>kube-system</span>
              <span class="muted">12 pods</span>
            </div>
            <div class="namespace-row">
              <span>default</span>
              <span class="muted">8 pods</span>
            </div>
            <div class="namespace-row">
              <span>mission-runtime</span>
              <span class="muted">4 pods</span>
            </div>
          </div>
        </div>
      </div>
    </section>
  </section>
</template>
