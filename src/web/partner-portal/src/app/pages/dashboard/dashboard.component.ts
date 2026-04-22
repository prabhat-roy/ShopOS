import { Component, OnInit } from '@angular/core'
import { CommonModule } from '@angular/common'
import { ApiService } from '../../services/api.service'
import type { DashStats } from '../../models'

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [CommonModule],
  template: `
    <h1 class="title">Dashboard</h1>
    <div class="grid">
      <div class="stat" *ngFor="let s of statCards">
        <p class="label">{{ s.label }}</p>
        <p class="value">{{ s.prefix }}{{ s.value }}</p>
      </div>
    </div>`,
  styles: [`
    .title { font-size: 1.5rem; font-weight: 700; margin-bottom: 1.5rem; }
    .grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 1rem; }
    .stat { background: #fff; border: 1px solid #e5e7eb; border-radius: 0.5rem; padding: 1.5rem; }
    .label { font-size: 0.75rem; color: #6b7280; text-transform: uppercase; margin-bottom: 0.5rem; }
    .value { font-size: 1.875rem; font-weight: 700; }
  `]
})
export class DashboardComponent implements OnInit {
  stats: DashStats | null = null
  get statCards() {
    return [
      { label: 'Active Contracts',      prefix: '',  value: this.stats?.activeContracts ?? 0 },
      { label: 'Pending Orders',        prefix: '',  value: this.stats?.pendingOrders ?? 0 },
      { label: 'Outstanding Invoices',  prefix: '',  value: this.stats?.outstandingInvoices ?? 0 },
      { label: 'Total Spend',           prefix: '$', value: (this.stats?.totalSpend ?? 0).toLocaleString() },
    ]
  }
  constructor(private api: ApiService) {}
  ngOnInit() { this.api.get<DashStats>('/partner/stats').subscribe({ next: d => this.stats = d, error: () => {} }) }
}
