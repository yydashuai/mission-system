<script setup>
import { computed, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useFocusStore } from '../stores/focus'
import { useDataStore } from '../stores/data'

const dataStore = useDataStore()
const { flightTasks } = storeToRefs(dataStore)
const isLoading = computed(() => dataStore.loading.flightTasks)

const selectedKey = ref('')
const query = ref('')
const statusFilter = ref('all')
const missionFilter = ref('all')
const stageFilter = ref('all')
const sortKey = ref('scheduledAt')
const sortDir = ref('desc')
const focusStore = useFocusStore()

const emptyTask = {
  name: '--',
  stage: '--',
  mission: '--',
  status: '--',
  pod: '--',
  node: '--',
  weapon: '--',
  attempts: '--',
  scheduledAt: '--',
  conditions: [],
  constraints: [],
  podStatus: '--',
  sidecars: [],
}

const statusOptions = [
  { value: 'all', label: '全部' },
  { value: 'Running', label: '执行中' },
  { value: 'Scheduled', label: '已调度' },
  { value: 'Pending', label: '待调度' },
  { value: 'Succeeded', label: '已完成' },
  { value: 'Failed', label: '失败' },
]

const statusScore = (value) => {
  const key = String(value || '').toLowerCase()
  if (key === 'succeeded') return 5
  if (key === 'running') return 4
  if (key === 'scheduled') return 3
  if (key === 'pending') return 2
  if (key === 'failed') return 1
  return 0
}

const timeScore = (value) => {
  if (!value || value === '--') return 0
  const parts = String(value).split(':')
  if (parts.length === 2) {
    const hours = Number(parts[0])
    const minutes = Number(parts[1])
    if (!Number.isNaN(hours) && !Number.isNaN(minutes)) {
      return hours * 60 + minutes
    }
  }
  const date = new Date(value)
  if (!Number.isNaN(date.getTime())) return date.getTime()
  return 0
}

const taskCounts = computed(() => {
  const counts = { all: flightTasks.value.length }
  flightTasks.value.forEach((task) => {
    const key = task.status || 'Unknown'
    counts[key] = (counts[key] || 0) + 1
  })
  return counts
})

const missionOptions = computed(() => {
  const set = new Set()
  flightTasks.value.forEach((task) => {
    if (task.mission && task.mission !== '--') set.add(task.mission)
  })
  return ['all', ...Array.from(set).sort((a, b) => a.localeCompare(b))]
})

const stageOptions = computed(() => {
  const set = new Set()
  flightTasks.value.forEach((task) => {
    if (task.stage && task.stage !== '--') set.add(task.stage)
  })
  return ['all', ...Array.from(set).sort((a, b) => a.localeCompare(b))]
})

const filteredTasks = computed(() => {
  const text = String(query.value || '').trim().toLowerCase()
  let list = flightTasks.value

  if (text) {
    list = list.filter((task) => {
      const haystack = [task.name, task.mission, task.stage, task.weapon, task.node]
        .join(' ')
        .toLowerCase()
      return haystack.includes(text)
    })
  }

  if (statusFilter.value !== 'all') {
    list = list.filter((task) => task.status === statusFilter.value)
  }

  if (missionFilter.value !== 'all') {
    list = list.filter((task) => task.mission === missionFilter.value)
  }

  if (stageFilter.value !== 'all') {
    list = list.filter((task) => task.stage === stageFilter.value)
  }

  const sorted = [...list].sort((a, b) => {
    const aValue = (() => {
      if (sortKey.value === 'name') return a.name
      if (sortKey.value === 'status') return statusScore(a.status)
      if (sortKey.value === 'attempts') return a.attempts || 0
      if (sortKey.value === 'node') return a.node
      return timeScore(a.scheduledAt)
    })()

    const bValue = (() => {
      if (sortKey.value === 'name') return b.name
      if (sortKey.value === 'status') return statusScore(b.status)
      if (sortKey.value === 'attempts') return b.attempts || 0
      if (sortKey.value === 'node') return b.node
      return timeScore(b.scheduledAt)
    })()

    let result = 0
    if (typeof aValue === 'number' && typeof bValue === 'number') {
      result = aValue - bValue
    } else {
      result = String(aValue).localeCompare(String(bValue))
    }
    return sortDir.value === 'asc' ? result : -result
  })

  return sorted
})

const selected = computed(() => (
  filteredTasks.value.find((item) => item.name === selectedKey.value) || null
))

const selectedSafe = computed(() => selected.value || emptyTask)

const toggleSort = () => {
  sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
}

watch(filteredTasks, (list) => {
  if (!list.length) {
    selectedKey.value = ''
    return
  }
  const exists = list.some((item) => item.name === selectedKey.value)
  if (!exists) {
    selectedKey.value = list[0].name
  }
}, { immediate: true })

const selectTask = (task) => {
  selectedKey.value = task.name
  focusStore.setFocus('flighttask', task)
}
</script>

