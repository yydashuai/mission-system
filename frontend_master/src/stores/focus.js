import { defineStore } from 'pinia'

export const useFocusStore = defineStore('focus', {
  state: () => ({
    kind: 'flighttask',
    item: {
      name: 'ft-alpha-9',
      status: '已调度',
      mission: 'SeaStrike-02',
      stage: 'Strike-Phase',
      node: 'worker-1',
      weapon: 'PL-15',
      attempts: 2,
      scheduledAt: '08:14',
      podStatus: '待执行',
    },
    events: [
      { time: '08:21', label: 'FT-ghost-12 待调度: 节点亲和性不满足', tone: 'warn' },
      { time: '08:18', label: '武器 PL-15 已注入 ft-rapid-3', tone: 'ok' },
      { time: '08:07', label: '任务 SeaStrike-02 阶段-2 运行中', tone: 'ok' },
      { time: '07:55', label: 'ft-alpha-9 镜像拉取失败', tone: 'err' },
    ],
  }),
  actions: {
    setFocus(kind, item) {
      this.kind = kind
      this.item = item
    },
  },
})
