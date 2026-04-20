'use strict';

const request = require('supertest');
const { createApp } = require('../src/index');
const { RateStore } = require('../src/rates');
const { convert, convertMany } = require('../src/converter');

const app = createApp();

// ── Unit tests: converter logic ───────────────────────────────────────────────

describe('convert() unit tests', () => {
  let store;

  beforeEach(() => {
    store = new RateStore();
  });

  test('USD → EUR uses correct math', () => {
    // With seed: EUR = 0.92, so 100 USD → 92 EUR
    const result = convert(100, 'USD', 'EUR', store);
    expect(result.from).toBe('USD');
    expect(result.to).toBe('EUR');
    expect(result.rate).toBeCloseTo(0.92, 5);
    expect(result.converted).toBeCloseTo(92, 2);
    expect(result.amount).toBe(100);
    expect(result.timestamp).toBeDefined();
  });

  test('EUR → USD is the inverse of USD → EUR', () => {
    const usdToEur = convert(100, 'USD', 'EUR', store);
    const eurToUsd = convert(usdToEur.converted, 'EUR', 'USD', store);
    expect(eurToUsd.converted).toBeCloseTo(100, 2);
  });

  test('same currency returns 1:1 (amount unchanged)', () => {
    const result = convert(250, 'USD', 'USD', store);
    expect(result.rate).toBe(1);
    expect(result.converted).toBe(250);
  });

  test('same currency — non-USD — returns 1:1', () => {
    const result = convert(500, 'EUR', 'EUR', store);
    expect(result.rate).toBe(1);
    expect(result.converted).toBe(500);
  });

  test('cross rate EUR → JPY', () => {
    // EUR/USD = 1/0.92, USD/JPY = 149.5
    // EUR → JPY rate = (1/0.92) * 149.5
    const expected = (1 / 0.92) * 149.5;
    const result = convert(1, 'EUR', 'JPY', store);
    expect(result.rate).toBeCloseTo(expected, 4);
  });

  test('unsupported source currency throws and returns 400 via API', async () => {
    const res = await request(app)
      .post('/currencies/convert')
      .send({ amount: 100, from: 'XYZ', to: 'USD' });
    expect(res.status).toBe(400);
    expect(res.body.error).toMatch(/unsupported currency/i);
  });

  test('unsupported target currency throws', () => {
    expect(() => convert(100, 'USD', 'FAKE', store)).toThrow(/unsupported currency/i);
  });

  test('convertMany returns correct number of results', () => {
    const results = convertMany(100, 'USD', ['EUR', 'GBP', 'JPY'], store);
    expect(results).toHaveLength(3);
    expect(results[0].to).toBe('EUR');
    expect(results[1].to).toBe('GBP');
    expect(results[2].to).toBe('JPY');
  });
});

// ── Integration tests: HTTP routes ───────────────────────────────────────────

describe('GET /healthz', () => {
  test('returns 200 with status ok', async () => {
    const res = await request(app).get('/healthz');
    expect(res.status).toBe(200);
    expect(res.body).toEqual({ status: 'ok' });
  });
});

describe('GET /currencies', () => {
  test('returns list of supported currencies', async () => {
    const res = await request(app).get('/currencies');
    expect(res.status).toBe(200);
    expect(res.body.base).toBe('USD');
    expect(Array.isArray(res.body.currencies)).toBe(true);
    expect(res.body.currencies.length).toBeGreaterThan(0);

    const usd = res.body.currencies.find((c) => c.code === 'USD');
    expect(usd).toBeDefined();
    expect(usd.rate_vs_usd).toBe(1);
  });
});

describe('GET /currencies/rates', () => {
  test('returns all rates with base USD', async () => {
    const res = await request(app).get('/currencies/rates');
    expect(res.status).toBe(200);
    expect(res.body.base).toBe('USD');
    expect(typeof res.body.rates).toBe('object');
    expect(res.body.rates.USD).toBe(1);
    expect(res.body.rates.EUR).toBeCloseTo(0.92, 2);
  });
});

describe('GET /currencies/rate', () => {
  test('returns rate for USD → EUR', async () => {
    const res = await request(app).get('/currencies/rate?from=USD&to=EUR');
    expect(res.status).toBe(200);
    expect(res.body.from).toBe('USD');
    expect(res.body.to).toBe('EUR');
    expect(res.body.rate).toBeCloseTo(0.92, 2);
  });

  test('returns 400 when from param is missing', async () => {
    const res = await request(app).get('/currencies/rate?to=EUR');
    expect(res.status).toBe(400);
    expect(res.body.error).toMatch(/"from" and "to"/);
  });

  test('returns 400 for unsupported currency', async () => {
    const res = await request(app).get('/currencies/rate?from=USD&to=NOPE');
    expect(res.status).toBe(400);
    expect(res.body.error).toMatch(/unsupported currency/i);
  });
});

describe('POST /currencies/convert', () => {
  test('converts USD to EUR correctly', async () => {
    const res = await request(app)
      .post('/currencies/convert')
      .send({ amount: 100, from: 'USD', to: 'EUR' });
    expect(res.status).toBe(200);
    expect(res.body.converted).toBeCloseTo(92, 1);
    expect(res.body.from).toBe('USD');
    expect(res.body.to).toBe('EUR');
  });

  test('same currency returns 1:1', async () => {
    const res = await request(app)
      .post('/currencies/convert')
      .send({ amount: 500, from: 'GBP', to: 'GBP' });
    expect(res.status).toBe(200);
    expect(res.body.converted).toBe(500);
    expect(res.body.rate).toBe(1);
  });

  test('returns 400 for unsupported currency', async () => {
    const res = await request(app)
      .post('/currencies/convert')
      .send({ amount: 100, from: 'USD', to: 'INVALID' });
    expect(res.status).toBe(400);
    expect(res.body.error).toMatch(/unsupported currency/i);
  });

  test('returns 400 when amount is missing', async () => {
    const res = await request(app)
      .post('/currencies/convert')
      .send({ from: 'USD', to: 'EUR' });
    expect(res.status).toBe(400);
  });

  test('returns 400 when amount is not a number', async () => {
    const res = await request(app)
      .post('/currencies/convert')
      .send({ amount: 'abc', from: 'USD', to: 'EUR' });
    expect(res.status).toBe(400);
  });
});

describe('POST /currencies/convert/many', () => {
  test('converts USD to multiple currencies', async () => {
    const res = await request(app)
      .post('/currencies/convert/many')
      .send({ amount: 100, from: 'USD', targets: ['EUR', 'GBP', 'JPY'] });
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.results)).toBe(true);
    expect(res.body.results).toHaveLength(3);
  });

  test('returns 400 when targets is empty', async () => {
    const res = await request(app)
      .post('/currencies/convert/many')
      .send({ amount: 100, from: 'USD', targets: [] });
    expect(res.status).toBe(400);
  });

  test('returns 400 when targets contains unsupported currency', async () => {
    const res = await request(app)
      .post('/currencies/convert/many')
      .send({ amount: 100, from: 'USD', targets: ['EUR', 'FAKE'] });
    expect(res.status).toBe(400);
  });
});
