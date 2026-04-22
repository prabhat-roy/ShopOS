import { Component, OnInit } from '@angular/core'
import { CommonModule } from '@angular/common'
import { ApiService } from '../../services/api.service'
import type { Order } from '../../models'

@Component({ selector: 'app-orders', standalone: true, imports: [CommonModule], template: `
    <h1 class="title">Purchase Orders</h1>
    <div class="table-wrap">
      <p *ngIf="loading" class="empty">Loading...</p>
      <table *ngIf="!loading && orders.length">
        <thead><tr><th>PO Number</th><th>Items</th><th>Total</th><th>Status</th><th>Date</th></tr></thead>
        <tbody>
          <tr *ngFor="let o of orders">
            <td>{{ o.poNumber }}</td><td>{{ o.items }}</td>
            <td>${{ o.total.toFixed(2) }}</td><td>{{ o.status }}</td>
            <td>{{ o.createdAt | date }}</td>
          </tr>
        </tbody>
      </table>
      <p *ngIf="!loading && !orders.length" class="empty">No orders found.</p>
    </div>`,
  styles: [`.title{font-size:1.5rem;font-weight:700;margin-bottom:1.5rem}.table-wrap{background:#fff;border:1px solid #e5e7eb;border-radius:.5rem;overflow:hidden}table{width:100%;border-collapse:collapse;font-size:.875rem}th{text-align:left;padding:.75rem 1rem;font-weight:600;border-bottom:2px solid #e5e7eb}td{padding:.75rem 1rem;border-bottom:1px solid #f3f4f6}.empty{text-align:center;padding:2rem;color:#9ca3af}`]
})
export class OrdersComponent implements OnInit {
  orders: Order[] = []; loading = true
  constructor(private api: ApiService) {}
  ngOnInit() { this.api.get<Order[]>('/partner/orders').subscribe({ next: d => { this.orders = d; this.loading = false }, error: () => this.loading = false }) }
}
