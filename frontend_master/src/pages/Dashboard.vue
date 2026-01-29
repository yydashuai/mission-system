<script setup>
import { computed } from 'vue'
import { storeToRefs } from 'pinia'
import { useDataStore } from '../stores/data'
import { useSystemStore } from '../stores/system'

const dataStore = useDataStore()
const systemStore = useSystemStore()
const { missions, stages, flightTasks, weapons } = storeToRefs(dataStore)

const totalMissions = computed(() => missions.value.length)
const activeMissions = computed(() => missions.value.filter((item) => (
  item.status !== '已完成' && item.status !== '失败'
)))
const runningStages = computed(() => stages.value.filter((item) => item.status === '运行中'))
const waitingTasks = computed(() => flightTasks.value.filter((item) => (
  item.status === '待执行' || item.status === '已调度'
)))
const readyWeapons = computed(() => weapons.value.filter((item) => item.status === '可用'))

const kpis = computed(() => ([
  {
    label: '活跃任务',
    value: String(activeMissions.value.length),
    delta: `${totalMissions.value} 个跟踪`,
    tone: activeMissions.value.length ? 'ok' : 'muted',
  },
  {
    label: '运行中阶段',
    value: String(runningStages.value.length),
    delta: `${stages.value.length} 个总计`,
    tone: runningStages.value.length ? 'ok' : 'muted',
  },
  {
    label: '飞行任务',
    value: String(flightTasks.value.length),
    delta: `${waitingTasks.value.length} 个等待`,
    tone: waitingTasks.value.length ? 'warn' : 'ok',
  },
  {
    label: '就绪武器',
    value: String(readyWeapons.value.length),
    delta: `${weapons.value.length} 个注册`,
    tone: readyWeapons.value.length ? 'ok' : 'muted',
  },
]))

const heroTags = computed(() => {
  const apiStatus = systemStore.apiStatus
  const apiTag = (() => {
    if (apiStatus === 'ok') return { label: 'API 已连接', tone: 'ok' }
    if (apiStatus === 'checking') return { label: 'API 检查中', tone: 'warn' }
    if (apiStatus === 'down') return { label: 'API 断开', tone: 'err' }
    if (apiStatus === 'disabled') return { label: 'API 已禁用', tone: 'muted' }
    return { label: 'API 未知', tone: 'muted' }
  })()

  const waiting = waitingTasks.value.length
  const tasksTag = waiting
    ? { label: `${waiting} 个任务等待`, tone: 'warn' }
    : { label: '无等待任务', tone: 'ok' }

  return [apiTag, tasksTag]
})

const activeStage = computed(() => {
  const list = stages.value
  if (!list.length) return null
  const pick = (status) => list.find((item) => item.status === status) || null
  return pick('运行中') || pick('已调度') || pick('待执行') || list[0]
})

const stageStatus = computed(() => String(activeStage.value?.status || '').toLowerCase())
const stageTasks = computed(() => activeStage.value?.tasks || [])
const activeTaskCount = computed(() => stageTasks.value.filter((item) => (
  item.status === '运行中' || item.status === '已调度'
)).length)

const timelineClass = computed(() => {
  if (!activeStage.value) return { init: '', tasks: '', complete: '' }
  const isSucceeded = stageStatus.value === 'succeeded'
  const isFailed = stageStatus.value === 'failed'
  const isRunning = stageStatus.value === 'running'
  const isScheduled = stageStatus.value === 'scheduled'
  const isPending = stageStatus.value === 'pending'

  return {
    init: isPending ? 'active' : 'done',
    tasks: (isRunning || isScheduled) ? 'active' : (isSucceeded || isFailed ? 'done' : ''),
    complete: (isSucceeded || isFailed) ? 'done' : '',
  }
})

const stageMeta = computed(() => ({
  mission: activeStage.value?.mission || '--',
  mode: activeStage.value?.mode || '--',
  name: activeStage.value?.name || '--',
  tasksTotal: stageTasks.value.length,
}))

