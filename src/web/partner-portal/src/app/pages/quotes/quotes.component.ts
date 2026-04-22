import { Component, OnInit } from '@angular/core'
import { CommonModule } from '@angular/common'
import { ApiService } from '../../services/api.service'
import type { Quote } from '../../models'

@Component({ selector: 'app-quotes', standalone: true, imports: [CommonModule], template: `
    <h1 class="title">Quotes / RFQ</h1>
    <div class="table-wrap">
      <p *ngIf="loading" class="empty">Loading...</p>
      <table *ngIf="!loading && quotes.length">
        <thead><tr><th>Reference</th><th>Total</th><th>Status</th><th>Date</th></tr></thead>
        <tbody>
          <tr *ngFor="let q of quotes">
            <td>{{ q.reference }}</td><td>${{ q.total.toFixed(2) }}</td>
            <td>{{ q.status }}</td><td>{{ q.createdAt | date }}</td>
          </tr>
        </tbody>
      </table>
      <p *ngIf="!loading && !quotes.length" class="empty">No quotes found.</p>
    </div>`,
  styles: [`.title{font-size:1.5rem;font-weight:700;margin-bottom:1.5rem}.table-wrap{background:#fff;border:1px solid #e5e7eb;border-radius:.5rem;overflow:hidden}table{width:100%;border-collapse:collapse;font-size:.875rem}th{text-align:left;padding:.75rem 1rem;font-weight:600;border-bottom:2px solid #e5e7eb}td{padding:.75rem 1rem;border-bottom:1px solid #f3f4f6}.empty{text-align:center;padding:2rem;color:#9ca3af}`]
})
export class QuotesComponent implements OnInit {
  quotes: Quote[] = []; loading = true
  constructor(private api: ApiService) {}
  ngOnInit() { this.api.get<Quote[]>('/partner/quotes').subscribe({ next: d => { this.quotes = d; this.loading = false }, error: () => this.loading = false }) }
}
