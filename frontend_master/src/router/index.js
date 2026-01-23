import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from '../pages/Dashboard.vue'
import Missions from '../pages/Missions.vue'
import Stages from '../pages/Stages.vue'
import FlightTasks from '../pages/FlightTasks.vue'
import Weapons from '../pages/Weapons.vue'
import Cluster from '../pages/Cluster.vue'
import Settings from '../pages/Settings.vue'

const routes = [
  { path: '/', name: 'dashboard', component: Dashboard },
  {
    path: '/missions',
    name: 'missions',
    component: Missions,
  },
  {
    path: '/stages',
    name: 'stages',
    component: Stages,
  },
  {
    path: '/flighttasks',
    name: 'flighttasks',
    component: FlightTasks,
  },
  {
    path: '/weapons',
    name: 'weapons',
    component: Weapons,
  },
  {
    path: '/cluster',
    name: 'cluster',
    component: Cluster,
  },
  {
    path: '/settings',
    name: 'settings',
    component: Settings,
  },
  { path: '/:pathMatch(.*)*', redirect: '/' },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

export default router