const stageSummary = computed(() => {
  if (!activeStage.value) return '--'
  if (stageStatus.value === 'succeeded') return '已完成'
  if (stageStatus.value === 'failed') return '已失败'
  if (stageStatus.value === 'running') return `${activeTaskCount.value} 个活跃`
  if (stageStatus.value === 'scheduled') return '已调度'
  if (stageStatus.value === 'pending') return '等待中'
  return '--'
})

const parseTimeScore = (value) => {
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

const queueTasks = computed(() => {
  const weight = {
    pending: 0,
    scheduled: 1,
    running: 2,
    succeeded: 3,
    failed: 4,
  }
  const list = [...flightTasks.value]
  list.sort((a, b) => {
    const aKey = weight[String(a.status || '').toLowerCase()] ?? 9
    const bKey = weight[String(b.status || '').toLowerCase()] ?? 9
    if (aKey !== bKey) return aKey - bKey
    return parseTimeScore(b.scheduledAt) - parseTimeScore(a.scheduledAt)
  })
  return list.slice(0, 5)
})
</script>

<template>
  <section class="page">
    <section class="hero">
      <div>
        <div class="hero-title">作战概览</div>
        <div class="hero-sub">
          任务链状态、调度压力和部署健康状况。
        </div>
      </div>
      <div class="hero-tags">
        <span
          v-for="tag in heroTags"
          :key="tag.label"
          class="badge"
          :class="tag.tone"
        >
          {{ tag.label }}
        </span>
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
            <div class="panel-title">活跃阶段时间线</div>
            <div class="panel-sub">{{ stageMeta.mission }} / {{ stageMeta.name }}</div>
          </div>
          <span class="badge">{{ stageMeta.mode }}</span>
        </div>
        <div class="timeline">
          <div class="timeline-step" :class="timelineClass.init">
            <span class="dot"></span>
            <div>
              <div class="step-title">阶段初始化</div>
              <div class="step-meta">{{ stageSummary }}</div>
            </div>
          </div>
          <div class="timeline-step" :class="timelineClass.tasks">
            <span class="dot"></span>
            <div>
              <div class="step-title">飞行任务调度</div>
              <div class="step-meta">{{ activeTaskCount }} 个活跃 · {{ stageMeta.tasksTotal }} 个总计</div>
            </div>
          </div>
          <div class="timeline-step" :class="timelineClass.complete">
            <span class="dot"></span>
            <div>
              <div class="step-title">阶段完成</div>
              <div class="step-meta">
                {{ stageStatus === 'succeeded' || stageStatus === 'failed' ? stageStatus : '等待中' }}
              </div>
            </div>
          </div>
          <div v-if="!activeStage" class="empty-state">无阶段数据。</div>
        </div>
      </div>

      <div class="panel">
        <div class="panel-header">
          <div>
            <div class="panel-title">飞行任务队列</div>
            <div class="panel-sub">按阶段的调度压力</div>
          </div>
          <button class="ghost small">查看全部</button>
        </div>
        <div class="task-table">
          <div class="task-row task-head">
            <div class="cell-start">名称</div>
            <div class="cell-center">状态</div>
            <div class="cell-center">预计时间</div>
            <div class="cell-center">节点</div>
          </div>
          <div v-for="task in queueTasks" :key="task.name" class="task-row">
            <div class="cell-start">
              <span class="task-name">{{ task.name }}</span>
            </div>
            <div class="cell-center">
              <span class="badge" :class="String(task.status).toLowerCase()">{{ task.status }}</span>
            </div>
            <div class="cell-center">
              <span>{{ task.scheduledAt || '--' }}</span>
            </div>
            <div class="cell-center">
              <span class="muted">{{ task.node || '--' }}</span>
            </div>
          </div>
          <div v-if="!queueTasks.length" class="empty-state">无飞行任务。</div>
        </div>
      </div>
    </section>
  </section>
</template>
