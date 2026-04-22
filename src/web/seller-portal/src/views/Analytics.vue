<template>
  <div>
    <h1 class="title">Analytics</h1>
    <div class="grid">
      <div v-for="t in titles" :key="t" class="chart-card">
        <h2>{{ t }}</h2>
        <div class="chart-placeholder">Connect to analytics-service</div>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
const titles = ['Revenue Over Time', 'Top Listings', 'Orders by Day', 'Conversion Rate']
</script>
<style scoped>
.title { font-size: 1.5rem; font-weight: 700; margin-bottom: 1.5rem; }
.grid { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
.chart-card { background: #fff; border: 1px solid #e5e7eb; border-radius: 0.5rem; padding: 1.5rem; }
.chart-card h2 { font-size: 0.875rem; font-weight: 600; color: #374151; margin-bottom: 1rem; }
.chart-placeholder { height: 160px; background: #f9fafb; border-radius: 0.25rem; display: flex; align-items: center; justify-content: center; color: #9ca3af; font-size: 0.875rem; }
</style>
