<template>
  <div style="max-width: 600px">
    <h1 class="title">Store Profile</h1>
    <div class="card">
      <div v-for="field in fields" :key="field.label" class="field">
        <label>{{ field.label }}</label>
        <input v-model="form[field.key]" :type="field.type ?? 'text'" />
      </div>
      <button class="btn-save" @click="save">{{ saved ? 'Saved!' : 'Save Changes' }}</button>
    </div>
  </div>
</template>
<script setup lang="ts">
import { ref } from 'vue'
const form = ref({ storeName: '', description: '', email: '', website: '', country: '' })
const saved = ref(false)
const fields = [
  { label: 'Store Name', key: 'storeName' as const },
  { label: 'Description', key: 'description' as const },
  { label: 'Contact Email', key: 'email' as const, type: 'email' },
  { label: 'Website', key: 'website' as const },
  { label: 'Country', key: 'country' as const },
]
function save() { saved.value = true; setTimeout(() => saved.value = false, 2000) }
</script>
<style scoped>
.title { font-size: 1.5rem; font-weight: 700; margin-bottom: 1.5rem; }
.card { background: #fff; border: 1px solid #e5e7eb; border-radius: 0.5rem; padding: 1.5rem; }
.field { margin-bottom: 1rem; }
.field label { display: block; font-size: 0.875rem; font-weight: 500; margin-bottom: 0.25rem; }
.field input { width: 100%; border: 1px solid #d1d5db; border-radius: 0.375rem; padding: 0.5rem 0.75rem; }
.btn-save { background: #111; color: #fff; border: none; border-radius: 0.375rem; padding: 0.5rem 1.5rem; cursor: pointer; }
</style>
