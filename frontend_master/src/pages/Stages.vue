<script setup>
import { computed, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useFocusStore } from '../stores/focus'
import { useDataStore } from '../stores/data'

const dataStore = useDataStore()
const { stages } = storeToRefs(dataStore)
const isLoading = computed(() => dataStore.loading.stages)

const selectedKey = ref('')
const query = ref('')
const statusFilter = ref('all')
const missionFilter = ref('all')
const modeFilter = ref('all')
const sortKey = ref('index')
const sortDir = ref('asc')
const focusStore = useFocusStore()

const emptyStage = {
  name: '--',
  mission: '--',
  index: '--',
  mode: '--',
  status: '--',
  timeout: '--',
  dependsOn: [],
  tasks: [],
}

const buildKey = (item) => `${item.mission}::${item.name}`

const statusOptions = [
  { value: 'all', label: '全部' },
  { value: 'Running', label: '运行中' },
  { value: 'Scheduled', label: '已调度' },
  { value: 'Pending', label: '待执行' },
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

const stageIndexScore = (value) => {
  const index = Number(value)
  return Number.isNaN(index) ? 0 : index
}

const stageCounts = computed(() => {
  const counts = { all: stages.value.length }
  stages.value.forEach((stage) => {
    const key = stage.status || 'Unknown'
    counts[key] = (counts[key] || 0) + 1
  })
  return counts
})

const missionOptions = computed(() => {
  const set = new Set()
  stages.value.forEach((stage) => {
    if (stage.mission && stage.mission !== '--') set.add(stage.mission)
  })
  return ['all', ...Array.from(set).sort((a, b) => a.localeCompare(b))]
})

const modeOptions = computed(() => {
  const set = new Set()
  stages.value.forEach((stage) => {
    if (stage.mode && stage.mode !== '--') set.add(stage.mode)
  })
  return ['all', ...Array.from(set)]
})

const filteredStages = computed(() => {
  const text = String(query.value || '').trim().toLowerCase()
  let list = stages.value

  if (text) {
    list = list.filter((stage) => {
      const haystack = [stage.name, stage.mission].join(' ').toLowerCase()
      return haystack.includes(text)
    })
  }

  if (statusFilter.value !== 'all') {
    list = list.filter((stage) => stage.status === statusFilter.value)
  }

  if (missionFilter.value !== 'all') {
    list = list.filter((stage) => stage.mission === missionFilter.value)
  }

  if (modeFilter.value !== 'all') {
    list = list.filter((stage) => stage.mode === modeFilter.value)
  }

  const sorted = [...list].sort((a, b) => {
    const aValue = (() => {
      if (sortKey.value === 'name') return a.name
      if (sortKey.value === 'mission') return a.mission
      if (sortKey.value === 'status') return statusScore(a.status)
      if (sortKey.value === 'mode') return a.mode
      return stageIndexScore(a.index)
    })()

    const bValue = (() => {
      if (sortKey.value === 'name') return b.name
      if (sortKey.value === 'mission') return b.mission
      if (sortKey.value === 'status') return statusScore(b.status)
      if (sortKey.value === 'mode') return b.mode
      return stageIndexScore(b.index)
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
  filteredStages.value.find((item) => buildKey(item) === selectedKey.value) || null
))

const selectedSafe = computed(() => selected.value || emptyStage)

const stageSubLabel = (stage) => {
  const parts = [`超时 ${stage.timeout}`]
  if (stage.dependsOn && stage.dependsOn.length) {
    parts.push(`依赖 ${stage.dependsOn.join(', ')}`)
  }
  return parts.join(' · ')
}

const toggleSort = () => {
  sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
}

watch(filteredStages, (list) => {
  if (!list.length) {
    selectedKey.value = ''
    return
  }
  const exists = list.some((item) => buildKey(item) === selectedKey.value)
  if (!exists) {
    selectedKey.value = buildKey(list[0])
  }
}, { immediate: true })

const selectStage = (stage) => {
  selectedKey.value = buildKey(stage)
  focusStore.setFocus('stage', stage)
}
</script>

<template>
  <section class="page">
    <header class="page-header">
      <div>
        <div class="page-title">任务阶段</div>
        <div class="page-sub">阶段执行流程与依赖关系。</div>
      </div>
      <div class="page-actions">
        <button class="ghost small">查看甘特图</button>
        <button class="ghost small">导出 YAML</button>
      </div>
    </header>

    <section class="split-grid">
      <div class="panel">
      <div class="panel-header">
        <div>
          <div class="panel-title">阶段列表</div>
          <div class="panel-sub">按任务和序列排序。</div>
        </div>
        <span v-if="isLoading" class="badge warn">加载中</span>
        <span v-else class="badge">{{ filteredStages.length }} / {{ stages.length }} 个阶段</span>
      </div>
      <div class="panel-toolbar">
        <div class="filter-row">
          <input
            v-model="query"
            class="input"
            type="search"
            placeholder="搜索阶段或任务"
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
              <span class="count-pill">{{ stageCounts[option.value] || 0 }}</span>
            </button>
          </div>
        </div>
        <div class="filter-row">
          <div class="filter-group">
            <span class="filter-label">任务</span>
            <select v-model="missionFilter" class="select">
              <option v-for="mission in missionOptions" :key="mission" :value="mission">
                {{ mission === 'all' ? '全部任务' : mission }}
              </option>
            </select>
          </div>
          <div class="filter-group">
            <span class="filter-label">模式</span>
            <select v-model="modeFilter" class="select">
              <option v-for="mode in modeOptions" :key="mode" :value="mode">
                {{ mode === 'all' ? '全部模式' : mode }}
              </option>
            </select>
          </div>
          <div class="filter-group">
            <span class="filter-label">排序</span>
            <select v-model="sortKey" class="select">
              <option value="index">序号</option>
              <option value="name">名称</option>
              <option value="mission">任务</option>
              <option value="status">状态</option>
              <option value="mode">模式</option>
            </select>
            <button type="button" class="ghost small" @click="toggleSort">
              {{ sortDir === 'asc' ? '升序' : '降序' }}
            </button>
          </div>
        </div>
      </div>
      <div class="data-table">
        <div class="data-row is-head" style="--cols: 1.2fr 0.9fr 0.7fr 0.7fr 0.6fr">
          <span>阶段</span>
          <span>任务</span>
            <span>状态</span>
            <span>模式</span>
            <span>序号</span>
          </div>
          <button
            v-for="stage in filteredStages"
            :key="stage.name + stage.mission"
            class="data-row is-selectable"
            :class="{ active: selectedKey === buildKey(stage) }"
            style="--cols: 1.2fr 0.9fr 0.7fr 0.7fr 0.6fr"
            type="button"
            @click="selectStage(stage)"
          >
            <div class="cell-main">
              <div class="cell-title">{{ stage.name }}</div>
              <div class="cell-sub">{{ stageSubLabel(stage) }}</div>
            </div>
            <span class="muted">{{ stage.mission }}</span>
            <span class="badge" :class="stage.status.toLowerCase()">{{ stage.status }}</span>
            <span class="badge muted">{{ stage.mode }}</span>
            <span class="muted">{{ stage.index }}</span>
          </button>
          <div v-if="!filteredStages.length && !isLoading" class="empty-state">
            没有符合当前筛选条件的阶段。
          </div>
        </div>
      </div>

      <div class="panel">
        <div class="panel-header">
          <div>
            <div class="panel-title">阶段详情</div>
            <div class="panel-sub">{{ selectedSafe.mission }}</div>
          </div>
          <span class="badge" :class="String(selectedSafe.status).toLowerCase()">{{ selectedSafe.status }}</span>
        </div>

        <div class="detail-card">
          <div class="detail-title">{{ selectedSafe.name }}</div>
          <div class="detail-meta">
            <span class="badge muted">{{ selectedSafe.mode }}</span>
            <span class="badge">超时 {{ selectedSafe.timeout }}</span>
          </div>
          <div class="detail-info">
            <div class="kv">
              <span>序列</span>
              <span>阶段 {{ selectedSafe.index }}</span>
            </div>
            <div class="kv">
              <span>依赖项</span>
              <span>{{ selectedSafe.dependsOn.length ? selectedSafe.dependsOn.join(', ') : '无' }}</span>
            </div>
            <div class="kv">
              <span>飞行任务</span>
              <span>{{ selectedSafe.tasks.length }}</span>
            </div>
          </div>
        </div>

        <div class="panel">
          <div class="panel-header">
            <div>
              <div class="panel-title">依赖关系</div>
              <div class="panel-sub">上游阶段执行门控。</div>
            </div>
          </div>
          <div class="chip-row">
            <span v-for="item in selectedSafe.dependsOn" :key="item" class="chip">{{ item }}</span>
            <div v-if="!selectedSafe.dependsOn.length" class="empty-state">无依赖项。</div>
          </div>
        </div>

        <div class="panel">
          <div class="panel-header">
            <div>
              <div class="panel-title">飞行任务</div>
              <div class="panel-sub">执行顺序与当前节点。</div>
            </div>
            <span class="badge">{{ selectedSafe.tasks.length }} 个任务</span>
          </div>
          <div class="task-table">
          <div class="task-row task-head">
            <div class="cell-start">名称</div>
            <div class="cell-center">状态</div>
            <div class="cell-center">预计时间</div>
            <div class="cell-center">节点</div>
          </div>
          <div v-for="task in selectedSafe.tasks" :key="task.name" class="task-row">
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
            <div v-if="!selectedSafe.tasks.length" class="empty-state">暂无可用任务。</div>
          </div>
        </div>
      </div>
    </section>
  </section>
</template>
