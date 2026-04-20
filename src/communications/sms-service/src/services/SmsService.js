'use strict';

const { v4: uuidv4 } = require('uuid');
const config = require('../config');
const { defaultStore } = require('../store/SmsStore');

// E.164 format: + followed by 7–15 digits
const E164_REGEX = /^\+[1-9]\d{6,14}$/;

// Aggregate counters
let counters = { sent: 0, delivered: 0, failed: 0 };

/**
 * Validates a phone number against E.164 format.
 * @param {string} phone
 * @returns {boolean}
 */
function isValidPhone(phone) {
  return typeof phone === 'string' && E164_REGEX.test(phone);
}

/**
 * Simulates sending an SMS message.
 * - Validates the recipient phone number (E.164 format)
 * - Simulates delivery with a configurable success rate
 * - Stores the result in the in-memory log
 *
 * @param {object} params
 * @param {string} params.to - Recipient phone number (E.164)
 * @param {string} params.message - Message text
 * @param {string} [params.messageId] - Optional idempotency key
 * @returns {Promise<object>} SmsRecord
 */
async function sendSms({ to, message, messageId }) {
  if (!isValidPhone(to)) {
    const err = new Error(
      `Invalid phone number "${to}". Must be in E.164 format (e.g. +14155552671)`,
    );
    err.code = 'INVALID_PHONE';
    throw err;
  }

  if (!message || typeof message !== 'string' || message.trim() === '') {
    const err = new Error('Field "message" is required and must be a non-empty string');
    err.code = 'INVALID_MESSAGE';
    throw err;
  }

  const id = messageId || uuidv4();
  const sentAt = new Date().toISOString();

  // Simulate delivery — 90% success rate by default
  const delivered = Math.random() < config.sms.successRate;
  const status = delivered ? 'delivered' : 'failed';

  const record = {
    messageId: id,
    to,
    message: message.trim(),
    status,
    sentAt,
    updatedAt: sentAt,
  };

  defaultStore.set(id, record);
  counters.sent++;

  if (delivered) {
    counters.delivered++;
    console.info(`[SmsService] SMS delivered | messageId=${id} to=${to}`);
  } else {
    counters.failed++;
    console.warn(`[SmsService] SMS delivery failed | messageId=${id} to=${to}`);
  }

  return record;
}

/**
 * Retrieves a previously sent SMS record by messageId.
 * @param {string} messageId
 * @returns {object|null}
 */
function getSmsLog(messageId) {
  return defaultStore.get(messageId) || null;
}

/**
 * Returns aggregate delivery statistics.
 * @returns {{ sent: number, delivered: number, failed: number }}
 */
function getStats() {
  return { ...counters };
}

/**
 * Resets counters and store — for testing only.
 */
function _reset() {
  counters = { sent: 0, delivered: 0, failed: 0 };
  defaultStore.clear();
}

module.exports = { sendSms, getSmsLog, getStats, isValidPhone, _reset };