<template>
  <section class="page">
    <header class="page-header">
      <div>
        <div class="page-title">飞行任务</div>
        <div class="page-sub">调度详情、Pod 绑定及状态信息</div>
      </div>
      <div class="page-actions">
        <button class="ghost small">复制 YAML</button>
        <button class="ghost small">检查 Pod</button>
      </div>
    </header>

    <section class="split-grid">
      <div class="panel">
      <div class="panel-header">
        <div>
          <div class="panel-title">任务队列</div>
          <div class="panel-sub">选择任务查看调度上下文</div>
        </div>
        <span v-if="isLoading" class="badge warn">加载中</span>
        <span v-else class="badge">{{ filteredTasks.length }} / {{ flightTasks.length }} 个任务</span>
      </div>
      <div class="panel-toolbar">
        <div class="filter-row">
          <input
            v-model="query"
            class="input"
            type="search"
            placeholder="搜索任务、任务组、武器、节点"
          />
          <div class="segmented">
            <button
              v-for="option in statusOptions"
              :key="option.value"
              type="button"
              class="segmented-btn"
              :class="{ active: statusFilter === option.value }"
              @click="statusFilter = option.value"
            >
              <span>{{ option.label }}</span>
              <span class="count-pill">{{ taskCounts[option.value] || 0 }}</span>
            </button>
          </div>
        </div>
        <div class="filter-row">
          <div class="filter-group">
            <span class="filter-label">任务组</span>
            <select v-model="missionFilter" class="select">
              <option v-for="mission in missionOptions" :key="mission" :value="mission">
                {{ mission === 'all' ? '全部任务组' : mission }}
              </option>
            </select>
          </div>
          <div class="filter-group">
            <span class="filter-label">阶段</span>
            <select v-model="stageFilter" class="select">
              <option v-for="stage in stageOptions" :key="stage" :value="stage">
                {{ stage === 'all' ? '全部阶段' : stage }}
              </option>
            </select>
          </div>
          <div class="filter-group">
            <span class="filter-label">排序</span>
            <select v-model="sortKey" class="select">
              <option value="scheduledAt">调度时间</option>
              <option value="name">名称</option>
              <option value="status">状态</option>
              <option value="attempts">尝试次数</option>
              <option value="node">节点</option>
            </select>
            <button type="button" class="ghost small" @click="toggleSort">
              {{ sortDir === 'asc' ? '升序' : '降序' }}
            </button>
          </div>
        </div>
      </div>
      <div class="data-table">
        <div class="data-row is-head" style="--cols: 1.2fr 0.8fr 0.7fr 0.7fr 0.6fr">
          <span>任务</span>
          <span>阶段</span>
            <span>状态</span>
            <span>节点</span>
            <span>武器</span>
          </div>
          <button
            v-for="task in filteredTasks"
            :key="task.name"
            class="data-row is-selectable"
            :class="{ active: selectedKey === task.name }"
            style="--cols: 1.2fr 0.8fr 0.7fr 0.7fr 0.6fr"
            type="button"
            @click="selectTask(task)"
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
          <div v-if="!filteredTasks.length && !isLoading" class="empty-state">
            没有符合当前筛选条件的任务
          </div>
        </div>
      </div>

      <div class="panel">
        <div class="panel-header">
          <div>
            <div class="panel-title">任务详情</div>
            <div class="panel-sub">{{ selectedSafe.mission }} / {{ selectedSafe.stage }}</div>
          </div>
          <span class="badge" :class="String(selectedSafe.status).toLowerCase()">{{ selectedSafe.status }}</span>
        </div>

        <div class="detail-card">
          <div class="detail-title">{{ selectedSafe.name }}</div>
          <div class="detail-meta">
            <span class="badge muted">Pod {{ selectedSafe.pod }}</span>
            <span class="badge">节点 {{ selectedSafe.node }}</span>
          </div>
          <div class="detail-info">
            <div class="kv">
              <span>调度时间</span>
              <span>{{ selectedSafe.scheduledAt }}</span>
            </div>
            <div class="kv">
              <span>调度尝试次数</span>
              <span>{{ selectedSafe.attempts }}</span>
            </div>
            <div class="kv">
              <span>Pod 状态</span>
              <span>{{ selectedSafe.podStatus }}</span>
            </div>
            <div class="kv">
              <span>边车容器</span>
              <span>{{ selectedSafe.sidecars.join(', ') }}</span>
            </div>
          </div>
        </div>

        <div class="panel">
          <div class="panel-header">
            <div>
              <div class="panel-title">调度约束</div>
              <div class="panel-sub">节点亲和性及机型过滤要求</div>
            </div>
          </div>
          <div class="chip-row">
            <span v-for="item in selectedSafe.constraints" :key="item" class="chip">{{ item }}</span>
            <div v-if="!selectedSafe.constraints.length" class="empty-state">无约束数据</div>
          </div>
        </div>

        <div class="panel">
          <div class="panel-header">
            <div>
              <div class="panel-title">状态条件</div>
              <div class="panel-sub">Pod 及调度器实时信号</div>
            </div>
          </div>
          <div class="condition-list">
            <div v-for="item in selectedSafe.conditions" :key="item.label" class="condition-item">
              <span class="badge" :class="item.tone">{{ item.label }}</span>
              <span class="condition-detail">{{ item.detail }}</span>
            </div>
            <div v-if="!selectedSafe.conditions.length" class="empty-state">无状态更新</div>
          </div>
        </div>
      </div>
    </section>
  </section>
</template>
