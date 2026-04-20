import { describe, it, expect, beforeAll, afterAll } from 'vitest'
import { buildApp } from '../../src/app.js'

describe('Catalog routes', () => {
  let app

  beforeAll(async () => { app = await buildApp({ logger: false }) })
  afterAll(async () => { await app.close() })

  it('GET /catalog/products returns 501 (stub)', async () => {
    const res = await app.inject({ method: 'GET', url: '/catalog/products' })
    expect(res.statusCode).toBe(501)
  })

  it('GET /catalog/search returns 400 without q param', async () => {
    const res = await app.inject({ method: 'GET', url: '/catalog/search' })
    expect(res.statusCode).toBe(400)
  })

  it('GET /catalog/search returns 501 with q param (stub)', async () => {
    const res = await app.inject({ method: 'GET', url: '/catalog/search?q=laptop' })
    expect(res.statusCode).toBe(501)
  })

  it('GET /catalog/feed returns 501 (stub)', async () => {
    const res = await app.inject({ method: 'GET', url: '/catalog/feed' })
    expect(res.statusCode).toBe(501)
  })

  it('GET /catalog/categories returns 501 (stub)', async () => {
    const res = await app.inject({ method: 'GET', url: '/catalog/categories' })
    expect(res.statusCode).toBe(501)
  })
})
