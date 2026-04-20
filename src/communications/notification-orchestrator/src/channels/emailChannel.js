'use strict';

const { v4: uuidv4 } = require('uuid');
const { publish } = require('../kafka/producer');
const config = require('../config');

/**
 * Validates an email notification event payload.
 * @param {object} payload
 * @returns {{ valid: boolean, errors: string[] }}
 */
function validate(payload) {
  const errors = [];

  if (!payload.to || typeof payload.to !== 'string' || !payload.to.includes('@')) {
    errors.push('Field "to" must be a valid email address');
  }

  if (!payload.subject && !payload.templateId) {
    errors.push('Field "subject" is required when "templateId" is not provided');
  }

  if (!payload.body && !payload.templateId) {
    errors.push('Field "body" or "templateId" is required');
  }

  return { valid: errors.length === 0, errors };
}

/**
 * Handles an incoming notification.email.requested event.
 * Validates required fields and publishes to the email.send topic.
 *
 * @param {object} event - Raw Kafka message payload
 * @returns {Promise<object>} - The published message envelope
 */
async function handle(event) {
  const { valid, errors } = validate(event);

  if (!valid) {
    const err = new Error(`Invalid email notification payload: ${errors.join('; ')}`);
    err.code = 'VALIDATION_ERROR';
    err.details = errors;
    throw err;
  }

  const envelope = {
    messageId: event.messageId || uuidv4(),
    channel: 'email',
    to: event.to,
    subject: event.subject || null,
    body: event.body || null,
    templateId: event.templateId || null,
    templateVariables: event.templateVariables || {},
    replyTo: event.replyTo || null,
    cc: event.cc || [],
    bcc: event.bcc || [],
    attachments: event.attachments || [],
    metadata: event.metadata || {},
    requestedAt: event.requestedAt || new Date().toISOString(),
    routedAt: new Date().toISOString(),
    routedBy: config.service.name,
  };

  await publish(config.kafka.outputTopics.email, envelope, envelope.messageId);

  return envelope;
}

module.exports = { handle, validate };
