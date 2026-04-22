import { create } from 'zustand'

interface Product { id: string; name: string; price: number; imageUrl: string }
interface CartItem { product: Product; quantity: number }

interface CartStore {
  items: CartItem[]
  add: (product: Product) => void
  remove: (id: string) => void
  update: (id: string, qty: number) => void
  clear: () => void
  total: () => number
  count: () => number
}

export const useCart = create<CartStore>((set, get) => ({
  items: [],
  add: (product) => set(s => {
    const ex = s.items.find(i => i.product.id === product.id)
    return ex
      ? { items: s.items.map(i => i.product.id === product.id ? { ...i, quantity: i.quantity + 1 } : i) }
      : { items: [...s.items, { product, quantity: 1 }] }
  }),
  remove: (id) => set(s => ({ items: s.items.filter(i => i.product.id !== id) })),
  update: (id, qty) => set(s => ({
    items: qty <= 0 ? s.items.filter(i => i.product.id !== id) : s.items.map(i => i.product.id === id ? { ...i, quantity: qty } : i)
  })),
  clear: () => set({ items: [] }),
  total: () => get().items.reduce((s, i) => s + i.product.price * i.quantity, 0),
  count: () => get().items.reduce((s, i) => s + i.quantity, 0),
}))
