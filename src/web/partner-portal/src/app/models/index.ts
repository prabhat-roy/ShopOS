export interface Contract { id: string; title: string; status: string; startDate: string; endDate: string; value: number }
export interface Order    { id: string; poNumber: string; total: number; status: string; createdAt: string; items: number }
export interface Invoice  { id: string; number: string; amount: number; status: string; dueDate: string }
export interface Quote    { id: string; reference: string; total: number; status: string; createdAt: string }
export interface DashStats { activeContracts: number; pendingOrders: number; outstandingInvoices: number; totalSpend: number }
