'use strict';

// Mock KafkaJS before requiring any consuming module
jest.mock('kafkajs', () => {
  const mockConnect = jest.fn().mockResolvedValue(undefined);
  const mockDisconnect = jest.fn().mockResolvedValue(undefined);
  const mockSubscribe = jest.fn().mockResolvedValue(undefined);
  const mockRun = jest.fn().mockResolvedValue(undefined);

  let _eachMessageHandler = null;

  const mockRunImpl = jest.fn().mockImplementation(async ({ eachMessage }) => {
    _eachMessageHandler = eachMessage;
  });

  return {
    Kafka: jest.fn().mockImplementation(() => ({
      consumer: jest.fn().mockReturnValue({
        connect: mockConnect,
        disconnect: mockDisconnect,
        subscribe: mockSubscribe,
        run: mockRunImpl,
      }),
    })),
    __getEachMessageHandler: () => _eachMessageHandler,
    __mockConnect: mockConnect,
  };
});

const kafkajs = require('kafkajs');
const smsService = require('../src/services/SmsService');
const consumer = require('../src/kafka/consumer');
const { SmsStore } = require('../src/store/SmsStore');
const request = require('supertest');
const { createApp } = require('../src/app');

const app = createApp();

beforeEach(() => {
  smsService._reset();
  jest.clearAllMocks();
});

// ─── 1. sendSms success ───────────────────────────────────────────────────────

test('sendSms returns a record with delivered or failed status for a valid phone', async () => {
  const result = await smsService.sendSms({ to: '+14155552671', message: 'Hello!' });

  expect(result.messageId).toBeDefined();
  expect(result.to).toBe('+14155552671');
  expect(result.message).toBe('Hello!');
  expect(['delivered', 'failed']).toContain(result.status);
  expect(result.sentAt).toBeDefined();
});

// ─── 2. sendSms invalid phone ─────────────────────────────────────────────────

test('sendSms throws INVALID_PHONE for a non-E.164 phone number', async () => {
  await expect(
    smsService.sendSms({ to: '555-1234', message: 'Test' }),
  ).rejects.toMatchObject({ code: 'INVALID_PHONE' });
});

test('sendSms throws INVALID_PHONE when "to" is missing', async () => {
  await expect(
    smsService.sendSms({ to: undefined, message: 'Test' }),
  ).rejects.toMatchObject({ code: 'INVALID_PHONE' });
});

// ─── 3. sendSms — messageId idempotency ───────────────────────────────────────

test('sendSms uses provided messageId', async () => {
  const result = await smsService.sendSms({
    to: '+447911123456',
    message: 'OTP: 999999',
    messageId: 'fixed-id-001',
  });

  expect(result.messageId).toBe('fixed-id-001');
});

// ─── 4. getStats ──────────────────────────────────────────────────────────────

test('getStats reflects sent/delivered/failed counts after sending', async () => {
  // Force-deliver: use a deterministic override by sending many messages
  // and checking that sent equals delivered + failed
  for (let i = 0; i < 5; i++) {
    await smsService.sendSms({ to: '+14155550001', message: `msg ${i}` });
  }

  const stats = smsService.getStats();
  expect(stats.sent).toBe(5);
  expect(stats.delivered + stats.failed).toBe(5);
});

// ─── 5. getSmsLog found ───────────────────────────────────────────────────────

test('getSmsLog returns the record for a known messageId', async () => {
  const sent = await smsService.sendSms({
    to: '+12025550100',
    message: 'Your code is 42',
    messageId: 'log-test-001',
  });

  const log = smsService.getSmsLog('log-test-001');

  expect(log).toBeDefined();
  expect(log.messageId).toBe('log-test-001');
  expect(log.to).toBe('+12025550100');
  expect(log.status).toBe(sent.status);
});

// ─── 6. getSmsLog not found ───────────────────────────────────────────────────

test('getSmsLog returns null for an unknown messageId', () => {
  const log = smsService.getSmsLog('does-not-exist');
  expect(log).toBeNull();
});

// ─── 7. SmsStore eviction ─────────────────────────────────────────────────────

test('SmsStore evicts the oldest entry when capacity is reached', () => {
  const store = new SmsStore(3);

  store.set('id-1', { data: 1 });
  store.set('id-2', { data: 2 });
  store.set('id-3', { data: 3 });

  // Adding a 4th entry should evict 'id-1'
  store.set('id-4', { data: 4 });

  expect(store.size()).toBe(3);
  expect(store.get('id-1')).toBeUndefined();
  expect(store.get('id-4')).toBeDefined();
});

// ─── 8. HTTP GET /sms/:messageId — found ─────────────────────────────────────

test('GET /sms/:messageId returns 200 and record when found', async () => {
  await smsService.sendSms({ to: '+14085550123', message: 'Hello', messageId: 'http-test-001' });

  const res = await request(app).get('/sms/http-test-001');

  expect(res.status).toBe(200);
  expect(res.body.messageId).toBe('http-test-001');
});

// ─── 9. HTTP GET /sms/:messageId — not found ─────────────────────────────────

test('GET /sms/:messageId returns 404 when record not found', async () => {
  const res = await request(app).get('/sms/ghost-message-id');

  expect(res.status).toBe(404);
  expect(res.body.error).toMatch(/not found/i);
});

// ─── 10. Kafka consumer processes sms.send event ─────────────────────────────

test('consumer processes a sms.send Kafka event and calls sendSms', async () => {
  const sendSmsSpy = jest.spyOn(smsService, 'sendSms');

  // Start the consumer so it registers the eachMessage handler
  await consumer.start();

  const handler = kafkajs.__getEachMessageHandler();
  expect(handler).toBeDefined();

  const payload = {
    messageId: 'kafka-evt-001',
    to: '+16505551234',
    message: 'Kafka-triggered SMS',
  };

  await handler({
    topic: 'sms.send',
    partition: 0,
    message: { value: Buffer.from(JSON.stringify(payload)) },
  });

  expect(sendSmsSpy).toHaveBeenCalledWith(
    expect.objectContaining({ to: '+16505551234', messageId: 'kafka-evt-001' }),
  );
});
