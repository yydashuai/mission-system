<script setup>
defineProps({
  title: {
    type: String,
    required: true,
  },
  subtitle: {
    type: String,
    default: '',
  },
})
</script>

<template>
  <section class="page">
    <div class="panel">
      <div class="panel-title">{{ title }}</div>
      <div class="panel-sub">{{ subtitle }}</div>
      <div class="placeholder-body">
        Live data wiring pending. This page will render Kubernetes resources once the
        API is connected.
      </div>
    </div>
  </section>
</template>
