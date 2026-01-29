<script setup>
import { computed, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useFocusStore } from '../stores/focus'
import { useDataStore } from '../stores/data'

const dataStore = useDataStore()
const { weapons } = storeToRefs(dataStore)
const isLoading = computed(() => dataStore.loading.weapons)

const selectedKey = ref('')
const query = ref('')
const statusFilter = ref('all')
const sortKey = ref('name')
const sortDir = ref('asc')
const focusStore = useFocusStore()

const emptyWeapon = {
  name: '--',
  status: '--',
  image: '--',
  version: '--',
  usage: '--',
  aircraft: [],
  hardpoints: [],
  type: '--',
}

const statusOptions = [
  { value: 'all', label: '全部' },
  { value: '可用', label: '可用' },
  { value: '更新中', label: '更新中' },
  { value: '已弃用', label: '已弃用' },
]

const statusScore = (value) => {
  const key = String(value || '').toLowerCase()
  if (key === '可用' || key === 'available') return 3
  if (key === '更新中' || key === 'updating') return 2
  if (key === '已弃用' || key === 'deprecated' || key === 'degraded' || key === 'retired') return 1
  return 0
}

const usageScore = (value) => {
  if (!value) return 0
  const match = String(value).match(/\d+/)
  return match ? Number(match[0]) : 0
}

const weaponCounts = computed(() => {
  const counts = { all: weapons.value.length }
  weapons.value.forEach((weapon) => {
    const key = weapon.status || 'Unknown'
    counts[key] = (counts[key] || 0) + 1
  })
  return counts
})

const filteredWeapons = computed(() => {
  const text = String(query.value || '').trim().toLowerCase()
  let list = weapons.value

  if (text) {
    list = list.filter((weapon) => {
      const haystack = [weapon.name, weapon.image, weapon.version, weapon.type]
        .join(' ')
        .toLowerCase()
      return haystack.includes(text)
    })
  }

  if (statusFilter.value !== 'all') {
    list = list.filter((weapon) => weapon.status === statusFilter.value)
  }

  const sorted = [...list].sort((a, b) => {
    const aValue = (() => {
      if (sortKey.value === 'status') return statusScore(a.status)
      if (sortKey.value === 'usage') return usageScore(a.usage)
      if (sortKey.value === 'version') return a.version
      return a.name
    })()

    const bValue = (() => {
      if (sortKey.value === 'status') return statusScore(b.status)
      if (sortKey.value === 'usage') return usageScore(b.usage)
      if (sortKey.value === 'version') return b.version
      return b.name
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
  filteredWeapons.value.find((item) => item.name === selectedKey.value) || null
))

const selectedSafe = computed(() => selected.value || emptyWeapon)

const toggleSort = () => {
  sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
}

watch(filteredWeapons, (list) => {
  if (!list.length) {
    selectedKey.value = ''
    return
  }
  const exists = list.some((item) => item.name === selectedKey.value)
  if (!exists) {
    selectedKey.value = list[0].name
  }
}, { immediate: true })

const selectWeapon = (weapon) => {
  selectedKey.value = weapon.name
  focusStore.setFocus('weapon', weapon)
}

const statusTone = (status) => {
  if (status === '可用') return 'ok'
  if (status === '更新中') return 'warn'
  if (status === '已弃用') return 'muted'
  return 'muted'
}
</script>

<template>
  <section class="page">
    <header class="page-header">
      <div>
        <div class="page-title">武器系统</div>
        <div class="page-sub">兼容性、容器规格与使用情况</div>
      </div>
      <div class="page-actions">
        <button class="ghost small">拉取镜像</button>
        <button class="ghost small">兼容性矩阵</button>
      </div>
    </header>

    <section class="split-grid">
      <div class="panel">
      <div class="panel-header">
        <div>
          <div class="panel-title">武器库</div>
          <div class="panel-sub">可用于飞行任务的边车容器包</div>
        </div>
        <span v-if="isLoading" class="badge warn">加载中</span>
        <span v-else class="badge">{{ filteredWeapons.length }} / {{ weapons.length }} 个包</span>
      </div>
      <div class="panel-toolbar">
        <div class="filter-row">
          <input
            v-model="query"
            class="input"
            type="search"
            placeholder="搜索武器、镜像、类型"
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
              <span class="count-pill">{{ weaponCounts[option.value] || 0 }}</span>
            </button>
          </div>
        </div>
        <div class="filter-row">
          <div class="filter-group">
            <span class="filter-label">排序</span>
            <select v-model="sortKey" class="select">
              <option value="name">名称</option>
              <option value="status">状态</option>
              <option value="usage">使用量</option>
              <option value="version">版本</option>
            </select>
            <button type="button" class="ghost small" @click="toggleSort">
              {{ sortDir === 'asc' ? '升序' : '降序' }}
            </button>
          </div>
        </div>
      </div>
      <div class="data-table">
        <div class="data-row is-head" style="--cols: 1.1fr 0.8fr 0.9fr 0.7fr 0.7fr">
          <span>武器</span>
          <span>状态</span>
            <span>镜像</span>
            <span>版本</span>
            <span>使用量</span>
          </div>
          <button
            v-for="weapon in filteredWeapons"
            :key="weapon.name"
            class="data-row is-selectable"
            :class="{ active: selectedKey === weapon.name }"
            style="--cols: 1.1fr 0.8fr 0.9fr 0.7fr 0.7fr"
            type="button"
            @click="selectWeapon(weapon)"
          >
            <div class="cell-main">
              <div class="cell-title">{{ weapon.name }}</div>
              <div class="cell-sub">{{ weapon.type }}</div>
            </div>
            <span class="badge" :class="statusTone(weapon.status)">{{ weapon.status }}</span>
            <span class="muted">{{ weapon.image }}</span>
            <span class="muted">{{ weapon.version }}</span>
            <span class="badge muted">{{ weapon.usage }}</span>
          </button>
          <div v-if="!filteredWeapons.length && !isLoading" class="empty-state">
            没有符合当前筛选条件的武器
          </div>
        </div>
      </div>

      <div class="panel">
        <div class="panel-header">
          <div>
            <div class="panel-title">武器详情</div>
            <div class="panel-sub">{{ selectedSafe.image }}</div>
          </div>
          <span class="badge" :class="statusTone(selectedSafe.status)">{{ selectedSafe.status }}</span>
        </div>

        <div class="detail-card">
          <div class="detail-title">{{ selectedSafe.name }}</div>
          <div class="detail-meta">
            <span class="badge muted">版本 {{ selectedSafe.version }}</span>
            <span class="badge">{{ selectedSafe.usage }}</span>
          </div>
          <div class="detail-info">
            <div class="kv">
              <span>类型</span>
              <span>{{ selectedSafe.type }}</span>
            </div>
            <div class="kv">
              <span>镜像</span>
              <span>{{ selectedSafe.image }}</span>
            </div>
          </div>
        </div>

        <div class="panel">
          <div class="panel-header">
            <div>
              <div class="panel-title">兼容性</div>
              <div class="panel-sub">已验证的机型与挂载点</div>
            </div>
          </div>
          <div class="chip-row">
            <span v-for="item in selectedSafe.aircraft" :key="item" class="chip">{{ item }}</span>
            <div v-if="!selectedSafe.aircraft.length" class="empty-state">无机型列表</div>
          </div>
          <div class="chip-row">
            <span v-for="item in selectedSafe.hardpoints" :key="item" class="chip">{{ item }}</span>
            <div v-if="!selectedSafe.hardpoints.length" class="empty-state">无挂载点列表</div>
          </div>
        </div>
      </div>
    </section>
  </section>
</template>
