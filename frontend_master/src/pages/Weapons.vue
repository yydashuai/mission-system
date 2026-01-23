<script setup>
import { ref, watch } from 'vue'
import { useFocusStore } from '../stores/focus'

const weapons = [
  {
    name: 'PL-15',
    status: 'Available',
    image: 'registry.airforce/weapon/pl-15:1.4.2',
    version: '1.4.2',
    usage: '18 active',
    aircraft: ['J-20', 'J-16'],
    hardpoints: ['hp-2', 'hp-4'],
    resources: 'cpu 200m / mem 128Mi',
  },
  {
    name: 'PL-10',
    status: 'Available',
    image: 'registry.airforce/weapon/pl-10:2.1.0',
    version: '2.1.0',
    usage: '9 active',
    aircraft: ['J-10', 'J-11', 'J-16'],
    hardpoints: ['hp-1', 'hp-2'],
    resources: 'cpu 150m / mem 96Mi',
  },
  {
    name: 'YJ-12',
    status: 'Degraded',
    image: 'registry.airforce/weapon/yj-12:0.9.8',
    version: '0.9.8',
    usage: '2 active',
    aircraft: ['H-6K'],
    hardpoints: ['hp-6'],
    resources: 'cpu 300m / mem 256Mi',
  },
  {
    name: 'PL-5E',
    status: 'Retired',
    image: 'registry.airforce/weapon/pl-5e:legacy',
    version: 'legacy',
    usage: '0 active',
    aircraft: ['J-7'],
    hardpoints: ['hp-1'],
    resources: 'cpu 80m / mem 64Mi',
  },
]

const selected = ref(weapons[0])
const focusStore = useFocusStore()

watch(selected, (value) => {
  focusStore.setFocus('weapon', value)
}, { immediate: true })

const statusTone = (status) => {
  if (status === 'Available') return 'ok'
  if (status === 'Degraded') return 'warn'
  if (status === 'Retired') return 'muted'
  return 'muted'
}
</script>

<template>
  <section class="page">
    <header class="page-header">
      <div>
        <div class="page-title">Weapons</div>
        <div class="page-sub">Compatibility, container spec, and usage.</div>
      </div>
      <div class="page-actions">
        <button class="ghost small">Pull Image</button>
        <button class="ghost small">Compatibility Matrix</button>
      </div>
    </header>

    <section class="split-grid">
      <div class="panel">
        <div class="panel-header">
          <div>
            <div class="panel-title">Weapon Registry</div>
            <div class="panel-sub">Sidecar packages available to FlightTasks.</div>
          </div>
          <span class="badge">{{ weapons.length }} packages</span>
        </div>
        <div class="data-table">
          <div class="data-row is-head" style="--cols: 1.1fr 0.8fr 0.9fr 0.7fr 0.7fr">
            <span>Weapon</span>
            <span>Status</span>
            <span>Image</span>
            <span>Version</span>
            <span>Usage</span>
          </div>
          <button
            v-for="weapon in weapons"
            :key="weapon.name"
            class="data-row is-selectable"
            :class="{ active: selected.name === weapon.name }"
            style="--cols: 1.1fr 0.8fr 0.9fr 0.7fr 0.7fr"
            type="button"
            @click="selected = weapon"
          >
            <div class="cell-main">
              <div class="cell-title">{{ weapon.name }}</div>
              <div class="cell-sub">{{ weapon.resources }}</div>
            </div>
            <span class="badge" :class="statusTone(weapon.status)">{{ weapon.status }}</span>
            <span class="muted">{{ weapon.image }}</span>
            <span class="muted">{{ weapon.version }}</span>
            <span class="badge muted">{{ weapon.usage }}</span>
          </button>
        </div>
      </div>

      <div class="panel">
        <div class="panel-header">
          <div>
            <div class="panel-title">Weapon Detail</div>
            <div class="panel-sub">{{ selected.image }}</div>
          </div>
          <span class="badge" :class="statusTone(selected.status)">{{ selected.status }}</span>
        </div>

        <div class="detail-card">
          <div class="detail-title">{{ selected.name }}</div>
          <div class="detail-meta">
            <span class="badge muted">Version {{ selected.version }}</span>
            <span class="badge">{{ selected.usage }}</span>
          </div>
          <div class="detail-info">
            <div class="kv">
              <span>Resources</span>
              <span>{{ selected.resources }}</span>
            </div>
            <div class="kv">
              <span>Image</span>
              <span>{{ selected.image }}</span>
            </div>
          </div>
        </div>

        <div class="panel">
          <div class="panel-header">
            <div>
              <div class="panel-title">Compatibility</div>
              <div class="panel-sub">Verified aircraft and hardpoints.</div>
            </div>
          </div>
          <div class="chip-row">
            <span v-for="item in selected.aircraft" :key="item" class="chip">{{ item }}</span>
          </div>
          <div class="chip-row">
            <span v-for="item in selected.hardpoints" :key="item" class="chip">{{ item }}</span>
          </div>
        </div>
      </div>
    </section>
  </section>
</template>
