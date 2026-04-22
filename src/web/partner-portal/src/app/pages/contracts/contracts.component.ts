import { Component, OnInit } from '@angular/core'
import { CommonModule } from '@angular/common'
import { ApiService } from '../../services/api.service'
import type { Contract } from '../../models'

@Component({
  selector: 'app-contracts',
  standalone: true,
  imports: [CommonModule],
  template: `
    <h1 class="title">Contracts</h1>
    <div class="table-wrap">
      <p *ngIf="loading" class="empty">Loading...</p>
      <table *ngIf="!loading && contracts.length">
        <thead><tr><th>Title</th><th>Status</th><th>Value</th><th>Start</th><th>End</th></tr></thead>
        <tbody>
          <tr *ngFor="let c of contracts">
            <td>{{ c.title }}</td>
            <td><span [class]="'badge badge-' + c.status">{{ c.status }}</span></td>
            <td>${{ c.value.toLocaleString() }}</td>
            <td>{{ c.startDate | date }}</td>
            <td>{{ c.endDate | date }}</td>
          </tr>
        </tbody>
      </table>
      <p *ngIf="!loading && !contracts.length" class="empty">No contracts found.</p>
    </div>`,
  styles: [`
    .title { font-size: 1.5rem; font-weight: 700; margin-bottom: 1.5rem; }
    .table-wrap { background: #fff; border: 1px solid #e5e7eb; border-radius: 0.5rem; overflow: hidden; }
    table { width: 100%; border-collapse: collapse; font-size: 0.875rem; }
    th { text-align: left; padding: 0.75rem 1rem; font-weight: 600; border-bottom: 2px solid #e5e7eb; }
    td { padding: 0.75rem 1rem; border-bottom: 1px solid #f3f4f6; }
    .badge { padding: 0.2rem 0.5rem; border-radius: 9999px; font-size: 0.75rem; font-weight: 600; }
    .badge-active { background: #d1fae5; color: #065f46; } .badge-expired { background: #fee2e2; color: #dc2626; } .badge-pending { background: #fef3c7; color: #92400e; }
    .empty { text-align: center; padding: 2rem; color: #9ca3af; }
  `]
})
export class ContractsComponent implements OnInit {
  contracts: Contract[] = []; loading = true
  constructor(private api: ApiService) {}
  ngOnInit() { this.api.get<Contract[]>('/partner/contracts').subscribe({ next: d => { this.contracts = d; this.loading = false }, error: () => this.loading = false }) }
}
