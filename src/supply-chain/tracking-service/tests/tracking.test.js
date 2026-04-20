'use strict';

const request = require('supertest');
const createApp = require('../src/app');
const TrackingController = require('../src/controllers/TrackingController');
const TrackingService = require('../src/services/TrackingService');
const ShipmentRepository = require('../src/repositories/ShipmentRepository');

// Mock the repository so no real MongoDB connection is needed
jest.mock('../src/repositories/ShipmentRepository');

const SAMPLE_SHIPMENT = {
  _id: '665f1a2b3c4d5e6f7a8b9c0d',
  trackingNumber: 'TRK-ABC123456789',
  carrier: 'FedEx',
  status: 'created',
  originAddress: '123 Sender St, New York, NY 10001',
  destinationAddress: '456 Receiver Ave, Los Angeles, CA 90001',
  estimatedDelivery: new Date('2025-01-15').toISOString(),
  actualDelivery: null,
  events: [
    {
      timestamp: new Date().toISOString(),
      location: '123 Sender St, New York, NY 10001',
      description: 'Shipment created',
      status: 'created',
    },
  ],
  metadata: {},
  createdAt: new Date().toISOString(),
  updatedAt: new Date().toISOString(),
};

function buildApp(repoOverrides = {}) {
  const mockRepo = new ShipmentRepository();
  Object.assign(mockRepo, repoOverrides);
  const service = new TrackingService(mockRepo);
  const controller = new TrackingController(service);
  return createApp({ controller });
}

describe('TrackingService — POST /shipments', () => {
  test('creates a shipment and returns 201', async () => {
    const app = buildApp({
      create: jest.fn().mockResolvedValue(SAMPLE_SHIPMENT),
    });

    const res = await request(app)
      .post('/shipments')
      .send({ carrier: 'FedEx', originAddress: '123 Sender St', destinationAddress: '456 Receiver Ave' });

    expect(res.status).toBe(201);
    expect(res.body.carrier).toBe('FedEx');
    expect(res.body.trackingNumber).toBe('TRK-ABC123456789');
  });

  test('returns 400 when carrier is missing', async () => {
    const app = buildApp();

    const res = await request(app)
      .post('/shipments')
      .send({ originAddress: '123 Sender St' });

    expect(res.status).toBe(400);
    expect(res.body.error).toMatch(/carrier is required/i);
  });
});

describe('TrackingService — GET /shipments/:trackingNumber', () => {
  test('returns 200 with shipment when found', async () => {
    const app = buildApp({
      findByTrackingNumber: jest.fn().mockResolvedValue(SAMPLE_SHIPMENT),
    });

    const res = await request(app).get('/shipments/TRK-ABC123456789');

    expect(res.status).toBe(200);
    expect(res.body.trackingNumber).toBe('TRK-ABC123456789');
    expect(res.body.carrier).toBe('FedEx');
  });

  test('returns 404 when shipment is not found', async () => {
    const app = buildApp({
      findByTrackingNumber: jest.fn().mockResolvedValue(null),
    });

    const res = await request(app).get('/shipments/NOTEXIST-999');

    expect(res.status).toBe(404);
    expect(res.body.error).toMatch(/not found/i);
  });
});

describe('TrackingService — GET /shipments (list)', () => {
  test('returns 200 with shipments array and total', async () => {
    const app = buildApp({
      list: jest.fn().mockResolvedValue({ shipments: [SAMPLE_SHIPMENT], total: 1 }),
    });

    const res = await request(app).get('/shipments?carrier=FedEx&status=created');

    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.shipments)).toBe(true);
    expect(res.body.total).toBe(1);
    expect(res.body.shipments[0].carrier).toBe('FedEx');
  });

  test('returns empty list when no shipments match', async () => {
    const app = buildApp({
      list: jest.fn().mockResolvedValue({ shipments: [], total: 0 }),
    });

    const res = await request(app).get('/shipments?carrier=DHL');

    expect(res.status).toBe(200);
    expect(res.body.shipments).toHaveLength(0);
    expect(res.body.total).toBe(0);
  });
});

