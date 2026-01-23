<script setup>
import { computed, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useFocusStore } from '../stores/focus'
import { useDataStore } from '../stores/data'

const dataStore = useDataStore()
const { weapons } = storeToRefs(dataStore)
const isLoading = computed(() => dataStore.loading.weapons)

const selectedKey = ref('')
const focusStore = useFocusStore()

const emptyWeapon = {
  name: '--',
  status: '--',
  image: '--',
  version: '--',
  usage: '--',
  aircraft: [],
  hardpoints: [],
  resources: '--',
}

const selected = computed(() => (
  weapons.value.find((item) => item.name === selectedKey.value) || null
))

const selectedSafe = computed(() => selected.value || emptyWeapon)

watch(weapons, (list) => {
  if (!list.length) return
  const exists = list.some((item) => item.name === selectedKey.value)
  if (!exists) {
    selectedKey.value = list[0].name
  }
}, { immediate: true })

watch(selected, (value) => {
  if (value) {
    focusStore.setFocus('weapon', value)
  }
}, { immediate: true })

const statusTone = (status) => {
  if (status === 'Available') return 'ok'
  if (status === 'Updating' || status === 'Degraded') return 'warn'
  if (status === 'Deprecated' || status === 'Retired') return 'muted'
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
          <span v-if="isLoading" class="badge warn">Loading</span>
          <span v-else class="badge">{{ weapons.length }} packages</span>
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
            :class="{ active: selectedKey === weapon.name }"
            style="--cols: 1.1fr 0.8fr 0.9fr 0.7fr 0.7fr"
            type="button"
            @click="selectedKey = weapon.name"
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
          <div v-if="!weapons.length && !isLoading" class="empty-state">No weapons available.</div>
        </div>
      </div>

      <div class="panel">
        <div class="panel-header">
          <div>
            <div class="panel-title">Weapon Detail</div>
            <div class="panel-sub">{{ selectedSafe.image }}</div>
          </div>
          <span class="badge" :class="statusTone(selectedSafe.status)">{{ selectedSafe.status }}</span>
        </div>

        <div class="detail-card">
          <div class="detail-title">{{ selectedSafe.name }}</div>
          <div class="detail-meta">
            <span class="badge muted">Version {{ selectedSafe.version }}</span>
            <span class="badge">{{ selectedSafe.usage }}</span>
          </div>
          <div class="detail-info">
            <div class="kv">
              <span>Resources</span>
              <span>{{ selectedSafe.resources }}</span>
            </div>
            <div class="kv">
              <span>Image</span>
              <span>{{ selectedSafe.image }}</span>
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
            <span v-for="item in selectedSafe.aircraft" :key="item" class="chip">{{ item }}</span>
            <div v-if="!selectedSafe.aircraft.length" class="empty-state">No aircraft listed.</div>
          </div>
          <div class="chip-row">
            <span v-for="item in selectedSafe.hardpoints" :key="item" class="chip">{{ item }}</span>
            <div v-if="!selectedSafe.hardpoints.length" class="empty-state">No hardpoints listed.</div>
          </div>
        </div>
      </div>
    </section>
  </section>
</template>
