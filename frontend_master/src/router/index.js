import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from '../pages/Dashboard.vue'
import Placeholder from '../pages/Placeholder.vue'

const routes = [
  { path: '/', name: 'dashboard', component: Dashboard },
  {
    path: '/missions',
    name: 'missions',
    component: Placeholder,
    props: {
      title: 'Missions',
      subtitle: 'Mission list, lifecycle, and stage composition.',
    },
  },
  {
    path: '/stages',
    name: 'stages',
    component: Placeholder,
    props: {
      title: 'MissionStages',
      subtitle: 'Stage execution flow and dependencies.',
    },
  },
  {
    path: '/flighttasks',
    name: 'flighttasks',
    component: Placeholder,
    props: {
      title: 'FlightTasks',
      subtitle: 'Scheduling detail, pod binding, and status.',
    },
  },
  {
    path: '/weapons',
    name: 'weapons',
    component: Placeholder,
    props: {
      title: 'Weapons',
      subtitle: 'Compatibility, container spec, and usage.',
    },
  },
  {
    path: '/cluster',
    name: 'cluster',
    component: Placeholder,
    props: {
      title: 'Cluster',
      subtitle: 'Nodes, pods, events, and basic health.',
    },
  },
  {
    path: '/settings',
    name: 'settings',
    component: Placeholder,
    props: {
      title: 'Settings',
      subtitle: 'API connectivity, refresh, and access.',
    },
  },
  { path: '/:pathMatch(.*)*', redirect: '/' },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

export default router
