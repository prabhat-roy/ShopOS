import { Injectable, signal } from '@angular/core'
import { Router } from '@angular/router'
import { ApiService } from './api.service'

@Injectable({ providedIn: 'root' })
export class AuthService {
  token = signal<string | null>(localStorage.getItem('partner_token'))
  email = signal<string | null>(localStorage.getItem('partner_email'))

  constructor(private api: ApiService, private router: Router) {}

  login(email: string, password: string) {
    return this.api.post<{ token: string }>('/partner/auth/login', { email, password })
  }

  setAuth(token: string, email: string) {
    this.token.set(token); this.email.set(email)
    localStorage.setItem('partner_token', token)
    localStorage.setItem('partner_email', email)
  }

  logout() {
    this.token.set(null); this.email.set(null)
    localStorage.removeItem('partner_token'); localStorage.removeItem('partner_email')
    this.router.navigate(['/login'])
  }

  isLoggedIn(): boolean { return !!this.token() }
}
