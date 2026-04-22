import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('seller_token'))
  const email = ref<string | null>(localStorage.getItem('seller_email'))

  function setAuth(t: string, e: string) {
    token.value = t; email.value = e
    localStorage.setItem('seller_token', t)
    localStorage.setItem('seller_email', e)
  }
  function logout() {
    token.value = null; email.value = null
    localStorage.removeItem('seller_token'); localStorage.removeItem('seller_email')
    window.location.href = '/login'
  }
  return { token, email, setAuth, logout }
})
