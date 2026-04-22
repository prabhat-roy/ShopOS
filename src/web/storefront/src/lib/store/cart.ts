'use client'
import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { CartItem, Product } from '../types'

interface CartStore {
  items: CartItem[]
  add: (product: Product, qty?: number) => void
  remove: (productId: string) => void
  update: (productId: string, qty: number) => void
  clear: () => void
  total: () => number
  count: () => number
}

export const useCart = create<CartStore>()(
  persist(
    (set, get) => ({
      items: [],
      add: (product, qty = 1) => set(s => {
        const existing = s.items.find(i => i.product.id === product.id)
        if (existing) {
          return { items: s.items.map(i => i.product.id === product.id ? { ...i, quantity: i.quantity + qty } : i) }
        }
        return { items: [...s.items, { product, quantity: qty }] }
      }),
      remove: (id) => set(s => ({ items: s.items.filter(i => i.product.id !== id) })),
      update: (id, qty) => set(s => ({
        items: qty <= 0
          ? s.items.filter(i => i.product.id !== id)
          : s.items.map(i => i.product.id === id ? { ...i, quantity: qty } : i)
      })),
      clear: () => set({ items: [] }),
      total: () => get().items.reduce((sum, i) => sum + i.product.price * i.quantity, 0),
      count: () => get().items.reduce((sum, i) => sum + i.quantity, 0),
    }),
    { name: 'shopos-cart' }
  )
)
