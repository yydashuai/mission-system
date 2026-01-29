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
        实时数据连接待建立。API 连接后，此页面将显示 Kubernetes 资源。
      </div>
    </div>
  </section>
</template>
