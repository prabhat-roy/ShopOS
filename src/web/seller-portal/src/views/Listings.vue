<template>
  <div>
    <div class="header-row">
      <h1 class="title">My Listings</h1>
      <button class="btn-add">+ New Listing</button>
    </div>
    <div class="table-wrap">
      <table v-if="listings.length">
        <thead><tr><th>Name</th><th>Category</th><th>Price</th><th>Stock</th><th>Sales</th><th>Status</th></tr></thead>
        <tbody>
          <tr v-for="l in listings" :key="l.id">
            <td>{{ l.name }}</td><td>{{ l.category }}</td>
            <td>${{ l.price.toFixed(2) }}</td><td :style="{ color: l.stock < 5 ? '#dc2626' : '#059669' }">{{ l.stock }}</td>
            <td>{{ l.sales }}</td>
            <td><span :class="['badge', l.status === 'active' ? 'badge-green' : 'badge-gray']">{{ l.status }}</span></td>
          </tr>
        </tbody>
      </table>
      <p v-else-if="!loading" class="empty">No listings yet.</p>
    </div>
  </div>
</template>
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../api/client'
import type { Listing } from '../types'
const listings = ref<Listing[]>([]); const loading = ref(true)
onMounted(() => api.get<Listing[]>('/seller/listings').then(d => listings.value = d).catch(() => {}).finally(() => loading.value = false))
</script>
<style scoped>
.title { font-size: 1.5rem; font-weight: 700; } .header-row { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem; }
.btn-add { background: #111; color: #fff; border: none; border-radius: 0.375rem; padding: 0.5rem 1rem; cursor: pointer; font-size: 0.875rem; }
.table-wrap { background: #fff; border: 1px solid #e5e7eb; border-radius: 0.5rem; overflow: hidden; }
table { width: 100%; border-collapse: collapse; font-size: 0.875rem; }
th { text-align: left; padding: 0.75rem 1rem; font-weight: 600; border-bottom: 2px solid #e5e7eb; }
td { padding: 0.75rem 1rem; border-bottom: 1px solid #f3f4f6; }
.badge { padding: 0.2rem 0.5rem; border-radius: 9999px; font-size: 0.75rem; font-weight: 600; }
.badge-green { background: #d1fae5; color: #065f46; } .badge-gray { background: #f3f4f6; color: #374151; }
.empty { text-align: center; padding: 2rem; color: #9ca3af; }
</style>
