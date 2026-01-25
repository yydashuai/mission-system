<script setup>
import { onBeforeUnmount, onMounted } from 'vue'
import TopBar from './components/layout/TopBar.vue'
import SideBar from './components/layout/SideBar.vue'
import DetailPanel from './components/layout/DetailPanel.vue'
import { useSystemStore } from './stores/system'
import { useDataStore } from './stores/data'

const systemStore = useSystemStore()
const dataStore = useDataStore()

onMounted(() => {
  systemStore.init()
  systemStore.startClock()
  systemStore.startApiPolling()
  dataStore.refreshAll()
  dataStore.startPolling()
})

onBeforeUnmount(() => {
  systemStore.stopClock()
  systemStore.stopApiPolling()
  dataStore.stopPolling()
})
</script>

<template>
  <div class="app-shell">
    <TopBar />
    <div class="app-body">
      <SideBar />
      <main class="main">
        <router-view />
      </main>
      <DetailPanel />
    </div>
  </div>
</template>
