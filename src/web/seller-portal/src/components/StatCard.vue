<template>
  <div class="card">
    <p class="label">{{ label }}</p>
    <p class="value">{{ prefix }}{{ formattedValue }}</p>
    <p v-if="change !== undefined" :class="['change', change >= 0 ? 'pos' : 'neg']">
      {{ change >= 0 ? '+' : '' }}{{ change.toFixed(1) }}% vs last month
    </p>
  </div>
</template>
<script setup lang="ts">
import { computed } from 'vue'
const props = defineProps<{ label: string; value: number; prefix?: string; change?: number }>()
const formattedValue = computed(() => props.prefix === '$' ? props.value.toLocaleString('en-US', { minimumFractionDigits: 2 }) : props.value)
</script>
<style scoped>
.card { background: #fff; border: 1px solid #e5e7eb; border-radius: 0.5rem; padding: 1.5rem; }
.label { font-size: 0.75rem; color: #6b7280; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 0.5rem; }
.value { font-size: 1.875rem; font-weight: 700; }
.change { font-size: 0.75rem; margin-top: 0.25rem; }
.pos { color: #059669; } .neg { color: #dc2626; }
</style>
