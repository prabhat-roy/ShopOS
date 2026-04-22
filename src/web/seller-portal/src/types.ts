export interface Listing { id: string; name: string; price: number; stock: number; status: string; category: string; sales: number }
export interface Order   { id: string; buyerEmail: string; total: number; status: string; createdAt: string; items: number }
export interface Payout  { id: string; amount: number; status: string; date: string; reference: string }
export interface Stats   { totalSales: number; totalRevenue: number; pendingOrders: number; activeListings: number }