describe('TrackingService — POST /shipments/:trackingNumber/events', () => {
  test('adds event and returns 201 with updated shipment', async () => {
    const updatedShipment = {
      ...SAMPLE_SHIPMENT,
      events: [
        ...SAMPLE_SHIPMENT.events,
        { timestamp: new Date().toISOString(), location: 'Chicago Hub', description: 'In transit', status: 'in_transit' },
      ],
    };

    const app = buildApp({
      addEvent: jest.fn().mockResolvedValue(updatedShipment),
    });

    const res = await request(app)
      .post('/shipments/TRK-ABC123456789/events')
      .send({ description: 'In transit', location: 'Chicago Hub', status: 'in_transit' });

    expect(res.status).toBe(201);
    expect(res.body.events).toHaveLength(2);
  });

  test('returns 404 when adding event to non-existent shipment', async () => {
    const app = buildApp({
      addEvent: jest.fn().mockResolvedValue(null),
    });

    const res = await request(app)
      .post('/shipments/NOTEXIST-999/events')
      .send({ description: 'Package scanned' });

    expect(res.status).toBe(404);
    expect(res.body.error).toMatch(/not found/i);
  });

  test('returns 400 when description is missing', async () => {
    const app = buildApp();

    const res = await request(app)
      .post('/shipments/TRK-ABC123456789/events')
      .send({ location: 'Chicago Hub' });

    expect(res.status).toBe(400);
    expect(res.body.error).toMatch(/description is required/i);
  });
});

describe('TrackingService — PATCH /shipments/:trackingNumber/status', () => {
  test('updates status and returns 200 with updated shipment', async () => {
    const updated = { ...SAMPLE_SHIPMENT, status: 'in_transit' };

    const app = buildApp({
      updateStatus: jest.fn().mockResolvedValue(updated),
    });

    const res = await request(app)
      .patch('/shipments/TRK-ABC123456789/status')
      .send({ status: 'in_transit', location: 'Chicago Hub' });

    expect(res.status).toBe(200);
    expect(res.body.status).toBe('in_transit');
  });

  test('returns 400 for invalid status value', async () => {
    const app = buildApp({
      updateStatus: jest.fn().mockResolvedValue({ ...SAMPLE_SHIPMENT, status: 'in_transit' }),
    });

    const res = await request(app)
      .patch('/shipments/TRK-ABC123456789/status')
      .send({ status: 'exploded' });

    expect(res.status).toBe(400);
    expect(res.body.error).toMatch(/invalid status/i);
  });

  test('returns 400 when status is missing', async () => {
    const app = buildApp();

    const res = await request(app)
      .patch('/shipments/TRK-ABC123456789/status')
      .send({});

    expect(res.status).toBe(400);
    expect(res.body.error).toMatch(/status is required/i);
  });
});

describe('TrackingService — POST /shipments/:trackingNumber/deliver', () => {
  test('marks shipment as delivered and returns 204', async () => {
    const delivered = { ...SAMPLE_SHIPMENT, status: 'delivered', actualDelivery: new Date().toISOString() };

    const app = buildApp({
      updateDelivered: jest.fn().mockResolvedValue(delivered),
    });

    const res = await request(app).post('/shipments/TRK-ABC123456789/deliver');

    expect(res.status).toBe(204);
    expect(res.body).toEqual({});
  });

  test('returns 404 when shipment to deliver is not found', async () => {
    const app = buildApp({
      updateDelivered: jest.fn().mockResolvedValue(null),
    });

    const res = await request(app).post('/shipments/NOTEXIST-999/deliver');

    expect(res.status).toBe(404);
    expect(res.body.error).toMatch(/not found/i);
  });
});

describe('Health check', () => {
  test('GET /healthz returns 200 with status ok', async () => {
    const app = buildApp();
    const res = await request(app).get('/healthz');
    expect(res.status).toBe(200);
    expect(res.body.status).toBe('ok');
  });
});
