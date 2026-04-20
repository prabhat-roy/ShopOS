'use strict';

/**
 * Tests for configurator-service.
 *
 * Mongoose is mocked so no real MongoDB connection is needed.
 */

// ─── Mock mongoose before requiring the app ───────────────────────────────────
jest.mock('mongoose', () => {
  const actualMongoose = jest.requireActual('mongoose');
  return {
    ...actualMongoose,
    connect: jest.fn().mockResolvedValue(undefined),
    disconnect: jest.fn().mockResolvedValue(undefined),
    model: actualMongoose.model.bind(actualMongoose),
    Schema: actualMongoose.Schema,
  };
});

// Mock the Configurator model methods used by the service
jest.mock('../src/models/configuratorModel', () => {
  const mockDoc = {
    productId: 'prod-123',
    options: [
      {
        name: 'ram',
        type: 'select',
        required: true,
        choices: [
          { value: '8gb', label: '8 GB', priceAdj: 0, available: true },
          { value: '16gb', label: '16 GB', priceAdj: 50, available: true },
          { value: '32gb', label: '32 GB', priceAdj: 150, available: false },
        ],
      },
      {
        name: 'color',
        type: 'radio',
        required: false,
        choices: [
          { value: 'silver', label: 'Silver', priceAdj: 0, available: true },
          { value: 'space-gray', label: 'Space Gray', priceAdj: 10, available: true },
        ],
      },
    ],
    rules: [
      {
        condition: { optionName: 'ram', value: '8gb' },
        effect: { optionName: 'color', allowedValues: ['silver'] },
      },
    ],
    createdAt: new Date(),
    updatedAt: new Date(),
  };

  return {
    findOne: jest.fn(),
    findOneAndUpdate: jest.fn(),
    findOneAndDelete: jest.fn(),
    __mockDoc: mockDoc,
  };
});

const request = require('supertest');
const app = require('../index');
const Configurator = require('../src/models/configuratorModel');

// Suppress console output during tests
beforeAll(() => {
  jest.spyOn(console, 'log').mockImplementation(() => {});
  jest.spyOn(console, 'error').mockImplementation(() => {});
});

afterAll(() => {
  jest.restoreAllMocks();
});

afterEach(() => {
  jest.clearAllMocks();
});

// ─── /healthz ─────────────────────────────────────────────────────────────────

describe('GET /healthz', () => {
  it('returns 200 with status ok', async () => {
    const res = await request(app).get('/healthz');
    expect(res.status).toBe(200);
    expect(res.body).toEqual({ status: 'ok' });
  });
});

// ─── POST /configurators/:productId ──────────────────────────────────────────

describe('POST /configurators/:productId', () => {
  it('creates a configurator and returns 201', async () => {
    const mockDoc = Configurator.__mockDoc;
    Configurator.findOneAndUpdate.mockReturnValue({ lean: jest.fn().mockResolvedValue(mockDoc) });

    const body = {
      options: [
        {
          name: 'ram',
          type: 'select',
          required: true,
          choices: [
            { value: '8gb', label: '8 GB', priceAdj: 0, available: true },
            { value: '16gb', label: '16 GB', priceAdj: 50, available: true },
          ],
        },
      ],
      rules: [],
    };

    const res = await request(app).post('/configurators/prod-123').send(body);

    expect(res.status).toBe(201);
    expect(res.body.productId).toBe('prod-123');
    expect(Configurator.findOneAndUpdate).toHaveBeenCalledTimes(1);
  });

  it('returns 400 when body fails validation', async () => {
    const res = await request(app)
      .post('/configurators/prod-bad')
      .send({ options: [{ name: 'ram' }] }); // missing 'type'

    expect(res.status).toBe(400);
    expect(res.body).toHaveProperty('error');
  });
});

// ─── GET /configurators/:productId ───────────────────────────────────────────

describe('GET /configurators/:productId', () => {
  it('returns 200 with the configurator when found', async () => {
    const mockDoc = Configurator.__mockDoc;
    Configurator.findOne.mockReturnValue({ lean: jest.fn().mockResolvedValue(mockDoc) });

    const res = await request(app).get('/configurators/prod-123');

    expect(res.status).toBe(200);
    expect(res.body.productId).toBe('prod-123');
    expect(res.body.options).toHaveLength(2);
  });

  it('returns 404 when the configurator does not exist', async () => {
    Configurator.findOne.mockReturnValue({ lean: jest.fn().mockResolvedValue(null) });

    const res = await request(app).get('/configurators/nonexistent');

    expect(res.status).toBe(404);
    expect(res.body).toHaveProperty('error');
  });
});

