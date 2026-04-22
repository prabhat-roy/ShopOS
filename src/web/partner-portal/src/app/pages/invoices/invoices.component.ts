import { Component, OnInit } from '@angular/core'
import { CommonModule } from '@angular/common'
import { ApiService } from '../../services/api.service'
import type { Invoice } from '../../models'

@Component({ selector: 'app-invoices', standalone: true, imports: [CommonModule], template: `
    <h1 class="title">Invoices</h1>
    <div class="table-wrap">
      <p *ngIf="loading" class="empty">Loading...</p>
      <table *ngIf="!loading && invoices.length">
        <thead><tr><th>Invoice #</th><th>Amount</th><th>Status</th><th>Due Date</th></tr></thead>
        <tbody>
          <tr *ngFor="let inv of invoices">
            <td>{{ inv.number }}</td><td>${{ inv.amount.toFixed(2) }}</td>
            <td><span [class]="'badge badge-' + inv.status">{{ inv.status }}</span></td>
            <td>{{ inv.dueDate | date }}</td>
          </tr>
        </tbody>
      </table>
      <p *ngIf="!loading && !invoices.length" class="empty">No invoices found.</p>
    </div>`,
  styles: [`.title{font-size:1.5rem;font-weight:700;margin-bottom:1.5rem}.table-wrap{background:#fff;border:1px solid #e5e7eb;border-radius:.5rem;overflow:hidden}table{width:100%;border-collapse:collapse;font-size:.875rem}th{text-align:left;padding:.75rem 1rem;font-weight:600;border-bottom:2px solid #e5e7eb}td{padding:.75rem 1rem;border-bottom:1px solid #f3f4f6}.badge{padding:.2rem .5rem;border-radius:9999px;font-size:.75rem;font-weight:600}.badge-paid{background:#d1fae5;color:#065f46}.badge-pending{background:#fef3c7;color:#92400e}.badge-overdue{background:#fee2e2;color:#dc2626}.empty{text-align:center;padding:2rem;color:#9ca3af}`]
})
export class InvoicesComponent implements OnInit {
  invoices: Invoice[] = []; loading = true
  constructor(private api: ApiService) {}
  ngOnInit() { this.api.get<Invoice[]>('/partner/invoices').subscribe({ next: d => { this.invoices = d; this.loading = false }, error: () => this.loading = false }) }
}
