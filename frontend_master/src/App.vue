<script setup>
import TopBar from './components/layout/TopBar.vue'
import SideBar from './components/layout/SideBar.vue'
import DetailPanel from './components/layout/DetailPanel.vue'
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
