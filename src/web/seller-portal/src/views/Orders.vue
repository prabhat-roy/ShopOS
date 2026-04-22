<template>
  <div>
    <h1 class="title">Orders</h1>
    <div class="table-wrap">
      <p v-if="loading" class="empty">Loading...</p>
      <table v-else-if="orders.length">
        <thead><tr><th>Order ID</th><th>Buyer</th><th>Items</th><th>Total</th><th>Status</th><th>Date</th></tr></thead>
        <tbody>
          <tr v-for="o in orders" :key="o.id">
            <td>#{{ o.id.slice(0,8).toUpperCase() }}</td><td>{{ o.buyerEmail }}</td><td>{{ o.items }}</td>
            <td>${{ o.total.toFixed(2) }}</td><td>{{ o.status }}</td>
            <td>{{ new Date(o.createdAt).toLocaleDateString() }}</td>
          </tr>
        </tbody>
      </table>
      <p v-else class="empty">No orders yet.</p>
    </div>
  </div>
</template>
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../api/client'
import type { Order } from '../types'
const orders = ref<Order[]>([]); const loading = ref(true)
onMounted(() => api.get<Order[]>('/seller/orders').then(d => orders.value = d).catch(() => {}).finally(() => loading.value = false))
</script>
<style scoped>
.title { font-size: 1.5rem; font-weight: 700; margin-bottom: 1.5rem; }
.table-wrap { background: #fff; border: 1px solid #e5e7eb; border-radius: 0.5rem; overflow: hidden; }
table { width: 100%; border-collapse: collapse; font-size: 0.875rem; }
th { text-align: left; padding: 0.75rem 1rem; font-weight: 600; border-bottom: 2px solid #e5e7eb; }
td { padding: 0.75rem 1rem; border-bottom: 1px solid #f3f4f6; }
.empty { text-align: center; padding: 2rem; color: #9ca3af; }
</style>
