import { Injectable } from '@angular/core'
import { HttpClient } from '@angular/common/http'
import { Observable } from 'rxjs'

@Injectable({ providedIn: 'root' })
export class ApiService {
  private base = '/api'
  constructor(private http: HttpClient) {}
  get<T>(path: string): Observable<T> { return this.http.get<T>(`${this.base}${path}`) }
  post<T>(path: string, body: unknown): Observable<T> { return this.http.post<T>(`${this.base}${path}`, body) }
}
