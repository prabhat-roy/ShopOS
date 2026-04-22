<template>
  <div>
    <h1 class="title">Dashboard</h1>
    <div class="stats-grid">
      <StatCard label="Total Revenue"   :value="stats?.totalRevenue ?? 0"   prefix="$" />
      <StatCard label="Total Sales"     :value="stats?.totalSales ?? 0" />
      <StatCard label="Pending Orders"  :value="stats?.pendingOrders ?? 0" />
      <StatCard label="Active Listings" :value="stats?.activeListings ?? 0" />
    </div>
  </div>
</template>
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import StatCard from '../components/StatCard.vue'
import { api } from '../api/client'
import type { Stats } from '../types'
const stats = ref<Stats | null>(null)
onMounted(() => api.get<Stats>('/seller/stats').then(d => stats.value = d).catch(() => {}))
</script>
<style scoped>
.title { font-size: 1.5rem; font-weight: 700; margin-bottom: 1.5rem; }
.stats-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 1rem; }
</style>
