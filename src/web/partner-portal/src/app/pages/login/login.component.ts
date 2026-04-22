import { Component } from '@angular/core'
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms'
import { Router } from '@angular/router'
import { CommonModule } from '@angular/common'
import { AuthService } from '../../services/auth.service'

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [ReactiveFormsModule, CommonModule],
  template: `
    <div class="wrap">
      <div class="box">
        <h1>Partner Login</h1>
        <p *ngIf="error" class="error">{{ error }}</p>
        <form [formGroup]="form" (ngSubmit)="submit()">
          <div class="field"><label>Email</label><input type="email" formControlName="email" /></div>
          <div class="field"><label>Password</label><input type="password" formControlName="password" /></div>
          <button type="submit" [disabled]="loading">{{ loading ? 'Logging in...' : 'Login' }}</button>
        </form>
      </div>
    </div>`,
  styles: [`
    .wrap { min-height: 100vh; display: flex; align-items: center; justify-content: center; background: #f9fafb; font-family: system-ui, sans-serif; }
    .box { background: #fff; border: 1px solid #e5e7eb; border-radius: 0.5rem; padding: 2rem; width: 360px; }
    h1 { font-size: 1.5rem; font-weight: 700; margin-bottom: 1.5rem; }
    .field { margin-bottom: 1rem; } .field label { display: block; font-size: 0.875rem; font-weight: 500; margin-bottom: 0.25rem; }
    .field input { width: 100%; border: 1px solid #d1d5db; border-radius: 0.375rem; padding: 0.5rem 0.75rem; box-sizing: border-box; }
    button { width: 100%; background: #111; color: #fff; border: none; border-radius: 0.375rem; padding: 0.75rem; font-weight: 500; cursor: pointer; }
    .error { color: #dc2626; font-size: 0.875rem; margin-bottom: 1rem; }
  `]
})
export class LoginComponent {
  form = this.fb.group({ email: ['', [Validators.required, Validators.email]], password: ['', Validators.required] })
  loading = false; error = ''

  constructor(private fb: FormBuilder, private auth: AuthService, private router: Router) {}

  submit() {
    if (this.form.invalid) return
    this.loading = true; this.error = ''
    const { email, password } = this.form.value
    this.auth.login(email!, password!).subscribe({
      next: ({ token }) => { this.auth.setAuth(token, email!); this.router.navigate(['/']) },
      error: () => { this.error = 'Invalid credentials'; this.loading = false },
    })
  }
}
