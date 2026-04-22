<template>
  <div class="layout">
    <aside class="sidebar">
      <div class="sidebar-logo">ShopOS Seller</div>
      <nav>
        <RouterLink v-for="l in links" :key="l.to" :to="l.to" class="nav-link" active-class="nav-link--active">{{ l.label }}</RouterLink>
      </nav>
      <button class="logout-btn" @click="auth.logout">Logout</button>
    </aside>
    <div class="content-wrap">
      <header class="top-bar">
        <span class="seller-email">{{ auth.email }}</span>
      </header>
      <main class="main"><RouterView /></main>
    </div>
  </div>
</template>
<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
import { useAuthStore } from '../stores/auth'
const auth = useAuthStore()
const links = [
  { to: '/',          label: 'Dashboard' },
  { to: '/listings',  label: 'Listings'  },
  { to: '/orders',    label: 'Orders'    },
  { to: '/analytics', label: 'Analytics' },
  { to: '/payouts',   label: 'Payouts'   },
  { to: '/profile',   label: 'Profile'   },
]
</script>
<style scoped>
.layout { display: flex; height: 100vh; }
.sidebar { width: 220px; background: #111; color: #fff; display: flex; flex-direction: column; flex-shrink: 0; }
.sidebar-logo { padding: 1.5rem; font-weight: 700; font-size: 1.1rem; border-bottom: 1px solid #374151; }
.nav-link { display: block; padding: 0.625rem 1.5rem; color: #9ca3af; font-size: 0.875rem; }
.nav-link--active { color: #fff; background: #374151; }
.logout-btn { margin-top: auto; padding: 1rem 1.5rem; background: none; border: none; color: #9ca3af; cursor: pointer; text-align: left; font-size: 0.875rem; }
.content-wrap { flex: 1; display: flex; flex-direction: column; overflow: hidden; }
.top-bar { height: 60px; background: #fff; border-bottom: 1px solid #e5e7eb; display: flex; align-items: center; padding: 0 1.5rem; flex-shrink: 0; }
.seller-email { color: #6b7280; font-size: 0.875rem; margin-left: auto; }
.main { flex: 1; overflow: auto; padding: 1.5rem; }
</style>
