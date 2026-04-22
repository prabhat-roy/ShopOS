import { Component } from '@angular/core'
import { RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router'
import { AuthService } from '../services/auth.service'
import { CommonModule } from '@angular/common'

@Component({
  selector: 'app-layout',
  standalone: true,
  imports: [RouterOutlet, RouterLink, RouterLinkActive, CommonModule],
  template: `
    <div class="layout">
      <aside class="sidebar">
        <div class="logo">ShopOS Partner</div>
        <nav>
          <a *ngFor="let l of links" [routerLink]="l.to" routerLinkActive="active" [routerLinkActiveOptions]="{exact: l.to === '/'}" class="nav-link">{{ l.label }}</a>
        </nav>
        <button class="logout" (click)="auth.logout()">Logout</button>
      </aside>
      <div class="content-wrap">
        <header class="topbar"><span class="email">{{ auth.email() }}</span></header>
        <main class="main"><router-outlet /></main>
      </div>
    </div>`,
  styles: [`
    .layout { display: flex; height: 100vh; font-family: system-ui, sans-serif; }
    .sidebar { width: 220px; background: #111; color: #fff; display: flex; flex-direction: column; }
    .logo { padding: 1.5rem; font-weight: 700; font-size: 1.1rem; border-bottom: 1px solid #374151; }
    .nav-link { display: block; padding: 0.625rem 1.5rem; color: #9ca3af; text-decoration: none; font-size: 0.875rem; }
    .nav-link.active { color: #fff; background: #374151; }
    .logout { margin-top: auto; padding: 1rem 1.5rem; background: none; border: none; color: #9ca3af; cursor: pointer; text-align: left; font-size: 0.875rem; }
    .content-wrap { flex: 1; display: flex; flex-direction: column; overflow: hidden; }
    .topbar { height: 60px; background: #fff; border-bottom: 1px solid #e5e7eb; display: flex; align-items: center; padding: 0 1.5rem; }
    .email { margin-left: auto; color: #6b7280; font-size: 0.875rem; }
    .main { flex: 1; overflow: auto; padding: 1.5rem; background: #f9fafb; }
  `]
})
export class LayoutComponent {
  constructor(public auth: AuthService) {}
  links = [
    { to: '/',           label: 'Dashboard'  },
    { to: '/contracts',  label: 'Contracts'  },
    { to: '/orders',     label: 'Orders'     },
    { to: '/invoices',   label: 'Invoices'   },
    { to: '/quotes',     label: 'Quotes'     },
    { to: '/profile',    label: 'Profile'    },
  ]
}
