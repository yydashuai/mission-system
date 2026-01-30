<script setup>
import { computed, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useFocusStore } from '../stores/focus'
import { useDataStore } from '../stores/data'

const dataStore = useDataStore()
const { missions } = storeToRefs(dataStore)
const isLoading = computed(() => dataStore.loading.missions)

const selectedKey = ref('')
const query = ref('')
const statusFilter = ref('all')
const sortKey = ref('updated')
const sortDir = ref('desc')
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

const statusOptions = [
  { value: 'all', label: '全部' },
  { value: '待执行', label: '待执行' },
  { value: '运行中', label: '运行中' },
  { value: '已完成', label: '已完成' },
  { value: '失败', label: '失败' },
  { value: '已取消', label: '已取消' },
]

const priorityScore = (value) => {
  const key = String(value || '').toLowerCase()
  if (key === 'critical') return 4
  if (key === 'high') return 3
  if (key === 'normal') return 2
  if (key === 'low') return 1
  return 0
}

const statusScore = (value) => {
  const key = String(value || '').toLowerCase()
  if (key === '已完成' || key === 'succeeded') return 5
  if (key === '运行中' || key === 'running') return 4
  if (key === '待执行' || key === 'pending') return 3
  if (key === '失败' || key === 'failed') return 2
  if (key === '已取消' || key === 'cancelled' || key === 'canceled') return 1
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

const missionCounts = computed(() => {
  const counts = { all: missions.value.length }
  missions.value.forEach((mission) => {
    const key = mission.status || 'Unknown'
    counts[key] = (counts[key] || 0) + 1
  })
  return counts
})

const filteredMissions = computed(() => {
  const text = String(query.value || '').trim().toLowerCase()
  const statusKey = statusFilter.value
  let list = missions.value

  if (text) {
    list = list.filter((mission) => {
      const haystack = [
        mission.name,
        mission.type,
        mission.commander,
        mission.region,
        mission.objective,
      ].join(' ').toLowerCase()
      return haystack.includes(text)
    })
  }

  if (statusKey !== 'all') {
    list = list.filter((mission) => mission.status === statusKey)
  }

  const sorted = [...list].sort((a, b) => {
    const aValue = (() => {
      if (sortKey.value === 'name') return a.name
      if (sortKey.value === 'priority') return priorityScore(a.priority)
      if (sortKey.value === 'status') return statusScore(a.status)
      if (sortKey.value === 'tasks') return a.tasks || 0
      if (sortKey.value === 'stages') return a.stages.length
      return timeScore(a.updated)
    })()

    const bValue = (() => {
      if (sortKey.value === 'name') return b.name
      if (sortKey.value === 'priority') return priorityScore(b.priority)
      if (sortKey.value === 'status') return statusScore(b.status)
      if (sortKey.value === 'tasks') return b.tasks || 0
      if (sortKey.value === 'stages') return b.stages.length
      return timeScore(b.updated)
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
  filteredMissions.value.find((item) => item.name === selectedKey.value) || null
))

const selectedSafe = computed(() => selected.value || emptyMission)

const toggleSort = () => {
  sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
}

watch(filteredMissions, (list) => {
  if (!list.length) {
    selectedKey.value = ''
    return
  }
  const exists = list.some((item) => item.name === selectedKey.value)
  if (!exists) {
    selectedKey.value = list[0].name
  }
}, { immediate: true })

const selectMission = (mission) => {
  selectedKey.value = mission.name
  focusStore.setFocus('mission', mission)
}

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
        <div class="page-title">任务管理</div>
        <div class="page-sub">任务列表、生命周期与阶段组成</div>
      </div>
      <div class="page-actions">
        <button class="ghost small">导出 JSON</button>
        <button class="primary">启动演示</button>
      </div>
    </header>

    <section class="split-grid">
      <div class="panel">
      <div class="panel-header">
        <div>
          <div class="panel-title">活跃任务</div>
          <div class="panel-sub">选择任务以查看阶段流程</div>
        </div>
        <span v-if="isLoading" class="badge warn">加载中</span>
        <span v-else class="badge">{{ filteredMissions.length }} / {{ missions.length }} 已跟踪</span>
      </div>

      <div class="panel-toolbar">
        <div class="filter-row">
          <input
            v-model="query"
            class="input"
            type="search"
            placeholder="搜索任务、指挥官、区域"
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
              <span class="count-pill">{{ missionCounts[option.value] || 0 }}</span>
            </button>
          </div>
        </div>
        <div class="filter-row">
          <div class="filter-group">
            <span class="filter-label">排序</span>
            <select v-model="sortKey" class="select">
              <option value="updated">更新时间</option>
              <option value="name">名称</option>
              <option value="priority">优先级</option>
              <option value="status">状态</option>
              <option value="tasks">飞行任务</option>
              <option value="stages">阶段</option>
            </select>
            <button type="button" class="ghost small" @click="toggleSort">
              {{ sortDir === 'asc' ? '升序' : '降序' }}
            </button>
          </div>
        </div>
      </div>

      <div class="data-table">
        <div class="data-row is-head" style="--cols: 1.3fr 0.7fr 0.7fr 0.6fr 0.7fr">
          <span>任务</span>
          <span>状态</span>
            <span>优先级</span>
            <span>阶段</span>
            <span>更新时间</span>
          </div>
          <button
            v-for="mission in filteredMissions"
            :key="mission.name"
            class="data-row is-selectable"
            :class="{ active: selectedKey === mission.name }"
            style="--cols: 1.3fr 0.7fr 0.7fr 0.6fr 0.7fr"
            type="button"
            @click="selectMission(mission)"
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
          <div v-if="!filteredMissions.length && !isLoading" class="empty-state">
            没有符合当前筛选条件的任务。
          </div>
        </div>
      </div>

      <div class="panel">
        <div class="panel-header">
          <div>
            <div class="panel-title">任务详情</div>
            <div class="panel-sub">{{ selectedSafe.region }}</div>
          </div>
          <span class="badge" :class="String(selectedSafe.status).toLowerCase()">{{ selectedSafe.status }}</span>
        </div>

        <div class="detail-card">
          <div class="detail-title">{{ selectedSafe.name }}</div>
          <div class="detail-meta">
            <span class="badge" :class="priorityTone(selectedSafe.priority)">{{ selectedSafe.priority }} 优先级</span>
            <span class="badge muted">{{ selectedSafe.type }}</span>
            <span class="badge">失败策略: {{ selectedSafe.failurePolicy }}</span>
          </div>
          <div class="detail-info">
            <div class="kv">
              <span>指挥官</span>
              <span>{{ selectedSafe.commander }}</span>
            </div>
            <div class="kv">
              <span>区域</span>
              <span>{{ selectedSafe.region }}</span>
            </div>
            <div class="kv">
              <span>飞行任务</span>
              <span>{{ selectedSafe.tasks }}</span>
            </div>
            <div class="kv">
              <span>目标</span>
              <span>{{ selectedSafe.objective }}</span>
            </div>
          </div>
        </div>

        <div class="panel">
          <div class="panel-header">
            <div>
              <div class="panel-title">阶段流程</div>
              <div class="panel-sub">并行与串行执行段</div>
            </div>
            <span class="badge">{{ selectedSafe.stages.length }} 个阶段</span>
          </div>
          <div class="flow-line">
          <div
            v-for="stage in selectedSafe.stages"
            :key="stage.name"
            class="flow-node"
            :class="stage.status.toLowerCase()"
          >
            <div class="flow-title">{{ stage.name }}</div>
            <div class="flow-meta">{{ stage.mode }} · {{ stage.tasks }} 个任务</div>
            <div v-if="stage.dependsOn && stage.dependsOn.length" class="flow-deps">
              依赖: {{ stage.dependsOn.join(', ') }}
            </div>
            <span class="badge" :class="stage.status.toLowerCase()">{{ stage.status }}</span>
          </div>
          <div v-if="!selectedSafe.stages.length" class="empty-state">暂无阶段信息。</div>
        </div>
        </div>
      </div>
    </section>
  </section>
</template>
