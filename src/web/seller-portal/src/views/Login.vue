<template>
  <div class="login-wrap">
    <div class="login-box">
      <h1>Seller Login</h1>
      <p v-if="error" class="error">{{ error }}</p>
      <form @submit.prevent="submit">
        <div class="field"><label>Email</label><input type="email" v-model="form.email" required /></div>
        <div class="field"><label>Password</label><input type="password" v-model="form.password" required /></div>
        <button type="submit" :disabled="loading">{{ loading ? 'Logging in...' : 'Login' }}</button>
      </form>
    </div>
  </div>
</template>
<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { api } from '../api/client'
const auth = useAuthStore(); const router = useRouter()
const form = ref({ email: '', password: '' }); const loading = ref(false); const error = ref('')
async function submit() {
  loading.value = true; error.value = ''
  try { const { token } = await api.post<{ token: string }>('/seller/auth/login', form.value); auth.setAuth(token, form.value.email); router.push('/') }
  catch { error.value = 'Invalid credentials' } finally { loading.value = false }
}
</script>
<style scoped>
.login-wrap { min-height: 100vh; display: flex; align-items: center; justify-content: center; background: #f9fafb; }
.login-box { background: #fff; border: 1px solid #e5e7eb; border-radius: 0.5rem; padding: 2rem; width: 360px; }
h1 { font-size: 1.5rem; font-weight: 700; margin-bottom: 1.5rem; }
.field { margin-bottom: 1rem; } .field label { display: block; font-size: 0.875rem; font-weight: 500; margin-bottom: 0.25rem; }
.field input { width: 100%; border: 1px solid #d1d5db; border-radius: 0.375rem; padding: 0.5rem 0.75rem; }
button { width: 100%; background: #111; color: #fff; border: none; border-radius: 0.375rem; padding: 0.75rem; font-weight: 500; cursor: pointer; }
.error { color: #dc2626; font-size: 0.875rem; margin-bottom: 1rem; }
</style>
