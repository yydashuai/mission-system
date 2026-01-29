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

watch(nodes, (list) => {
  if (!list.length) return
  const exists = list.some((item) => item.name === selectedKey.value)
  if (!exists) {
    selectedKey.value = list[0].name
  }
}, { immediate: true })

const selectNode = (node) => {
  selectedKey.value = node.name
  focusStore.setFocus('node', node)
}

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
        <div class="page-title">集群</div>
        <div class="page-sub">节点、容器组、事件及基础健康状态。</div>
      </div>
      <div class="page-actions">
        <button class="ghost small">刷新</button>
        <button class="ghost small">下载事件</button>
      </div>
    </header>

    <section class="stat-grid">
      <div class="stat-card">
        <div class="stat-label">节点就绪</div>
        <div class="stat-value">{{ readyCount }} / {{ totalNodes }}</div>
        <div class="stat-meta">{{ totalNodes - readyCount }} 节点污点</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">运行容器组</div>
        <div class="stat-value">{{ podsInUse }}</div>
        <div class="stat-meta">分布于 {{ totalNodes }} 个节点</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">CoreDNS</div>
        <div class="stat-value">健康</div>
        <div class="stat-meta">2 副本</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">Calico</div>
        <div class="stat-value">降级</div>
        <div class="stat-meta">1 警告</div>
      </div>
    </section>

    <section class="split-grid">
      <div class="panel">
        <div class="panel-header">
          <div>
            <div class="panel-title">节点</div>
            <div class="panel-sub">容量与利用率快照。</div>
          </div>
          <span v-if="isLoading" class="badge warn">加载中</span>
          <span v-else class="badge">{{ nodes.length }} 节点</span>
        </div>
        <div class="node-grid">
          <button
            v-for="node in nodes"
            :key="node.name"
            class="node-card"
            :class="{ 'is-active': selectedKey === node.name }"
            type="button"
            @click="selectNode(node)"
          >
            <div class="node-head">
              <div>
                <div class="node-name">{{ node.name }}</div>
                <div class="node-meta">{{ node.role }} · 区域 {{ node.zone }}</div>
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
              <span class="muted">内存</span>
              <div class="util-track">
                <div class="util-fill" :style="{ width: node.memory + '%' }"></div>
              </div>
              <span class="muted">{{ node.memory }}%</span>
            </div>
            <div class="kv">
              <span>容器组</span>
              <span>{{ node.pods }}</span>
            </div>
          </button>
          <div v-if="!nodes.length && !isLoading" class="empty-state">无可用节点。</div>
        </div>
      </div>

      <div class="panel">
        <div class="panel-header">
          <div>
            <div class="panel-title">近期事件</div>
            <div class="panel-sub">调度器与 kubelet 最近 24 小时。</div>
          </div>
          <span class="badge">{{ events.length }} 事件</span>
        </div>
        <div class="event-stream">
          <div v-for="item in events" :key="item.time + item.message" class="event-row">
            <span class="event-time">{{ item.time }}</span>
            <span class="badge" :class="item.level">{{ item.scope }}</span>
            <span class="event-text">{{ item.message }}</span>
          </div>
          <div v-if="!events.length" class="empty-state">无可用事件。</div>
        </div>

        <div class="panel">
          <div class="panel-header">
            <div>
              <div class="panel-title">命名空间</div>
              <div class="panel-sub">容器组分布快照。</div>
            </div>
          </div>
          <div class="namespace-list">
            <div class="namespace-row">
              <span>airforce-system</span>
              <span class="muted">14 容器组</span>
            </div>
            <div class="namespace-row">
              <span>kube-system</span>
              <span class="muted">12 容器组</span>
            </div>
            <div class="namespace-row">
              <span>default</span>
              <span class="muted">8 容器组</span>
            </div>
            <div class="namespace-row">
              <span>mission-runtime</span>
              <span class="muted">4 容器组</span>
            </div>
          </div>
        </div>
      </div>
    </section>
  </section>
</template>
