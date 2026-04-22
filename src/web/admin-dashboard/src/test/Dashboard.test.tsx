import { render, screen } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { MemoryRouter } from 'react-router-dom'
import Dashboard from '../pages/Dashboard'

vi.mock('../api/client', () => ({
  statsApi: { get: vi.fn().mockResolvedValue({ totalRevenue: 50000, totalOrders: 120, totalUsers: 800, totalProducts: 450 }) },
}))

describe('Dashboard', () => {
  it('renders stat cards', () => {
    render(<MemoryRouter><Dashboard /></MemoryRouter>)
    expect(screen.getByText('Dashboard')).toBeTruthy()
    expect(screen.getByText('Total Revenue')).toBeTruthy()
  })
})
