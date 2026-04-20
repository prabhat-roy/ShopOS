'use strict';

const { v4: uuidv4 } = require('uuid');
const { publish } = require('../kafka/producer');
const config = require('../config');

// E.164 phone format: + followed by 7–15 digits
const E164_REGEX = /^\+[1-9]\d{6,14}$/;

/**
 * Validates an SMS notification event payload.
 * @param {object} payload
 * @returns {{ valid: boolean, errors: string[] }}
 */
function validate(payload) {
  const errors = [];

  if (!payload.to || !E164_REGEX.test(payload.to)) {
    errors.push('Field "to" must be a valid E.164 phone number (e.g. +14155552671)');
  }

  if (!payload.message || typeof payload.message !== 'string' || payload.message.trim() === '') {
    errors.push('Field "message" is required and must be a non-empty string');
  }

  if (payload.message && payload.message.length > 1600) {
    errors.push('Field "message" must not exceed 1600 characters');
  }

  return { valid: errors.length === 0, errors };
}

/**
 * Handles an incoming notification.sms.requested event.
 * Validates required fields and publishes to the sms.send topic.
 *
 * @param {object} event - Raw Kafka message payload
 * @returns {Promise<object>} - The published message envelope
 */
async function handle(event) {
  const { valid, errors } = validate(event);

  if (!valid) {
    const err = new Error(`Invalid SMS notification payload: ${errors.join('; ')}`);
    err.code = 'VALIDATION_ERROR';
    err.details = errors;
    throw err;
  }

  const envelope = {
    messageId: event.messageId || uuidv4(),
    channel: 'sms',
    to: event.to,
    message: event.message.trim(),
    from: event.from || null,
    templateId: event.templateId || null,
    templateVariables: event.templateVariables || {},
    metadata: event.metadata || {},
    requestedAt: event.requestedAt || new Date().toISOString(),
    routedAt: new Date().toISOString(),
    routedBy: config.service.name,
  };

  await publish(config.kafka.outputTopics.sms, envelope, envelope.messageId);

  return envelope;
}

module.exports = { handle, validate };
