import { Routes } from '@angular/router'
import { authGuard } from './guards/auth.guard'
import { LoginComponent } from './pages/login/login.component'
import { LayoutComponent } from './layout/layout.component'
import { DashboardComponent } from './pages/dashboard/dashboard.component'
import { ContractsComponent } from './pages/contracts/contracts.component'
import { OrdersComponent } from './pages/orders/orders.component'
import { InvoicesComponent } from './pages/invoices/invoices.component'
import { QuotesComponent } from './pages/quotes/quotes.component'
import { ProfileComponent } from './pages/profile/profile.component'

export const routes: Routes = [
  { path: 'login', component: LoginComponent },
  {
    path: '',
    component: LayoutComponent,
    canActivate: [authGuard],
    children: [
      { path: '',          component: DashboardComponent },
      { path: 'contracts', component: ContractsComponent },
      { path: 'orders',    component: OrdersComponent },
      { path: 'invoices',  component: InvoicesComponent },
      { path: 'quotes',    component: QuotesComponent },
      { path: 'profile',   component: ProfileComponent },
    ],
  },
  { path: '**', redirectTo: '' },
]
