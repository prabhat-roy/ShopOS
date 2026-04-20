'use strict';

const { v4: uuidv4 } = require('uuid');
const { publish } = require('../kafka/producer');
const config = require('../config');

/**
 * Validates a push notification event payload.
 * @param {object} payload
 * @returns {{ valid: boolean, errors: string[] }}
 */
function validate(payload) {
  const errors = [];

  const hasTarget = payload.deviceToken || payload.userId;
  if (!hasTarget) {
    errors.push('Field "deviceToken" or "userId" is required');
  }

  if (!payload.title || typeof payload.title !== 'string' || payload.title.trim() === '') {
    errors.push('Field "title" is required and must be a non-empty string');
  }

  if (!payload.body || typeof payload.body !== 'string' || payload.body.trim() === '') {
    errors.push('Field "body" is required and must be a non-empty string');
  }

  return { valid: errors.length === 0, errors };
}

/**
 * Handles an incoming notification.push.requested event.
 * Validates required fields and publishes to the push.send topic.
 *
 * @param {object} event - Raw Kafka message payload
 * @returns {Promise<object>} - The published message envelope
 */
async function handle(event) {
  const { valid, errors } = validate(event);

  if (!valid) {
    const err = new Error(`Invalid push notification payload: ${errors.join('; ')}`);
    err.code = 'VALIDATION_ERROR';
    err.details = errors;
    throw err;
  }

  const envelope = {
    messageId: event.messageId || uuidv4(),
    channel: 'push',
    deviceToken: event.deviceToken || null,
    userId: event.userId || null,
    title: event.title.trim(),
    body: event.body.trim(),
    imageUrl: event.imageUrl || null,
    data: event.data || {},
    badge: event.badge || null,
    sound: event.sound || 'default',
    priority: event.priority || 'high',
    ttl: event.ttl || 3600,
    templateId: event.templateId || null,
    metadata: event.metadata || {},
    requestedAt: event.requestedAt || new Date().toISOString(),
    routedAt: new Date().toISOString(),
    routedBy: config.service.name,
  };

  await publish(config.kafka.outputTopics.push, envelope, envelope.messageId);

  return envelope;
}

module.exports = { handle, validate };
