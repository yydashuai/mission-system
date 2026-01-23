import { defineStore } from 'pinia'

export const useFocusStore = defineStore('focus', {
  state: () => ({
    kind: 'flighttask',
    item: {
      name: 'ft-alpha-9',
      status: 'Scheduled',
      mission: 'SeaStrike-02',
      stage: 'Strike-Phase',
      node: 'worker-1',
      weapon: 'PL-15',
      attempts: 2,
      scheduledAt: '08:14',
      podStatus: 'Pending',
    },
    events: [
      { time: '08:21', label: 'FT-ghost-12 pending: node affinity unmet', tone: 'warn' },
      { time: '08:18', label: 'Weapon PL-15 injected into ft-rapid-3', tone: 'ok' },
      { time: '08:07', label: 'Mission SeaStrike-02 stage-2 running', tone: 'ok' },
      { time: '07:55', label: 'ImagePullBackOff on ft-alpha-9', tone: 'err' },
    ],
  }),
  actions: {
    setFocus(kind, item) {
      this.kind = kind
      this.item = item
    },
  },
})
