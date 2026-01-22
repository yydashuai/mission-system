import { defineStore } from 'pinia'

export const useUiStore = defineStore('ui', {
  state: () => ({
    navItems: [
      { label: 'Dashboard', path: '/' },
      { label: 'Missions', path: '/missions' },
      { label: 'Stages', path: '/stages' },
      { label: 'FlightTasks', path: '/flighttasks' },
      { label: 'Weapons', path: '/weapons' },
      { label: 'Cluster', path: '/cluster' },
      { label: 'Settings', path: '/settings' },
    ],
  }),
})
