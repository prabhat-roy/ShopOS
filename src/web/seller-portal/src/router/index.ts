import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const routes = [
  { path: '/login', component: () => import('../views/Login.vue'), meta: { public: true } },
  {
    path: '/',
    component: () => import('../components/AppLayout.vue'),
    children: [
      { path: '',          component: () => import('../views/Dashboard.vue') },
      { path: 'listings',  component: () => import('../views/Listings.vue') },
      { path: 'orders',    component: () => import('../views/Orders.vue') },
      { path: 'analytics', component: () => import('../views/Analytics.vue') },
      { path: 'payouts',   component: () => import('../views/Payouts.vue') },
      { path: 'profile',   component: () => import('../views/Profile.vue') },
    ],
  },
]

const router = createRouter({ history: createWebHistory(), routes })

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (!to.meta.public && !auth.token) return '/login'
})

export default router
