import { Injectable } from '@angular/core'
import { HttpClient, HttpHeaders } from '@angular/common/http'
import { Observable, throwError } from 'rxjs'
import { catchError } from 'rxjs/operators'

@Injectable({ providedIn: 'root' })
export class ApiService {
  private base = '/api'

  constructor(private http: HttpClient) {}

  private headers(): HttpHeaders {
    const token = localStorage.getItem('partner_token')
    return new HttpHeaders(token ? { Authorization: `Bearer ${token}` } : {})
  }

  get<T>(path: string): Observable<T> {
    return this.http.get<T>(`${this.base}${path}`, { headers: this.headers() }).pipe(
      catchError(e => { if (e.status === 401) { localStorage.removeItem('partner_token'); window.location.href = '/login' } return throwError(() => e) })
    )
  }

  post<T>(path: string, body: unknown): Observable<T> {
    return this.http.post<T>(`${this.base}${path}`, body, { headers: this.headers() })
  }

  put<T>(path: string, body: unknown): Observable<T> {
    return this.http.put<T>(`${this.base}${path}`, body, { headers: this.headers() })
  }
}