// ─── POST /configurators/:productId/validate ─────────────────────────────────

describe('POST /configurators/:productId/validate', () => {
  beforeEach(() => {
    const mockDoc = Configurator.__mockDoc;
    Configurator.findOne.mockReturnValue({ lean: jest.fn().mockResolvedValue(mockDoc) });
  });

  it('returns valid=true and correct totalPriceAdj for a good selection', async () => {
    const res = await request(app)
      .post('/configurators/prod-123/validate')
      .send({ selections: { ram: '16gb', color: 'silver' } });

    expect(res.status).toBe(200);
    expect(res.body.valid).toBe(true);
    expect(res.body.errors).toHaveLength(0);
    expect(res.body.totalPriceAdj).toBe(50); // 16gb adds $50, silver adds $0
  });

  it('returns valid=false when a required option is missing', async () => {
    const res = await request(app)
      .post('/configurators/prod-123/validate')
      .send({ selections: { color: 'silver' } }); // ram is required but missing

    expect(res.status).toBe(422);
    expect(res.body.valid).toBe(false);
    expect(res.body.errors.some((e) => e.includes('ram'))).toBe(true);
  });

  it('returns valid=false when a rule is violated', async () => {
    // Rule: when ram=8gb, color must be silver. We pick space-gray — violation.
    const res = await request(app)
      .post('/configurators/prod-123/validate')
      .send({ selections: { ram: '8gb', color: 'space-gray' } });

    expect(res.status).toBe(422);
    expect(res.body.valid).toBe(false);
    expect(res.body.errors.some((e) => e.includes('space-gray'))).toBe(true);
  });

  it('returns valid=false when an invalid choice value is provided', async () => {
    const res = await request(app)
      .post('/configurators/prod-123/validate')
      .send({ selections: { ram: '64gb' } }); // '64gb' is not a valid choice

    expect(res.status).toBe(422);
    expect(res.body.valid).toBe(false);
  });

  it('returns valid=false when an unavailable choice is selected', async () => {
    const res = await request(app)
      .post('/configurators/prod-123/validate')
      .send({ selections: { ram: '32gb' } }); // available: false

    expect(res.status).toBe(422);
    expect(res.body.valid).toBe(false);
    expect(res.body.errors.some((e) => e.includes('unavailable'))).toBe(true);
  });

  it('returns 400 when selections field is missing', async () => {
    const res = await request(app)
      .post('/configurators/prod-123/validate')
      .send({});

    expect(res.status).toBe(400);
  });

  it('returns 404 when product configurator does not exist', async () => {
    Configurator.findOne.mockReturnValue({ lean: jest.fn().mockResolvedValue(null) });

    const res = await request(app)
      .post('/configurators/no-product/validate')
      .send({ selections: { ram: '8gb' } });

    expect(res.status).toBe(404);
  });
});

// ─── PUT /configurators/:productId ───────────────────────────────────────────

describe('PUT /configurators/:productId', () => {
  it('returns 200 with updated configurator', async () => {
    const updatedDoc = { ...Configurator.__mockDoc, rules: [] };
    Configurator.findOneAndUpdate.mockReturnValue({ lean: jest.fn().mockResolvedValue(updatedDoc) });

    const res = await request(app)
      .put('/configurators/prod-123')
      .send({ options: Configurator.__mockDoc.options, rules: [] });

    expect(res.status).toBe(200);
    expect(res.body.rules).toHaveLength(0);
  });

  it('returns 404 when configurator to update does not exist', async () => {
    Configurator.findOneAndUpdate.mockReturnValue({ lean: jest.fn().mockResolvedValue(null) });

    const res = await request(app)
      .put('/configurators/missing')
      .send({ options: [], rules: [] });

    expect(res.status).toBe(404);
  });
});

// ─── DELETE /configurators/:productId ────────────────────────────────────────

describe('DELETE /configurators/:productId', () => {
  it('returns 204 on successful delete', async () => {
    Configurator.findOneAndDelete.mockResolvedValue(Configurator.__mockDoc);

    const res = await request(app).delete('/configurators/prod-123');

    expect(res.status).toBe(204);
  });

  it('returns 404 when configurator does not exist', async () => {
    Configurator.findOneAndDelete.mockResolvedValue(null);

    const res = await request(app).delete('/configurators/missing');

    expect(res.status).toBe(404);
  });
});
