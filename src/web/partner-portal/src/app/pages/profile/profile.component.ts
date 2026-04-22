import { Component } from '@angular/core'
import { FormBuilder, ReactiveFormsModule } from '@angular/forms'
import { CommonModule } from '@angular/common'

@Component({ selector: 'app-profile', standalone: true, imports: [ReactiveFormsModule, CommonModule], template: `
    <div style="max-width:600px">
      <h1 class="title">Company Profile</h1>
      <div class="card">
        <form [formGroup]="form" (ngSubmit)="save()">
          <div class="field" *ngFor="let f of fields">
            <label>{{ f.label }}</label><input [formControlName]="f.key" />
          </div>
          <button type="submit">{{ saved ? 'Saved!' : 'Save Changes' }}</button>
        </form>
      </div>
    </div>`,
  styles: [`.title{font-size:1.5rem;font-weight:700;margin-bottom:1.5rem}.card{background:#fff;border:1px solid #e5e7eb;border-radius:.5rem;padding:1.5rem}.field{margin-bottom:1rem}.field label{display:block;font-size:.875rem;font-weight:500;margin-bottom:.25rem}.field input{width:100%;border:1px solid #d1d5db;border-radius:.375rem;padding:.5rem .75rem;box-sizing:border-box}button{background:#111;color:#fff;border:none;border-radius:.375rem;padding:.5rem 1.5rem;cursor:pointer}`]
})
export class ProfileComponent {
  form = this.fb.group({ companyName: [''], vatNumber: [''], email: [''], phone: [''], country: [''] })
  saved = false
  fields = [{ label: 'Company Name', key: 'companyName' }, { label: 'VAT Number', key: 'vatNumber' }, { label: 'Email', key: 'email' }, { label: 'Phone', key: 'phone' }, { label: 'Country', key: 'country' }]
  constructor(private fb: FormBuilder) {}
  save() { this.saved = true; setTimeout(() => this.saved = false, 2000) }
}
