import { describe, it, expect, beforeAll, afterAll } from 'vitest'
import { buildApp } from '../../src/app.js'

describe('POST /device/register', () => {
  let app

  beforeAll(async () => { app = await buildApp({ logger: false }) })
  afterAll(async () => { await app.close() })

  it('returns 400 when token is missing', async () => {
    const res = await app.inject({
      method: 'POST',
      url: '/device/register',
      payload: { platform: 'ios' },
    })
    expect(res.statusCode).toBe(400)
  })

  it('returns 400 when platform is invalid', async () => {
    const res = await app.inject({
      method: 'POST',
      url: '/device/register',
      payload: { token: 'abc123', platform: 'windows' },
    })
    expect(res.statusCode).toBe(400)
  })

  it('returns 501 with valid payload (stub)', async () => {
    const res = await app.inject({
      method: 'POST',
      url: '/device/register',
      payload: { token: 'fcm-token-xyz', platform: 'android', device_id: 'd1' },
      headers: { 'x-user-id': 'u1' },
    })
    expect(res.statusCode).toBe(501)
  })
})
