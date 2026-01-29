import { defineStore } from 'pinia'

export const useUiStore = defineStore('ui', {
  state: () => ({
    navItems: [
      { label: '仪表盘', path: '/' },
      { label: '任务', path: '/missions' },
      { label: '阶段', path: '/stages' },
      { label: '飞行任务', path: '/flighttasks' },
      { label: '武器', path: '/weapons' },
      { label: '集群', path: '/cluster' },
      { label: '设置', path: '/settings' },
    ],
  }),
})
