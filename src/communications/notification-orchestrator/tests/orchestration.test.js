'use strict';

// Mock KafkaJS before requiring any module that imports it
jest.mock('kafkajs', () => {
  const mockSend = jest.fn().mockResolvedValue([{ topicName: 'mocked', partition: 0 }]);
  const mockConnect = jest.fn().mockResolvedValue(undefined);
  const mockDisconnect = jest.fn().mockResolvedValue(undefined);
  const mockSubscribe = jest.fn().mockResolvedValue(undefined);
  const mockRun = jest.fn().mockResolvedValue(undefined);

  return {
    Kafka: jest.fn().mockImplementation(() => ({
      producer: jest.fn().mockReturnValue({
        connect: mockConnect,
        disconnect: mockDisconnect,
        send: mockSend,
      }),
      consumer: jest.fn().mockReturnValue({
        connect: mockConnect,
        disconnect: mockDisconnect,
        subscribe: mockSubscribe,
        run: mockRun,
      }),
    })),
    __mockSend: mockSend,
    __mockConnect: mockConnect,
  };
});

const kafkajs = require('kafkajs');
const orchestrationService = require('../src/services/OrchestrationService');
const emailChannel = require('../src/channels/emailChannel');
const smsChannel = require('../src/channels/smsChannel');
const pushChannel = require('../src/channels/pushChannel');

beforeEach(() => {
  orchestrationService.resetStats();
  kafkajs.__mockSend.mockClear();
});

// ─── Email routing ────────────────────────────────────────────────────────────

describe('OrchestrationService.route — email', () => {
  test('routes a valid email event to the email channel', async () => {
    const event = {
      to: 'user@example.com',
      subject: 'Welcome',
      body: 'Hello World',
    };

    const result = await orchestrationService.route('notification.email.requested', event);

    expect(result.channel).toBe('email');
    expect(result.to).toBe('user@example.com');
    expect(result.subject).toBe('Welcome');
    expect(kafkajs.__mockSend).toHaveBeenCalledTimes(1);

    const callArg = kafkajs.__mockSend.mock.calls[0][0];
    expect(callArg.topic).toBe('email.send');
  });

  test('routes a valid email event with templateId (no body required)', async () => {
    const event = {
      to: 'user@example.com',
      templateId: 'tmpl-welcome-001',
      templateVariables: { name: 'Alice' },
    };

    const result = await orchestrationService.route('notification.email.requested', event);

    expect(result.channel).toBe('email');
    expect(result.templateId).toBe('tmpl-welcome-001');
  });

  test('throws VALIDATION_ERROR for email event missing "to" field', async () => {
    const event = { subject: 'Test', body: 'Hello' };

    await expect(
      orchestrationService.route('notification.email.requested', event),
    ).rejects.toMatchObject({ code: 'VALIDATION_ERROR' });
  });

  test('throws VALIDATION_ERROR for email event missing both body and templateId', async () => {
    const event = { to: 'user@example.com', subject: 'Test' };

    await expect(
      orchestrationService.route('notification.email.requested', event),
    ).rejects.toMatchObject({ code: 'VALIDATION_ERROR' });
  });
});

// ─── SMS routing ─────────────────────────────────────────────────────────────

describe('OrchestrationService.route — SMS', () => {
  test('routes a valid SMS event to the SMS channel', async () => {
    const event = {
      to: '+14155552671',
      message: 'Your OTP is 123456',
    };

    const result = await orchestrationService.route('notification.sms.requested', event);

    expect(result.channel).toBe('sms');
    expect(result.to).toBe('+14155552671');

    const callArg = kafkajs.__mockSend.mock.calls[0][0];
    expect(callArg.topic).toBe('sms.send');
  });

  test('throws VALIDATION_ERROR for SMS event with invalid phone number', async () => {
    const event = { to: '0800-INVALID', message: 'Hello' };

    await expect(
      orchestrationService.route('notification.sms.requested', event),
    ).rejects.toMatchObject({ code: 'VALIDATION_ERROR' });
  });
});

// ─── Push routing ─────────────────────────────────────────────────────────────

describe('OrchestrationService.route — push', () => {
  test('routes a valid push event using deviceToken', async () => {
    const event = {
      deviceToken: 'abc123devicetoken',
      title: 'New Order',
      body: 'Your order has shipped',
    };

    const result = await orchestrationService.route('notification.push.requested', event);

    expect(result.channel).toBe('push');
    expect(result.deviceToken).toBe('abc123devicetoken');

    const callArg = kafkajs.__mockSend.mock.calls[0][0];
    expect(callArg.topic).toBe('push.send');
  });

  test('routes a valid push event using userId instead of deviceToken', async () => {
    const event = {
      userId: 'user-uuid-9999',
      title: 'Sale Alert',
      body: '50% off today only!',
    };

    const result = await orchestrationService.route('notification.push.requested', event);

    expect(result.channel).toBe('push');
    expect(result.userId).toBe('user-uuid-9999');
  });

  test('throws VALIDATION_ERROR for push event missing both deviceToken and userId', async () => {
    const event = { title: 'Hello', body: 'World' };

    await expect(
      orchestrationService.route('notification.push.requested', event),
    ).rejects.toMatchObject({ code: 'VALIDATION_ERROR' });
  });
});

// ─── Unknown topic ────────────────────────────────────────────────────────────

describe('OrchestrationService.route — unknown topic', () => {
  test('throws UNKNOWN_TOPIC error for unrecognised topic', async () => {
    const event = { foo: 'bar' };

    await expect(
      orchestrationService.route('notification.carrier.pigeon', event),
    ).rejects.toMatchObject({ code: 'UNKNOWN_TOPIC' });
  });
});

// ─── Stats ────────────────────────────────────────────────────────────────────

describe('OrchestrationService stats', () => {
  test('increments processed and succeeded counters on success', async () => {
    const event = { to: 'a@b.com', subject: 'Hi', body: 'Hey' };
    await orchestrationService.route('notification.email.requested', event);

    const stats = orchestrationService.getStats();
    expect(stats.processed).toBe(1);
    expect(stats.succeeded).toBe(1);
    expect(stats.failed).toBe(0);
    expect(stats.byChannel.email).toBe(1);
  });

  test('increments failed counter on validation error', async () => {
    await expect(
      orchestrationService.route('notification.sms.requested', { to: 'bad', message: 'hi' }),
    ).rejects.toBeDefined();

    const stats = orchestrationService.getStats();
    expect(stats.failed).toBe(1);
    expect(stats.succeeded).toBe(0);
  });
});
