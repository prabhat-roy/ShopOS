'use strict';

const request = require('supertest');
const { createApp } = require('../src/app');

// Mock the Kafka consumer so tests don't try to connect
jest.mock('../src/kafka/consumer', () => ({
  start: jest.fn().mockResolvedValue(undefined),
  stop: jest.fn().mockResolvedValue(undefined),
  isRunning: jest.fn().mockReturnValue(false),
}));

describe('whatsapp-service', () => {
  let app;

  beforeAll(() => {
    app = createApp();
  });

  describe('GET /healthz', () => {
    it('returns 200 with status ok', async () => {
      const res = await request(app).get('/healthz');
      expect(res.status).toBe(200);
      expect(res.body.status).toBe('ok');
      expect(res.body.service).toBe('whatsapp-service');
    });
  });

  describe('GET /whatsapp/stats', () => {
    it('returns stats object', async () => {
      const res = await request(app).get('/whatsapp/stats');
      expect(res.status).toBe(200);
      expect(res.body).toHaveProperty('stats');
      expect(res.body.stats).toHaveProperty('sent');
      expect(res.body.stats).toHaveProperty('failed');
      expect(res.body.stats).toHaveProperty('total');
    });
  });

  describe('GET /unknown', () => {
    it('returns 404', async () => {
      const res = await request(app).get('/unknown');
      expect(res.status).toBe(404);
    });
  });
});
