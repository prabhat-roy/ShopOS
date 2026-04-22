<template>
  <div>
    <h1 class="title">Payouts</h1>
    <div class="table-wrap">
      <p v-if="loading" class="empty">Loading...</p>
      <table v-else-if="payouts.length">
        <thead><tr><th>Reference</th><th>Amount</th><th>Status</th><th>Date</th></tr></thead>
        <tbody>
          <tr v-for="p in payouts" :key="p.id">
            <td>{{ p.reference }}</td>
            <td>${{ p.amount.toFixed(2) }}</td>
            <td><span :class="['badge', p.status === 'completed' ? 'badge-green' : 'badge-yellow']">{{ p.status }}</span></td>
            <td>{{ new Date(p.date).toLocaleDateString() }}</td>
          </tr>
        </tbody>
      </table>
      <p v-else class="empty">No payouts yet.</p>
    </div>
  </div>
</template>
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../api/client'
import type { Payout } from '../types'
const payouts = ref<Payout[]>([]); const loading = ref(true)
onMounted(() => api.get<Payout[]>('/seller/payouts').then(d => payouts.value = d).catch(() => {}).finally(() => loading.value = false))
</script>
<style scoped>
.title { font-size: 1.5rem; font-weight: 700; margin-bottom: 1.5rem; }
.table-wrap { background: #fff; border: 1px solid #e5e7eb; border-radius: 0.5rem; overflow: hidden; }
table { width: 100%; border-collapse: collapse; font-size: 0.875rem; }
th { text-align: left; padding: 0.75rem 1rem; font-weight: 600; border-bottom: 2px solid #e5e7eb; }
td { padding: 0.75rem 1rem; border-bottom: 1px solid #f3f4f6; }
.badge { padding: 0.2rem 0.5rem; border-radius: 9999px; font-size: 0.75rem; font-weight: 600; }
.badge-green { background: #d1fae5; color: #065f46; } .badge-yellow { background: #fef3c7; color: #92400e; }
.empty { text-align: center; padding: 2rem; color: #9ca3af; }
</style>
