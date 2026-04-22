export interface Order {
  id: string; status: string; total: number; createdAt: string; customerEmail: string; itemCount: number
}
export interface Product {
  id: string; name: string; price: number; category: string; stock: number; imageUrl: string
}
export interface User {
  id: string; email: string; firstName: string; lastName: string; createdAt: string; status: string
}
export interface DashboardStats {
  totalOrders: number; totalRevenue: number; totalUsers: number; totalProducts: number
  revenueChange: number; ordersChange: number; usersChange: number
}
