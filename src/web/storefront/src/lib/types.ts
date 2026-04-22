export interface Product {
  id: string
  name: string
  slug: string
  price: number
  currency: string
  imageUrl: string
  category: string
  rating: number
  reviewCount: number
  inStock: boolean
  description?: string
}

export interface CartItem {
  product: Product
  quantity: number
}

export interface Order {
  id: string
  status: 'pending' | 'confirmed' | 'shipped' | 'delivered' | 'cancelled'
  createdAt: string
  total: number
  items: CartItem[]
}

export interface User {
  id: string
  email: string
  firstName: string
  lastName: string
}
