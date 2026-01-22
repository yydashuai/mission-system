<script setup>
import { useRoute } from 'vue-router'
import { useUiStore } from '../../stores/ui'

const route = useRoute()
const uiStore = useUiStore()

const isActive = (path) => route.path === path
</script>

<template>
  <aside class="sidebar">
    <div class="section-title">Navigation</div>
    <nav class="nav-list">
      <router-link
        v-for="item in uiStore.navItems"
        :key="item.label"
        :to="item.path"
        class="nav-item"
        :class="{ active: isActive(item.path) }"
      >
        <span class="nav-dot"></span>
        {{ item.label }}
      </router-link>
    </nav>

    <div class="section-title">System Health</div>
    <div class="health-grid">
      <div class="health-card">
        <div class="health-title">Nodes Ready</div>
        <div class="health-value">2 / 3</div>
        <div class="health-meta">Master: Ready</div>
      </div>
      <div class="health-card">
        <div class="health-title">Events (24h)</div>
        <div class="health-value">17</div>
        <div class="health-meta">3 warnings</div>
      </div>
    </div>
  </aside>
</template>
