<script setup>
import { computed, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useFocusStore } from '../stores/focus'
import { useDataStore } from '../stores/data'

const dataStore = useDataStore()
const { nodes, events } = storeToRefs(dataStore)
const isLoading = computed(() => dataStore.loading.nodes)

const selectedKey = ref('')
const focusStore = useFocusStore()

const selected = computed(() => (
  nodes.value.find((item) => item.name === selectedKey.value) || null
))

watch(nodes, (list) => {
  if (!list.length) return
  const exists = list.some((item) => item.name === selectedKey.value)
  if (!exists) {
    selectedKey.value = list[0].name
  }
}, { immediate: true })

watch(selected, (value) => {
  if (value) {
    focusStore.setFocus('node', value)
  }
}, { immediate: true })

const readyCount = computed(() => nodes.value.filter((item) => item.status === 'Ready').length)
const totalNodes = computed(() => nodes.value.length)
const podsInUse = computed(() => {
  return nodes.value.reduce((sum, item) => {
    const used = Number(String(item.pods).split('/')[0])
    return Number.isFinite(used) ? sum + used : sum
  }, 0)
})
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
        <div class="stat-value">{{ readyCount }} / {{ totalNodes }}</div>
        <div class="stat-meta">{{ totalNodes - readyCount }} node tainted</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">Pods Running</div>
        <div class="stat-value">{{ podsInUse }}</div>
        <div class="stat-meta">Across {{ totalNodes }} nodes</div>
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
          <span v-if="isLoading" class="badge warn">Loading</span>
          <span v-else class="badge">{{ nodes.length }} nodes</span>
        </div>
        <div class="node-grid">
          <button
            v-for="node in nodes"
            :key="node.name"
            class="node-card"
            :class="{ 'is-active': selectedKey === node.name }"
            type="button"
            @click="selectedKey = node.name"
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
          <div v-if="!nodes.length && !isLoading" class="empty-state">No nodes available.</div>
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
          <div v-if="!events.length" class="empty-state">No events available.</div>
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
