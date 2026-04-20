import { describe, it, expect, beforeAll, afterAll } from 'vitest'
import { buildApp } from '../../src/app.js'

describe('POST /auth/login', () => {
  let app

  beforeAll(async () => { app = await buildApp({ logger: false }) })
  afterAll(async () => { await app.close() })

  it('returns 400 when email is missing', async () => {
    const res = await app.inject({
      method: 'POST',
      url: '/auth/login',
      payload: { password: 'secret' },
    })
    expect(res.statusCode).toBe(400)
  })

  it('returns 400 when password is missing', async () => {
    const res = await app.inject({
      method: 'POST',
      url: '/auth/login',
      payload: { email: 'a@b.com' },
    })
    expect(res.statusCode).toBe(400)
  })

  it('returns 501 when service not yet implemented', async () => {
    const res = await app.inject({
      method: 'POST',
      url: '/auth/login',
      payload: { email: 'a@b.com', password: 'secret' },
    })
    expect(res.statusCode).toBe(501)
  })
})

describe('POST /auth/register', () => {
  let app

  beforeAll(async () => { app = await buildApp({ logger: false }) })
  afterAll(async () => { await app.close() })

  it('returns 400 when required fields missing', async () => {
    const res = await app.inject({
      method: 'POST',
      url: '/auth/register',
      payload: { email: 'a@b.com' },
    })
    expect(res.statusCode).toBe(400)
  })

  it('returns 400 when password too short', async () => {
    const res = await app.inject({
      method: 'POST',
      url: '/auth/register',
      payload: { email: 'a@b.com', password: '123', first_name: 'John' },
    })
    expect(res.statusCode).toBe(400)
  })
})
