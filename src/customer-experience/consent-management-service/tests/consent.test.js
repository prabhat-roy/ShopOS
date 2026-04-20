'use strict';

const request = require('supertest');
const createApp = require('../src/app');
const ConsentService = require('../src/services/ConsentService');
const { ConsentType } = require('../src/models/consent');

jest.mock('../src/services/ConsentService');

describe('Consent Management Service', () => {
  let app;
  let mockConsentService;

  beforeEach(() => {
    jest.clearAllMocks();
    mockConsentService = new ConsentService();

    const ConsentController = require('../src/controllers/ConsentController');
    const controller = new ConsentController(mockConsentService);

    app = createApp({ consentController: controller });
  });

  // Test 1: Health check
  test('GET /healthz returns status ok', async () => {
    const res = await request(app).get('/healthz');
    expect(res.status).toBe(200);
    expect(res.body.status).toBe('ok');
    expect(res.body.service).toBe('consent-management-service');
  });

  // Test 2: Grant consent successfully
  test('POST /consent grants consent', async () => {
    const mockRecord = {
      customer_id: 'cust-001',
      type: ConsentType.MARKETING_EMAIL,
      granted: true,
      source: 'web-signup',
      updated_at: new Date().toISOString(),
    };
    mockConsentService.grantConsent = jest.fn().mockResolvedValue(mockRecord);

    const res = await request(app)
      .post('/consent')
      .send({ customerId: 'cust-001', type: ConsentType.MARKETING_EMAIL, source: 'web-signup' });

    expect(res.status).toBe(200);
    expect(res.body.granted).toBe(true);
    expect(res.body.type).toBe(ConsentType.MARKETING_EMAIL);
    expect(mockConsentService.grantConsent).toHaveBeenCalledTimes(1);
  });

  // Test 3: Grant consent missing required fields
  test('POST /consent returns 400 when required fields missing', async () => {
    const res = await request(app)
      .post('/consent')
      .send({ customerId: 'cust-001' });

    expect(res.status).toBe(400);
    expect(res.body.error).toBe('Bad Request');
  });

  // Test 4: Revoke consent by type
  test('DELETE /consent revokes consent by type', async () => {
    const mockRecord = {
      customer_id: 'cust-001',
      type: ConsentType.ANALYTICS,
      granted: false,
      source: 'user-settings',
      updated_at: new Date().toISOString(),
    };
    mockConsentService.revokeConsent = jest.fn().mockResolvedValue(mockRecord);

    const res = await request(app)
      .delete('/consent')
      .send({ customerId: 'cust-001', type: ConsentType.ANALYTICS, source: 'user-settings' });

    expect(res.status).toBe(200);
    expect(res.body.granted).toBe(false);
    expect(mockConsentService.revokeConsent).toHaveBeenCalledTimes(1);
  });

  // Test 5: Cannot revoke ESSENTIAL consent
  test('DELETE /consent returns 422 when revoking ESSENTIAL consent', async () => {
    mockConsentService.revokeConsent = jest.fn().mockRejectedValue(
      Object.assign(new Error('ESSENTIAL consent cannot be revoked'), { statusCode: 422 })
    );

    const res = await request(app)
      .delete('/consent')
      .send({ customerId: 'cust-001', type: ConsentType.ESSENTIAL, source: 'user-settings' });

    expect(res.status).toBe(422);
    expect(res.body.message).toMatch(/ESSENTIAL/);
  });

  // Test 6: Get all consent statuses for customer
  test('GET /consent/:customerId returns all consent statuses', async () => {
    const statusMap = {
      MARKETING_EMAIL: true,
      MARKETING_SMS: false,
      ANALYTICS: true,
      PERSONALIZATION: false,
      THIRD_PARTY_SHARING: false,
      ESSENTIAL: true,
    };
    mockConsentService.getConsentStatus = jest.fn().mockResolvedValue(statusMap);

    const res = await request(app).get('/consent/cust-001');

    expect(res.status).toBe(200);
    expect(res.body.customerId).toBe('cust-001');
    expect(res.body.consents).toEqual(statusMap);
    expect(mockConsentService.getConsentStatus).toHaveBeenCalledWith('cust-001');
  });

  // Test 7: Check single consent type - granted
  test('GET /consent/:customerId/:type returns consent status true', async () => {
    mockConsentService.checkConsent = jest.fn().mockResolvedValue(true);

    const res = await request(app).get('/consent/cust-001/MARKETING_EMAIL');

    expect(res.status).toBe(200);
    expect(res.body.granted).toBe(true);
    expect(res.body.customerId).toBe('cust-001');
    expect(res.body.type).toBe('MARKETING_EMAIL');
  });

  // Test 8: Check single consent type - not granted
  test('GET /consent/:customerId/:type returns consent status false', async () => {
    mockConsentService.checkConsent = jest.fn().mockResolvedValue(false);

    const res = await request(app).get('/consent/cust-001/ANALYTICS');

    expect(res.status).toBe(200);
    expect(res.body.granted).toBe(false);
  });

  // Test 9: ESSENTIAL consent always returns true
  test('GET /consent/:customerId/ESSENTIAL always returns granted true', async () => {
    mockConsentService.checkConsent = jest.fn().mockResolvedValue(true);

    const res = await request(app).get('/consent/cust-001/ESSENTIAL');

    expect(res.status).toBe(200);
    expect(res.body.granted).toBe(true);
  });

  // Test 10: Revoke all consents for customer
  test('DELETE /consent/:customerId revokes all consents', async () => {
    const revokedConsents = [
      { type: 'MARKETING_EMAIL' },
      { type: 'ANALYTICS' },
    ];
    mockConsentService.revokeAllConsents = jest.fn().mockResolvedValue(revokedConsents);

    const res = await request(app)
      .delete('/consent/cust-001')
      .send({ source: 'gdpr-request' });

    expect(res.status).toBe(200);
    expect(res.body.revokedCount).toBe(2);
    expect(mockConsentService.revokeAllConsents).toHaveBeenCalledWith('cust-001', 'gdpr-request', expect.anything());
  });

  // Test 11: Revoke all - missing source
  test('DELETE /consent/:customerId returns 400 when source missing', async () => {
    const res = await request(app)
      .delete('/consent/cust-001')
      .send({});

    expect(res.status).toBe(400);
    expect(res.body.error).toBe('Bad Request');
  });

  // Test 12: Get consent history for customer and type
  test('GET /consent/:customerId/:type/history returns history', async () => {
    const history = [
      { id: 'h1', customer_id: 'cust-001', type: 'ANALYTICS', action: 'grant', created_at: new Date().toISOString() },
      { id: 'h2', customer_id: 'cust-001', type: 'ANALYTICS', action: 'revoke', created_at: new Date().toISOString() },
    ];
    mockConsentService.getConsentHistory = jest.fn().mockResolvedValue(history);

    const res = await request(app).get('/consent/cust-001/ANALYTICS/history');

    expect(res.status).toBe(200);
    expect(res.body.history).toHaveLength(2);
    expect(res.body.history[0].action).toBe('grant');
    expect(mockConsentService.getConsentHistory).toHaveBeenCalledWith('cust-001', 'ANALYTICS', undefined);
  });

  // Test 13: Service error propagates as 500
  test('GET /consent/:customerId returns 500 on unexpected error', async () => {
    mockConsentService.getConsentStatus = jest.fn().mockRejectedValue(new Error('DB connection lost'));

    const res = await request(app).get('/consent/cust-001');

    expect(res.status).toBe(500);
    expect(res.body.error).toBe('Internal Server Error');
  });

  // Test 14: Invalid consent type returns 400 from service
  test('POST /consent returns 400 for invalid consent type', async () => {
    mockConsentService.grantConsent = jest.fn().mockRejectedValue(
      Object.assign(new Error('Invalid consent type: UNKNOWN_TYPE'), { statusCode: 400 })
    );

    const res = await request(app)
      .post('/consent')
      .send({ customerId: 'cust-001', type: 'UNKNOWN_TYPE', source: 'web' });

    expect(res.status).toBe(400);
    expect(res.body.message).toMatch(/Invalid consent type/);
  });
});
