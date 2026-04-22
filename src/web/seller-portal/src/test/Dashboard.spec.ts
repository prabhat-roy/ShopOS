import { mount } from '@vue/test-utils'
import { describe, it, expect, vi } from 'vitest'
import { createPinia } from 'pinia'
import Dashboard from '../views/Dashboard.vue'

vi.mock('../api/client', () => ({
  api: { get: vi.fn().mockResolvedValue({ totalSales: 100, totalRevenue: 5000, pendingOrders: 3, activeListings: 20 }) },
}))

describe('Dashboard', () => {
  it('renders title', () => {
    const wrapper = mount(Dashboard, { global: { plugins: [createPinia()] } })
    expect(wrapper.text()).toContain('Dashboard')
  })
})
